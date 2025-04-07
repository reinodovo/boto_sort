package database

import (
	"errors"
	"fmt"
	"reflect"
)

type InMemoryDatabase struct {
	objects map[string]any
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{
		objects: make(map[string]any),
	}
}

func generateKey(bucket, key string) string {
	return fmt.Sprintf("%v_%v", bucket, key)
}

func (db *InMemoryDatabase) SaveObject(bucket string, key string, object any) error {
	key = generateKey(bucket, key)
	db.objects[key] = object
	return nil
}

func (db *InMemoryDatabase) GetObject(bucket string, key string, object any) error {
	key = generateKey(bucket, key)
	value, ok := db.objects[key]
	if !ok {
		return ErrKeyNotFound
	}

	objVal := reflect.ValueOf(object)
	if objVal.Kind() != reflect.Ptr || objVal.IsNil() {
		return errors.New("object must be a non-nil pointer")
	}

	valToSet := reflect.ValueOf(value)
	if !valToSet.Type().AssignableTo(objVal.Elem().Type()) {
		return errors.New("type mismatch")
	}

	objVal.Elem().Set(valToSet)
	return nil
}

func (db *InMemoryDatabase) Close() error {
	return nil
}
