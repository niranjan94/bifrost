package utils

import (
	"crypto/sha1"
	"crypto/sha512"
	"fmt"
)

func SHA1Hash(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func SHA512Hash(input string) string {
	h := sha512.New()
	h.Write([]byte(input))
	return fmt.Sprintf("%x", h.Sum(nil))
}
