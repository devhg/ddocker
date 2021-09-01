package util

import (
	"math/rand"
	"time"
)

// RandStringBytes creates n length byte slice,
// then convert it into string and return it
func RandStringBytes(n int) string {
	letter := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
