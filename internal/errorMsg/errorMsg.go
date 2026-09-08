package errorMsg

import "errors"

var (
	ErrNotFound           = errors.New("Record Not Found")
	ErrDuplicateKey       = errors.New("Duplicate Key Violation")
	ErrInvalidInput       = errors.New("Invalid Input")
	ErrInvalidId          = errors.New("Invalid Id")
	ErrTransactionStart   = errors.New("Failed To Begin Transaction")
	ErrTransactionCommit  = errors.New("Failed To Commit Transaction")
)
