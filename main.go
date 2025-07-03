package main

import (
	"fmt"
	"go1f/pkg/server"
	"log"
	"os"

	"github.com/mariya-goncharenko/go_final_project_my/pkg/db"
)

func main() {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	if err := db.Init(dbFile); err != nil {
		log.Fatal("Ошибка при подключении к БД:", err)
		os.Exit(1)
	}

	if err := server.Run(); err != nil {
		fmt.Println("Ошибка создания сервера:", err)
		os.Exit(1)
	}
}
