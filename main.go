package main

import (
	"fmt"
	"go1f/pkg/server"
	"os"
)

func main() {
	if err := server.Run(); err != nil {
		fmt.Println("Ошибка создания сервера:", err)
		os.Exit(1)
	}
}
