package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/verniyyy/verniy-mq/src"
)

// MQManagerHandler ...
type MQManagerHandler interface {
	Create(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

// newMQManagerHandler ...
func newMQManagerHandler(mqm src.MQManager) MQManagerHandler {
	return mqManagerHandler{
		handlerHelper: handlerHelper{},
		mqManager:     mqm,
	}
}

// mqManagerHandler ...
type mqManagerHandler struct {
	handlerHelper
	mqManager src.MQManager
}

// Create ...
func (h mqManagerHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("uid")
	queueName := r.URL.Query().Get("qn")

	app := src.NewMessageQueueApplication(h.mqManager)
	if err := app.CreateQueue(r.Context(), userID, queueName); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.ResponseJSON(w, http.StatusOK, nil)
}

// List ...
func (h mqManagerHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("uid")

	app := src.NewMessageQueueApplication(h.mqManager)
	out, err := app.ListQueues(r.Context(), userID)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.ResponseJSON(w, http.StatusOK, out)
}

// Delete ...
func (h mqManagerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("uid")
	queueName := chi.URLParam(r, "queueName")

	app := src.NewMessageQueueApplication(h.mqManager)
	if err := app.DeleteQueue(r.Context(), userID, queueName); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.ResponseJSON(w, http.StatusOK, nil)
}

// handlerHelper ...
type handlerHelper struct{}

// ResponseJSON ...
func (h handlerHelper) ResponseJSON(w http.ResponseWriter, code int, payload any) {
	w.WriteHeader(code)
	if payload == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		panic(err)
	}
}
