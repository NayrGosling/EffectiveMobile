package main

import (
	"log"
	"net/http"

	"effect/internal/config"
	"effect/internal/db"
	"effect/internal/handler"
)

func main() {
	connStr := config.GetDBConnString()
	database, err := db.NewDB(connStr)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	h := &handler.PersonHandler{DB: database}

	mux := http.NewServeMux()
	mux.HandleFunc("/persons", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.Create(w, r)
		case http.MethodGet:
			h.GetAll(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/persons/", func(w http.ResponseWriter, r *http.Request) {
		// пути вида /persons/{id}
		switch r.Method {
		case http.MethodPut:
			h.Update(w, r)
		case http.MethodDelete:
			h.Delete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
