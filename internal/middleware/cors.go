package middleware

import (
	"net/http"
)

// CORS - это middleware функция, которая устанавливает заголовки для поддержки CORS (Cross-Origin Resource Sharing).
// Она принимает следующий обработчик (next http.Handler) и возвращает новый обработчик (http.Handler).
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем заголовки для поддержки CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Если метод запроса - OPTIONS, то возвращаем статус No Content и завершаем обработку запроса
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Если метод запроса не OPTIONS, то передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	})
}
