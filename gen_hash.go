package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	pass := "Pooled2026!xK9m"
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), 10)
	fmt.Println(string(hash))
}
