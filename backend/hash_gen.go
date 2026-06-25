//go:build ignore

package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := map[string]string{
		"moderator": "Moderator123!",
		"uye":       "Uye123!",
	}

	for name, pass := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s şifresi (%s):\n%s\n\n", name, pass, hash)
	}
}
