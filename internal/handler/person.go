package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"effect/internal/model"
	"effect/internal/service"
)

type PersonHandler struct {
	DB *sql.DB
}

func (h *PersonHandler) Create(w http.ResponseWriter, r *http.Request) {
	log.Debug("PersonHandler.Create: decoding request body")
	var p model.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.WithError(err).Warn("PersonHandler.Create: invalid request payload")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("PersonHandler.Create: enriching data for name=%s", p.Name)
	info, err := service.Enrich(p.Name)
	if err != nil {
		log.WithError(err).Error("PersonHandler.Create: enrich error")
		http.Error(w, "enrich error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	p.Age, p.Gender, p.Nationality = info.Age, info.Gender, info.Nationality

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

	query := `
		INSERT INTO persons (name, surname, patronymic, age, gender, nationality)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at`
	log.Debug("PersonHandler.Create: executing DB insert")
	if err := h.DB.QueryRow(query,
		p.Name, p.Surname, p.Patronymic, p.Age, p.Gender, p.Nationality,
	).Scan(&p.ID, &p.CreatedAt); err != nil {
		log.WithError(err).Error("PersonHandler.Create: failed to insert person")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("PersonHandler.Create: created person ID=%d", p.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *PersonHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	log.Debug("PersonHandler.GetAll: parsing query params")
	q := r.URL.Query()
	var where []string
	var args []interface{}
	idx := 1
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

	base := `
		SELECT id, name, surname, patronymic, age, gender, nationality, created_at
		FROM persons`
	if len(where) > 0 {
		base += " WHERE " + strings.Join(where, " AND ")
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	offset, _ := strconv.Atoi(q.Get("offset"))
	base += fmt.Sprintf(" ORDER BY id LIMIT %d OFFSET %d", limit, offset)

	log.Debugf("PersonHandler.GetAll: executing query: %s args=%v", base, args)
	rows, err := h.DB.Query(base, args...)
	if err != nil {
		log.WithError(err).Error("PersonHandler.GetAll: query failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []model.Person
	for rows.Next() {
		var p model.Person
		rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Patronymic,
			&p.Age, &p.Gender, &p.Nationality, &p.CreatedAt)

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

	log.Infof("PersonHandler.GetAll: returning %d persons", len(result))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *PersonHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/persons/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.WithError(err).Warnf("PersonHandler.GetByID: invalid id %s", idStr)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	log.Infof("PersonHandler.GetByID: fetching person id=%d", id)

	if h.DB == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var p model.Person
	err = h.DB.QueryRow(
		`SELECT id, name, surname, patronymic, age, gender, nationality, created_at FROM persons WHERE id=$1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Surname, &p.Patronymic,
		&p.Age, &p.Gender, &p.Nationality, &p.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.WithError(err).Error("PersonHandler.GetByID: query failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PersonHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/persons/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.WithError(err).Warnf("PersonHandler.Update: invalid id %s", idStr)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	log.Infof("PersonHandler.Update: updating person id=%d", id)

	var p model.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.WithError(err).Warn("PersonHandler.Update: invalid request payload")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := h.DB.Exec(
		`UPDATE persons SET name=$1, surname=$2, patronymic=$3 WHERE id=$4`,
		p.Name, p.Surname, p.Patronymic, id,
	)
	if err != nil {
		log.WithError(err).Error("PersonHandler.Update: exec failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PersonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/persons/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.WithError(err).Warnf("PersonHandler.Delete: invalid id %s", idStr)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	log.Infof("PersonHandler.Delete: deleting person id=%d", id)

	res, err := h.DB.Exec("DELETE FROM persons WHERE id=$1", id)
	if err != nil {
		log.WithError(err).Error("PersonHandler.Delete: exec failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ptrToString[T any](p *T, def string) string {
	if p == nil {
		return def
	}
	return fmt.Sprint(*p)
}
