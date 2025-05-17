package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"effect/internal/model"
	"effect/internal/service"
)

type PersonHandler struct {
	DB *sql.DB
}

func (h *PersonHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p model.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	age, gender, nat, err := service.Enrich(p.Name)
	if err != nil {
		http.Error(w, "enrich error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	p.Age, p.Gender, p.Nationality = age, gender, nat

	query := `
      INSERT INTO persons (name, surname, patronymic, age, gender, nationality)
      VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at`
	err = h.DB.QueryRow(query,
		p.Name, p.Surname, p.Patronymic, p.Age, p.Gender, p.Nationality,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *PersonHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// фильтры: ?name=..&surname=..; пагинация: ?limit=10&offset=20
	q := r.URL.Query()
	var (
		where []string
		args  []interface{}
	)
	i := 1
	if v := q.Get("name"); v != "" {
		where = append(where, fmt.Sprintf("name ILIKE $%d", i))
		args = append(args, "%"+v+"%")
		i++
	}
	if v := q.Get("surname"); v != "" {
		where = append(where, fmt.Sprintf("surname ILIKE $%d", i))
		args = append(args, "%"+v+"%")
		i++
	}
	base := "SELECT id,name,surname,patronymic,age,gender,nationality,created_at FROM persons"
	if len(where) > 0 {
		base += " WHERE " + strings.Join(where, " AND ")
	}
	// пагинация
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	base += fmt.Sprintf(" ORDER BY id LIMIT %d OFFSET %d", limit, offset)

	rows, err := h.DB.Query(base, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var list []model.Person
	for rows.Next() {
		var p model.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Patronymic,
			&p.Age, &p.Gender, &p.Nationality, &p.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, p)
	}
	json.NewEncoder(w).Encode(list)
}

func (h *PersonHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/persons/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var p model.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `
      UPDATE persons SET name=$1,surname=$2,patronymic=$3
      WHERE id=$4`
	if _, err := h.DB.Exec(query, p.Name, p.Surname, p.Patronymic, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PersonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/persons/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := h.DB.Exec("DELETE FROM persons WHERE id=$1", id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
