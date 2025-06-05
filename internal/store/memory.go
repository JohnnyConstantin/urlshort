package store

// Хранит мапу запросов в памяти. Должно быть заменено на БД, но БД не проходит через CI тесты
var (
	URLStore = make(map[string]string) // shortID: originalURL
)
