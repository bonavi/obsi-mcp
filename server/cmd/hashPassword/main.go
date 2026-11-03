// Команда hashPassword читает пароль из stdin и печатает хэш argon2id для AUTH_PASSWORD_HASH
package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      = 64 * 1024 // Память в КиБ
	iterations  = 3         // Число проходов
	parallelism = 2         // Число потоков
	saltBytes   = 16        // Размер соли
	hashBytes   = 32        // Размер хэша
)

func main() {
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && password == "" {
		fmt.Fprintln(os.Stderr, "password is required")
		os.Exit(1)
	}
	password = strings.TrimRight(password, "\r\n")
	salt := make([]byte, saltBytes)
	if _, err = rand.Read(salt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, hashBytes)
	fmt.Printf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s\n", argon2.Version, memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
}
