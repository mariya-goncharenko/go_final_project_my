package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB // Глобальное подключение

const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    repeat VARCHAR(128) NOT NULL DEFAULT ''
);
CREATE INDEX idx_date ON scheduler(date);
`

// Init инициализирует БД, создаёт файл и таблицу при необходимости
func Init(dbFile string) error {
	install := false
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		install = true
	}

	var err error
	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("Ошибка подключения к БД: %w", err)
	}

	if install {
		_, err := DB.Exec(schema)
		if err != nil {
			return fmt.Errorf("Ошибка при создании таблицы в БД: %w", err)
		}
		fmt.Println("Таблица успешно создана.")
	}

	return nil
}
