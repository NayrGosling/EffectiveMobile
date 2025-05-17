package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"effect/internal/config"
	"effect/internal/db"
	"effect/internal/handler"
	"effect/internal/middleware"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(cfg.LogLevel)
	log.Infof("starting service on port %d, log level=%s", cfg.Port, cfg.LogLevel)

	dbConn, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	log.Debug("database connection established")

	h := &handler.PersonHandler{DB: dbConn}

	mux := http.NewServeMux()

	logged := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Infof("%s %s", r.Method, r.URL.Path)
			next(w, r)
		}
	}

	mux.HandleFunc("/persons", logged(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.Create(w, r)
		case http.MethodGet:
			h.GetAll(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/persons/", logged(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			h.Update(w, r)
		case http.MethodDelete:
			h.Delete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	handlerWithCORS := middleware.CORS(mux)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Infof("listening on %s", addr)
	if err := http.ListenAndServe(addr, handlerWithCORS); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
