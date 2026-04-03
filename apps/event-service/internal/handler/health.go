package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type HealthHandler struct {
	mysql   *sql.DB
	central *sql.DB
}

func NewHealthHandler(mysql, central *sql.DB) *HealthHandler {
	return &HealthHandler{mysql: mysql, central: central}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	mysqlOK := h.mysql.Ping() == nil
	centralOK := h.central.Ping() == nil
	status := map[string]interface{}{
		"mysql": mysqlOK, "central_db": centralOK, "ready": mysqlOK && centralOK,
	}
	if !mysqlOK || !centralOK {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
