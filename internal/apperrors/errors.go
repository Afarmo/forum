package apperrors

import (
	"errors"
	"log"
)

var (
	ErrNotFound           = errors.New("record not found")
	ErrDuplicateKey       = errors.New("duplicate key violation")
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidID          = errors.New("invalid ID")
	ErrTransactionStart   = errors.New("failed to begin transaction")
	ErrTransactionCommit  = errors.New("failed to commit transaction")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrForbidden          = errors.New("forbidden")
)

func Log(err error) {
	log.Printf("\033[31m  [ERROR]\033[0m    %v", err)
}
