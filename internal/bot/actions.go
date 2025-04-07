package bot

import (
	"fmt"
	"math/rand"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/reinodovo/boto-sort/internal/store"
)

func userWaitMessageText(sorting store.Sorting) string {
	usersStr := ""
	for _, user := range sorting.Users {
		usersStr += fmt.Sprintf("- %v\n", user)
	}
	return fmt.Sprintf("Who will be participating in this poll?\n\n%v", usersStr)
}

func userWaitKeyboard() tgbotapi.InlineKeyboardMarkup {
	joinButton := tgbotapi.NewInlineKeyboardButtonData("Me", "join_sorting")
	startButton := tgbotapi.NewInlineKeyboardButtonData("Start", "start_sorting")
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(joinButton), tgbotapi.NewInlineKeyboardRow(startButton))
}

func pollMessageText(sorting store.Sorting, id string) string {
	pendingUsers := make(map[int64]string)
	for userId, user := range sorting.Users {
		pendingUsers[userId] = user
	}
	votes := sorting.CompareResults[id].Votes
	for userId, _ := range votes {
		delete(pendingUsers, userId)
	}
	pendingUsersStr := ""
	for _, user := range pendingUsers {
		pendingUsersStr += fmt.Sprintf("- %v\n", user)
	}
	str := "Which one is better?"
	if len(pendingUsers) > 0 {
		str = fmt.Sprintf("%v\n\nPending Votes:\n%v", str, pendingUsersStr)
	}
	return str
}

func pollKeyboard(sorting store.Sorting, id string) tgbotapi.InlineKeyboardMarkup {
	info := sorting.CompareResults[id]
	a := info.A
	b := info.B
	votes := info.Votes
	aVotes := 0
	bVotes := 0
	for _, vote := range votes {
		if vote == -1 {
			aVotes += 1
		} else {
			bVotes += 1
		}
	}
	aButton := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%v - %v votes", a, aVotes), fmt.Sprintf("poll_%v_a", id))
	bButton := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%v - %v votes", b, bVotes), fmt.Sprintf("poll_%v_b", id))
	revokeButton := tgbotapi.NewInlineKeyboardButtonData("Revoke Vote", fmt.Sprintf("poll_%v_revoke", id))
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(aButton), tgbotapi.NewInlineKeyboardRow(bButton), tgbotapi.NewInlineKeyboardRow(revokeButton))
}

func (bot *TelegramBot) newSorting(chatId int64) error {
	sorting, err := bot.store.GetSorting(chatId)
	if err != nil {
		return err
	}
	if len(sorting.Items) > 0 {
		_, err = bot.SendMessage(chatId, "There is already a sorting happening, use /cancel_sort before starting another", nil)
		return err
	}

	messageId, err := bot.SendMessage(chatId, "Send me the list of items to sort", nil)
	if err != nil {
		return err
	}

	sorting.LastMessageId = messageId

	return bot.store.SaveSorting(chatId, sorting)
}

func (bot *TelegramBot) cancelSorting(chatId int64) error {
	err := bot.store.SaveSorting(chatId, store.NewSorting(chatId))
	if err != nil {
		return err
	}

	_, err = bot.SendMessage(chatId, "Done, you can start a new sorting with /boto_sort", nil)
	return err
}

func (bot *TelegramBot) receiveSortingItems(chatId int64, text string) error {
	sorting, err := bot.store.GetSorting(chatId)
	if err != nil {
		return err
	}

	keyboard := userWaitKeyboard()
	messageId, err := bot.SendMessage(chatId, userWaitMessageText(sorting), &keyboard)
	if err != nil {
		return err
	}

	// TODO: escape items
	sorting.Items = strings.Split(text, "\n")
	rand.Shuffle(len(sorting.Items), func(i, j int) {
		sorting.Items[i], sorting.Items[j] = sorting.Items[j], sorting.Items[i]
	})
	sorting.LastMessageId = messageId

	return bot.store.SaveSorting(chatId, sorting)
}

func (bot *TelegramBot) createPoll(req CompareRequest) error {
	sorting, err := bot.store.GetSorting(req.chatId)
	if err != nil {
		return err
	}

	keyboard := pollKeyboard(sorting, req.id)
	messageId, err := bot.SendMessage(req.chatId, pollMessageText(sorting, req.id), &keyboard)
	if err != nil {
		return err
	}

	cmpResult := sorting.CompareResults[req.id]
	cmpResult.MessageId = messageId
	sorting.CompareResults[req.id] = cmpResult

	return bot.store.SaveSorting(req.chatId, sorting)
}

func (bot *TelegramBot) startSort(sorting store.Sorting) {
	bot.currentSortings[sorting.ChatId] = true
	go sort(sorting.Items, bot.comparator, sorting.ChatId, bot.sortedItems, bot.finishedSortings)
}

func (bot *TelegramBot) assertSorting(sorting store.Sorting) {
	if _, ok := bot.currentSortings[sorting.ChatId]; ok {
		return
	}
	bot.startSort(sorting)
}
