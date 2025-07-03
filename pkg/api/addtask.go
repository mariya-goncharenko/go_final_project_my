package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mariya-goncharenko/go_final_project_my/pkg/db"
)

// addTaskHandler обрабатывает POST-запросы на создание задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	// Чтение JSON из тела запроса
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": fmt.Sprintf("ошибка декодирования JSON: %v", err)})
		return
	}

	// Проверка обязательного поля title
	if task.Title == "" {
		writeJson(w, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	// Проверка даты и правила повторения
	err = checkDate(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	// Добавляем задачу в БД
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": fmt.Sprintf("ошибка записи в БД: %v", err)})
		return
	}

	// Возвращаем id добавленной задачи
	writeJson(w, map[string]string{"id": fmt.Sprintf("%d", id)})
}

// checkDate проверяет дату и правило повторения и изменяет дату задачи, если необходимо
func checkDate(task *db.Task) error {
	now := time.Now()
	layout := "20060102"

	// Если дата не указана — подставляем сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format(layout)
	}

	// Проверка формата даты
	t, err := time.Parse(layout, task.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты")
	}

	// Если указано правило повторения — проверяем и вычисляем следующую дату
	if len(task.Repeat) > 0 {
		next, err := CalculateNextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("неверное правило повторения: %v", err)
		}

		// Меняем дату только если она раньше сегодня (т.е. в прошлом)
		if t.Before(now) {
			task.Date = next
		}
	} else {
		if t.Before(now) {
			task.Date = now.Format(layout)
		}
	}

	return nil
}

// writeJson сериализует данные в JSON и пишет их в ответ
func writeJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_ = json.NewEncoder(w).Encode(data)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
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
		writeJson(w, map[string]string{})
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
