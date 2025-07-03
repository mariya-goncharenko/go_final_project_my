package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// layout задаёт формат даты в виде строки "ГГГГММДД"
const layout = "20060102"

// afterNow возвращает true, если date позже времени now
func afterNow(date, now time.Time) bool {
	return date.After(now)
}

// convertToInts разбирает строку вида "1,2,3" в срез целых чисел
func convertToInts(input string) ([]int, error) {
	if input == "" {
		return nil, errors.New("input string is empty")
	}
	// Делим строку на подстроки по запятой
	parts := strings.Split(input, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		// Удаляем пробелы и конвертируем в число
		num, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, errors.New("failed to parse integer from input")
		}
		result = append(result, num)
	}
	return result, nil
}

// validWeekdays проверяет, что числа соответствуют дням недели (от 1 до 7)
func validWeekdays(days []int) bool {
	for _, d := range days {
		if d < 1 || d > 7 {
			return false
		}
	}
	return true
}

/*
validMonthDays -  проверяет корректность дней месяца:
допустимые значения от 1 до 31, а также отрицательные -1, -2 (обозначающие дни с конца месяца)
*/
func validMonthDays(days []int) bool {
	for _, d := range days {
		if d == 0 || d < -2 || d > 31 {
			return false
		}
	}
	return true
}

/*
validMonths -  проверяет корректность месяцев:
допустимые значения от 1 до 31
*/
func validMonths(months []int) bool {
	for _, m := range months {
		if m < 1 || m > 12 {
			return false
		}
	}
	return true
}

// lastDayOfMonth возвращает последний день месяца для переданной даты t
func lastDayOfMonth(t time.Time) int {
	// Сдвигаемся на один месяц вперёд
	nextMonth := t.AddDate(0, 1, 0)
	// Создаём дату первого числа следующего месяца
	firstOfNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, t.Location())
	// Отнимаем один день и получаем последний день текущего месяца
	lastDay := firstOfNextMonth.AddDate(0, 0, -1)
	return lastDay.Day()
}

// CalculateNextDate вычисляет следующую дату согласно правилу повторения repeatRule
func CalculateNextDate(current time.Time, startDateStr, repeatRule string) (string, error) {
	// Пробуем разобрать дату начала в формате layout
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		return "", errors.New("неверный формат даты начала")
	}

	// Если правило повторения пустое — ошибка
	if repeatRule == "" {
		return "", errors.New("правило повторения пустое")
	}

	// Разбиваем правило повторения на части через пробел
	// Например: "d 5", "w 1,3,5", "m 10,-1 1,2,3"
	parts := strings.Fields(repeatRule)
	if len(parts) == 0 {
		return "", errors.New("правило повторения не задано")
	}

	switch parts[0] {
	case "d":
		// Правило повторения по дням
		// Ожидается формат: d N
		if len(parts) != 2 {
			return "", errors.New("ошибка в формате правила для дней")
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval <= 0 || interval > 400 {
			return "", errors.New("недопустимый интервал для дней")
		}

		// Если стартовая дата ещё не наступила, прибавляем интервал сразу
		if !startDate.Before(current) {
			startDate = startDate.AddDate(0, 0, interval)
			return startDate.Format(layout), nil
		}

		// Иначе двигаем стартовую дату вперёд циклами до тех пор,
		// пока она не окажется в будущем
		for {
			startDate = startDate.AddDate(0, 0, interval)
			if afterNow(startDate, current) {
				break
			}
		}
		return startDate.Format(layout), nil

	case "y":
		// Ежегодное повторение
		for {
			startDate = startDate.AddDate(1, 0, 0)
			if afterNow(startDate, current) {
				break
			}
		}
		return startDate.Format(layout), nil

	case "w":
		// Повторение по дням недели
		// Формат: w 1,3,5 (пн, ср, пт)
		if len(parts) != 2 {
			return "", errors.New("ошибка в формате правила для недели")
		}
		daysOfWeek, err := convertToInts(parts[1])
		if err != nil {
			return "", errors.New("неверный формат для дней недели")
		}
		if !validWeekdays(daysOfWeek) {
			return "", errors.New("дни недели должны быть в диапазоне от 1 до 7")
		}

		// Проверяем день за днём, пока не найдём нужный день недели
		for {
			startDate = startDate.AddDate(0, 0, 1)
			weekday := int(startDate.Weekday())
			if weekday == 0 {
				weekday = 7 // Go считает воскресенье нулевым днём недели, но у нас оно = 7
			}
			for _, d := range daysOfWeek {
				if d == weekday && afterNow(startDate, current) {
					return startDate.Format(layout), nil
				}
			}
		}

	case "m":
		// Повторение по дням месяца
		// Формат:
		//   m 10,20         (10-е и 20-е число каждого месяца)
		//   m -1,-2 1,2,3   (последний и предпоследний день месяцев январь, февраль, март)
		if len(parts) < 2 || len(parts) > 3 {
			return "", errors.New("неверный формат правила для месяцев")
		}

		monthDays, err := convertToInts(parts[1])
		if err != nil {
			return "", errors.New("неверный формат дней")
		}
		if !validMonthDays(monthDays) {
			return "", errors.New("дни месяца должны быть в диапозоне от 1 до 31 или -1, -2")
		}

		var months []int
		if len(parts) == 3 {
			months, err = convertToInts(parts[2])
			if err != nil {
				return "", errors.New("неверный формат месяцев")
			}
			if !validMonths(months) {
				return "", errors.New("месяцы могут быть в диапозоне от 1 до 12")
			}
		} else {
			// Если месяцы не указаны, используем все 12 месяцев
			months = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		}

		// Проходим каждый день, пока не попадём на подходящий день месяца
		for {
			startDate = startDate.AddDate(0, 0, 1)
			currMonth := int(startDate.Month())
			currDay := startDate.Day()

			for _, m := range months {
				if currMonth != m {
					continue
				}
				for _, md := range monthDays {
					// Проверяем обычные числа месяца (положительные)
					if md > 0 && currDay == md {
						if afterNow(startDate, current) {
							return startDate.Format(layout), nil
						}
					} else if md < 0 {
						// Для отрицательных чисел проверяем смещение от конца месяца
						last := lastDayOfMonth(startDate)
						if currDay == last+md+1 {
							if afterNow(startDate, current) {
								return startDate.Format(layout), nil
							}
						}
					}
				}
			}

			// Защита от бесконечного цикла (например, если введены невозможные условия)
			if startDate.After(current.AddDate(100, 0, 0)) {
				return "", errors.New("не удалось найти подходящую дату за разумный период")
			}
		}

	default:
		return "", errors.New("правило повторения не поддерживается")
	}
}

// nextDayHandler — HTTP-обработчик, вычисляющий следующую дату повторения
func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что запрос GET
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Чтение параметров из запроса
	nowParam := r.FormValue("now")
	startParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")

	// Если параметр now пустой — используем текущую дату
	if nowParam == "" {
		nowParam = time.Now().Format(layout)
	}

	// Преобразуем строку now в тип time.Time
	now, err := time.Parse(layout, nowParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка параметра now: %v", err), http.StatusBadRequest)
		return
	}

	// Вызываем вычисление следующей даты
	next, err := CalculateNextDate(now, startParam, repeatParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка вычисления даты: %v", err), http.StatusBadRequest)
		return
	}

	// Отправляем результат клиенту
	w.Write([]byte(next))
}
