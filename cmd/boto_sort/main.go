package main

import (
	"github.com/reinodovo/boto-sort/internal/bot"
	"github.com/reinodovo/boto-sort/internal/database"
	"github.com/reinodovo/boto-sort/internal/store"
)

func main() {
	db := database.NewMongoDatabase()
	defer db.Close()

	store := store.NewStore(db)
	bot := bot.NewTelegramBot(store)
	bot.Start()
}
