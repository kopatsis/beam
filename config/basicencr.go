package config

import (
	"crypto/md5"
	"encoding/base64"
	"os"
	"strconv"
	"time"
)

func getXORKey() byte {
	return md5.Sum([]byte(os.Getenv("CUST_ENCR_KEY")))[0]
}

func EncryptInt(value int) string {
	key := getXORKey()
	encrypted := value ^ int(key)
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(encrypted)))
}

func DecryptInt(encrypted string) (int, error) {
	key := getXORKey()
	decoded, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return 0, err
	}
	encInt, err := strconv.Atoi(string(decoded))
	if err != nil {
		return 0, err
	}
	return encInt ^ int(key), nil
}

func EncodeTime(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 36)
}

func DecodeTime(encoded string) (time.Time, error) {
	seconds, err := strconv.ParseInt(encoded, 36, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0), nil
}
