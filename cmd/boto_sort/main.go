package main

import (
	"github.com/reinodovo/boto-sort/internal/bot"
	"github.com/reinodovo/boto-sort/internal/database"
	"github.com/reinodovo/boto-sort/internal/store"
)

func main() {
	db := database.NewMongoDatabase()
	store := store.NewStore(db)
	bot := bot.NewTelegramBot(store)
	bot.Start()
}
