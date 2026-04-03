package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/eventhub/event-service/internal/db"
	"github.com/eventhub/event-service/internal/models"
)

type EventHandler struct {
	mysql     *sql.DB
	userStore *db.UserStore
}

func NewEventHandler(mysql *sql.DB, us *db.UserStore) *EventHandler {
	return &EventHandler{mysql: mysql, userStore: us}
}

func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	var rows *sql.Rows
	var err error
	if orgID != "" {
		rows, err = h.mysql.QueryContext(r.Context(),
			"SELECT id, org_id, title, description, venue, start_time, end_time, max_attendees, status, created_by, created_at FROM events WHERE org_id = ? ORDER BY start_time DESC LIMIT 50", orgID)
	} else {
		rows, err = h.mysql.QueryContext(r.Context(),
			"SELECT id, org_id, title, description, venue, start_time, end_time, max_attendees, status, created_by, created_at FROM events ORDER BY start_time DESC LIMIT 50")
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.OrgID, &e.Title, &e.Description, &e.Venue, &e.StartTime, &e.EndTime, &e.MaxAttendees, &e.Status, &e.CreatedBy, &e.CreatedAt); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		events = append(events, e)
	}
	jsonResponse(w, http.StatusOK, events)
}

func (h *EventHandler) Get(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventID")
	var e models.Event
	err := h.mysql.QueryRowContext(r.Context(),
		"SELECT id, org_id, title, description, venue, start_time, end_time, max_attendees, status, created_by, created_at FROM events WHERE id = ?", eventID,
	).Scan(&e.ID, &e.OrgID, &e.Title, &e.Description, &e.Venue, &e.StartTime, &e.EndTime, &e.MaxAttendees, &e.Status, &e.CreatedBy, &e.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, e)
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if _, err := h.userStore.GetUser(r.Context(), req.UserID); err != nil {
		http.Error(w, "invalid user: "+err.Error(), http.StatusForbidden)
		return
	}
	if _, err := h.userStore.ValidateMembership(r.Context(), req.UserID, req.OrgID); err != nil {
		http.Error(w, "access denied: "+err.Error(), http.StatusForbidden)
		return
	}

	event := models.Event{
		ID: uuid.New().String(), OrgID: req.OrgID, Title: req.Title,
		Description: req.Description, Venue: req.Venue, StartTime: req.StartTime,
		EndTime: req.EndTime, MaxAttendees: req.MaxAttendees, Status: "draft",
		CreatedBy: req.UserID, CreatedAt: time.Now(),
	}
	_, err := h.mysql.ExecContext(r.Context(),
		"INSERT INTO events (id, org_id, title, description, venue, start_time, end_time, max_attendees, status, created_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		event.ID, event.OrgID, event.Title, event.Description, event.Venue, event.StartTime, event.EndTime, event.MaxAttendees, event.Status, event.CreatedBy)
	if err != nil {
		http.Error(w, "failed to create event: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, event)
}

func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventID")
	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	_, err := h.mysql.ExecContext(r.Context(),
		"UPDATE events SET title=?, description=?, venue=?, start_time=?, end_time=?, max_attendees=? WHERE id=?",
		req.Title, req.Description, req.Venue, req.StartTime, req.EndTime, req.MaxAttendees, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventID")
	h.mysql.ExecContext(r.Context(), "DELETE FROM events WHERE id=?", eventID)
	w.WriteHeader(http.StatusNoContent)
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
