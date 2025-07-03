package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mariya-goncharenko/go_final_project_my/pkg/db"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Создание новой задачи
		addTaskHandler(w, r)

	case http.MethodGet:
		// Получение задачи по ID, переданному в параметрах URL
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJson(w, map[string]string{"error": "Не указан идентификатор"})
			return
		}

		task, err := db.GetTask(id)
		if err != nil {
			writeJson(w, map[string]string{"error": "Задача не найдена"})
			return
		}

		writeJson(w, task)

	case http.MethodPut:
		// Обновление существующей задачи
		var task db.Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			writeJson(w, map[string]string{"error": fmt.Sprintf("ошибка декодирования JSON: %v", err)})
			return
		}

		if task.ID == "" {
			writeJson(w, map[string]string{"error": "Не указан идентификатор задачи"})
			return
		}
		if task.Title == "" {
			writeJson(w, map[string]string{"error": "Не указан заголовок задачи"})
			return
		}

		err = checkDate(&task)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}

		err = db.UpdateTask(&task)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}

		// Возвращаем пустой JSON при успешном обновлении
		writeJson(w, map[string]string{})

	case http.MethodDelete:
		// Удаление задачи по ID, переданному в параметрах URL
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJson(w, map[string]string{"error": "Не указан идентификатор"})
			return
		}

		err := db.DeleteTask(id)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}

		// Возвращаем пустой JSON при успешном удалении
		writeJson(w, map[string]string{})

	default:
		// Метод HTTP не поддерживается
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": "Задача не найдена"})
		return
	}

	if task.Repeat == "" {
		// Одноразовая задача — удаляем
		err := db.DeleteTask(id)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}
		writeJson(w, map[string]string{})
		return
	}

	// Используем дату из задачи, а не текущее время
	layout := "20060102"
	baseDate, err := time.Parse(layout, task.Date)
	if err != nil {
		writeJson(w, map[string]string{"error": "неверный формат даты в задаче"})
		return
	}

	nextDate, err := CalculateNextDate(baseDate, task.Date, task.Repeat)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	err = db.UpdateDate(nextDate, id)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, map[string]string{})
}
