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

	// обогащаем
	info, err := service.Enrich(p.Name)
	if err != nil {
		http.Error(w, "enrich error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	p.Age, p.Gender, p.Nationality = info.Age, info.Gender, info.Nationality

	// формируем сообщение
	p.Message = fmt.Sprintf(
		"%s %s%s: age %v, gender %v, nationality %v",
		p.Name, p.Surname,
		func() string {
			if p.Patronymic != nil {
				return " " + *p.Patronymic
			}
			return ""
		}(),
		ptrToString(p.Age, "unknown"),
		ptrToString(p.Gender, "unknown"),
		ptrToString(p.Nationality, "unknown"),
	)

	// сохраняем в БД
	query := `
      INSERT INTO persons (name, surname, patronymic, age, gender, nationality)
      VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at`
	if err := h.DB.QueryRow(query,
		p.Name, p.Surname, p.Patronymic, p.Age, p.Gender, p.Nationality,
	).Scan(&p.ID, &p.CreatedAt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// вспомогалка для форматирования
func ptrToString[T any](p *T, def string) string {
	if p == nil {
		return def
	}
	return fmt.Sprint(*p)
}

func (h *PersonHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// собираем фильтры
	var (
		where []string
		args  []interface{}
		idx   = 1
	)
	if v := q.Get("name"); v != "" {
		where = append(where, fmt.Sprintf("name ILIKE $%d", idx))
		args = append(args, "%"+v+"%")
		idx++
	}
	if v := q.Get("surname"); v != "" {
		where = append(where, fmt.Sprintf("surname ILIKE $%d", idx))
		args = append(args, "%"+v+"%")
		idx++
	}

	// базовый SELECT
	base := `
        SELECT id, name, surname, patronymic, age, gender, nationality, created_at
        FROM persons`
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

	var result []model.Person
	for rows.Next() {
		var p model.Person
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Surname, &p.Patronymic,
			&p.Age, &p.Gender, &p.Nationality, &p.CreatedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// формируем обогащённое сообщение
		fullName := p.Name + " " + p.Surname
		if p.Patronymic != nil {
			fullName += " " + *p.Patronymic
		}
		p.Message = fmt.Sprintf(
			"%s: age %s, gender %s, nationality %s",
			fullName,
			ptrToString(p.Age, "unknown"),
			ptrToString(p.Gender, "unknown"),
			ptrToString(p.Nationality, "unknown"),
		)

		result = append(result, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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
