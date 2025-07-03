package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Добавление задачи в таблицу:
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Получаем список задач:
func Tasks(limit int, search string) ([]*Task, error) {
	tasks := make([]*Task, 0, limit)

	layoutSearch := "02.01.2006" // формат даты для поиска из параметра search
	layoutDB := "20060102"       // формат даты в БД

	// Проверяем, соответствует ли search дате в формате 02.01.2006
	searchDate, err := time.Parse(layoutSearch, search)
	isDateSearch := (err == nil)

	var rows *sql.Rows

	if search == "" {
		// Просто берем задачи с сортировкой по дате и ограничением limit
		query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`
		rows, err = DB.Query(query, limit)
	} else if isDateSearch {
		// Ищем задачи по дате, преобразуем дату в формат БД
		searchDateStr := searchDate.Format(layoutDB)
		query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?`
		rows, err = DB.Query(query, searchDateStr, limit)
	} else {
		// Ищем задачи, где title или comment LIKE '%search%'
		likePattern := fmt.Sprintf("%%%s%%", search)
		query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?`
		rows, err = DB.Query(query, likePattern, likePattern, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		t := new(Task)
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Если tasks == nil, возвращаем пустой слайс (чтобы не было null в JSON)
	if tasks == nil {
		tasks = make([]*Task, 0)
	}
	return tasks, nil
}
