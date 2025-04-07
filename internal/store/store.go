package store

import (
	"errors"
	"fmt"

	"github.com/reinodovo/boto-sort/internal/database"
)

type CompareResult struct {
	Id        string
	A, B      string
	Votes     map[int64]int
	MessageId int
}

type Sorting struct {
	ChatId         int64
	Items          []string
	Users          map[int64]string
	CompareResults map[string]CompareResult
	LastMessageId  int
}

type Store struct {
	db database.Database
}

func NewStore(db database.Database) *Store {
	return &Store{db: db}
}

func NewSorting(chatId int64) Sorting {
	return Sorting{
		ChatId:         chatId,
		Users:          make(map[int64]string),
		CompareResults: make(map[string]CompareResult),
	}
}

func (s *Store) GetSorting(chatId int64) (Sorting, error) {
	sorting := Sorting{}
	err := s.db.GetObject("sortings", fmt.Sprintf("%v", chatId), &sorting)
	if errors.Is(err, database.ErrKeyNotFound) {
		return NewSorting(chatId), nil
	}
	return sorting, err
}

func (s *Store) SaveSorting(chatId int64, sorting Sorting) error {
	return s.db.SaveObject("sortings", fmt.Sprintf("%v", chatId), sorting)
}
