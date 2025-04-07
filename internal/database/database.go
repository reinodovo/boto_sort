package database

import (
	"errors"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Database interface {
	SaveObject(bucket string, key string, object any) error
	GetObject(bucket string, key string, object any) error
	Close() error
}
