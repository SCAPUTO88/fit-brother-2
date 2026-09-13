//go:build ignore

package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
)

func main() {
	password := os.Args[1]
	pepper := os.Args[2]
	hash := os.Args[3]

	peppered := append([]byte(password), []byte(pepper)...)
	err := bcrypt.CompareHashAndPassword([]byte(hash), peppered)
	if err != nil {
		fmt.Println("❌ SENHA ERRADA:", err)
	} else {
		fmt.Println("✅ SENHA CORRETA!")
	}
}
