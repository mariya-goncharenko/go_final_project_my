package api

import (
	"net/http"

	"github.com/mariya-goncharenko/go_final_project_my/pkg/db"
)

// TasksResp — структура ответа с задачами
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler — обработчик GET /api/tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры из URL
	search := r.URL.Query().Get("search")

	// Ограничение количества возвращаемых задач
	limit := 50

	tasks, err := db.Tasks(limit, search)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJson(w, http.StatusOK, TasksResp{Tasks: tasks})
}
