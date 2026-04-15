//go:build ignore

package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := os.Args[1]
	pepper := os.Args[2]

	pepperedPassword := append([]byte(password), []byte(pepper)...)
	hash, err := bcrypt.GenerateFromPassword(pepperedPassword, 12)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(hash))
}
