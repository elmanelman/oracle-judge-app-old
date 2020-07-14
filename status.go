package main

var statusDescription = map[string]string {
	"UNKNOWN": "Статус неизвестен",
	"PENDING_CHECK": "Ожидает проверки",
	"ACCEPTED": "Решение принято",
	"EXECUTION_ERROR": "Ошибка исполнения",
	"RESTRICTION_VIOLATED": "Нарушено ограничение",
	"INCORRECT_COLUMN_COUNT": "Неверное число столбцов",
	"INCORRECT_COLUMN_NAMES": "Неверные названия столбцов",
	"INCORRECT_CONTENT": "Неверное содержимое результата",
	"INCORRECT_ORDER": "Неверный порядок строк результата",
}
