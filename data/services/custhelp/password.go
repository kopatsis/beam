package custhelp

import (
	"beam/config"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func EncryptPassword(password string) (string, error) {
	hashedInput := hashPassword(password)
	hash, err := bcrypt.GenerateFromPassword([]byte(hashedInput), 6)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(storedHash, password string) bool {
	hashedInput := hashPassword(password)
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(hashedInput))
	return err == nil
}

func PasswordMeetsRequirements(password, confirmPassword string, checkConfirm bool) bool {
	if checkConfirm && password != confirmPassword {
		return false
	}

	if len(password) < config.PASSWORD_MIN_CHARS || len(password) > config.PASSWORD_MAX_CHARS {
		return false
	}

	var specials, letters, numbers int
	for _, char := range password {
		switch {
		case strings.ContainsRune(config.SPECIAL_CHAR_LIST, char):
			specials++
		case strings.ContainsRune(config.LETTER_LIST, char):
			letters++
		case strings.ContainsRune(config.NUMBER_LIST, char):
			numbers++
		default:
			return false
		}
	}

	return specials >= config.PASSWORD_MIN_SPECIALS && letters >= config.PASSWORD_MIN_LETTER && numbers >= config.PASSWORD_MIN_NUMBERS

}
