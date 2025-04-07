package bot

import (
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/reinodovo/boto-sort/internal/store"
)

type SortRequests map[int64]int

type SortingResult struct {
	chatId      int64
	sortedItems []string
}

type TelegramBot struct {
	api             *tgbotapi.BotAPI
	comparator      *Comparator
	store           *store.Store
	sortingResults  chan SortingResult
	currentSortings map[int64]bool
}

func NewTelegramBot(store *store.Store) *TelegramBot {
	api, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot := &TelegramBot{
		api:             api,
		store:           store,
		currentSortings: make(map[int64]bool),
		sortingResults:  make(chan SortingResult),
	}
	bot.comparator = NewComparator(bot.createPoll, store)
	go bot.comparator.Start()

	return bot
}

func (bot *TelegramBot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.api.GetUpdatesChan(u)

	go bot.handleSortings()
	bot.handleUpdates(updates)
	return nil
}

func (bot *TelegramBot) handleSortings() {
	for result := range bot.sortingResults {
		_, err := bot.SendMessage(result.chatId, strings.Join(result.sortedItems, "\n"), nil)
		if err != nil {
			log.Println(err)
		}
		err = bot.store.SaveSorting(result.chatId, store.NewSorting(result.chatId))
		if err != nil {
			log.Println(err)
		}
	}
}

func (bot *TelegramBot) handleUpdate(update tgbotapi.Update) error {
	if update.CallbackQuery != nil {
		chatId := update.CallbackQuery.Message.Chat.ID
		sorting, err := bot.store.GetSorting(chatId)
		if err != nil {
			return err
		}

		data := update.CallbackQuery.Data
		if data == "start_sorting" {
			err = bot.EditMessage(chatId, sorting.LastMessageId, "Sorting started", nil)
			if err != nil {
				return err
			}
			bot.startSort(sorting)
		}
		if data == "join_sorting" {
			sorting.Users[update.CallbackQuery.From.ID] = update.CallbackQuery.From.UserName
			err := bot.store.SaveSorting(chatId, sorting)
			if err != nil {
				return err
			}

			keyboard := userWaitKeyboard()
			return bot.EditMessage(chatId, sorting.LastMessageId, userWaitMessageText(sorting), &keyboard)
		}
		if strings.HasPrefix(data, "poll") {
			bot.assertSorting(sorting)
			tokens := strings.Split(data, "_")
			id := tokens[1]
			option := tokens[2]
			err = bot.comparator.receiveVote(chatId, id, update.CallbackQuery.From.ID, option)
			if err != nil {
				return err
			}

			sorting, err := bot.store.GetSorting(chatId)
			if err != nil {
				return err
			}

			text := pollMessageText(sorting, id)
			keyboard := pollKeyboard(sorting, id)
			return bot.EditMessage(chatId, sorting.CompareResults[id].MessageId, text, &keyboard)
		}
	}
	if update.Message != nil {
		if update.Message.ReplyToMessage != nil {
			chatId := update.Message.Chat.ID
			sorting, err := bot.store.GetSorting(chatId)
			if err != nil {
				return err
			}
			if sorting.LastMessageId == update.Message.ReplyToMessage.MessageID {
				return bot.receiveSortingItems(chatId, update.Message.Text)
			}
		}
		if update.Message.IsCommand() {
			if update.Message.Command() == "boto_sort" {
				return bot.newSorting(update.Message.Chat.ID)
			}
			if update.Message.Command() == "cancel_sort" {
				return bot.cancelSorting(update.Message.Chat.ID)
			}
		}
	}
	return nil
}

func (bot *TelegramBot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		err := bot.handleUpdate(update)
		if err != nil {
			log.Println(err)
		}
	}
}

func (bot *TelegramBot) SendMessage(chatId int64, text string, keyboard *tgbotapi.InlineKeyboardMarkup) (int, error) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "html"
	if keyboard != nil {
		msg.ReplyMarkup = *keyboard
	}
	sentMsg, err := bot.api.Send(msg)
	if err != nil {
		return 0, err
	}
	return sentMsg.MessageID, nil
}

func (bot *TelegramBot) EditMessage(chatId int64, messageId int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewEditMessageText(chatId, messageId, text)
	msg.ParseMode = "html"
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	_, err := bot.api.Send(msg)
	return err
}
