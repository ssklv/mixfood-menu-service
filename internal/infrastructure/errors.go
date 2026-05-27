package infrastructure

import (
	"errors"
)

var (
	ErrDishNotFound = errors.New("dish database record not found")
	//ErrNoChanges    = errors.New("no changes to update in database")
)
