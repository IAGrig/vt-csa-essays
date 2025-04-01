package crypto

import "golang.org/x/crypto/bcrypt"

func GenerateHash(input []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(input, 10)
}
