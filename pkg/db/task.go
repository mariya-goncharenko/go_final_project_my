package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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

	layoutSearch := "02.01.2006"
	layoutDB := "20060102"

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

// GetTask - получение задачb по id
func GetTask(id string) (*Task, error) {
	task := new(Task)

	// конвертируем id в int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errors.New("некорректный идентификатор")
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	err = DB.QueryRow(query, idInt).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("задача не найдена")
		}
		return nil, err
	}

	return task, nil
}

// UpdateTask - позволяет обновить задачу по id
func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача с id %d не найдена для обновления", task.ID)
	}

	return nil
}
