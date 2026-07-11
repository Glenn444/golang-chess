package random

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

)
const alphabet = "abcdefghijklmopqrstuvwxyz"
func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	//rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random integer between min and max
func RandomInt(min,max int64) int64{
	return min + rand.Int63n(max - min + 1)
}

func RandomString(n int)string{
	var sb strings.Builder
	k := len(alphabet)

	for range n{
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return  sb.String()
}

func RandomUsername() string{
	return RandomString(6)
}



func RandomEmail()string{
	return  fmt.Sprintf("%s@gmail.com",RandomString(5))
}