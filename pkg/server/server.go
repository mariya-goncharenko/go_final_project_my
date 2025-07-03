package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// Run запускает HTTP-файл сервер на указанном или дефолтном порту
func Run() error {
	port := 7540

	// Проверяем переменную окружения TODO_PORT
	if p := os.Getenv("TODO_PORT"); p != "" {
		val, err := strconv.Atoi(p)
		if err == nil {
			port = val
		} else {
			return fmt.Errorf("invalid TODO_PORT: %v", err)
		}
	}

	// Создаём файловый сервер для директории ./web
	fs := http.FileServer(http.Dir("web"))

	http.Handle("/", fs)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Сервер запущен по адресу: http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}
