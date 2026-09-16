package auth

import (
	"bytes"
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

const (
	iterations = 1
	memory     = 64 * 1024
	threads    = 4
	keyLen     = 32
)

func HashPassword(password string) ([]byte, []byte, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return nil, nil, err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		threads,
		keyLen,
	)

	return hash, salt, nil
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	return salt, nil
}

func VerifyPassword(password string, hash, salt []byte) bool {
	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		threads,
		keyLen,
	)

	return bytes.Equal(newHash, hash)
}
