package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/eventhub/event-service/internal/db"
	"github.com/eventhub/event-service/internal/models"
)

type BookingHandler struct {
	mysql     *sql.DB
	userStore *db.UserStore
}

func NewBookingHandler(mysql *sql.DB, us *db.UserStore) *BookingHandler {
	return &BookingHandler{mysql: mysql, userStore: us}
}

func (h *BookingHandler) List(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventID")
	rows, err := h.mysql.QueryContext(r.Context(),
		"SELECT id, user_id, event_id, ticket_type_id, status, booked_at FROM bookings WHERE event_id = ? ORDER BY booked_at DESC", eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	bookings := []models.Booking{}
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.EventID, &b.TicketTypeID, &b.Status, &b.BookedAt); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		bookings = append(bookings, b)
	}
	jsonResponse(w, http.StatusOK, bookings)
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventID")
	var req models.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.userStore.GetUser(r.Context(), req.UserID)
	if err != nil {
		http.Error(w, "invalid user", http.StatusForbidden)
		return
	}
	booking := models.Booking{
		ID: uuid.New().String(), UserID: user.ID, EventID: eventID,
		TicketTypeID: req.TicketTypeID, Status: "confirmed",
	}
	_, err = h.mysql.ExecContext(r.Context(),
		"INSERT INTO bookings (id, user_id, event_id, ticket_type_id, status) VALUES (?, ?, ?, ?, ?)",
		booking.ID, booking.UserID, booking.EventID, booking.TicketTypeID, booking.Status)
	if err != nil {
		http.Error(w, "booking failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, booking)
}
