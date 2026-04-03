package models

import "time"

type Event struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"org_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Venue        string    `json:"venue"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	MaxAttendees int       `json:"max_attendees"`
	Status       string    `json:"status"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TicketType struct {
	ID       string  `json:"id"`
	EventID  string  `json:"event_id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Sold     int     `json:"sold"`
}

type Booking struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	EventID      string    `json:"event_id"`
	TicketTypeID string    `json:"ticket_type_id"`
	Status       string    `json:"status"`
	BookedAt     time.Time `json:"booked_at"`
}

type CreateEventRequest struct {
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Venue        string    `json:"venue"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	MaxAttendees int       `json:"max_attendees"`
	OrgID        string    `json:"org_id"`
	UserID       string    `json:"user_id"`
}

type CreateBookingRequest struct {
	UserID       string `json:"user_id"`
	TicketTypeID string `json:"ticket_type_id"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Organization struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	PlanTier string `json:"plan_tier"`
}
