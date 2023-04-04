package main

import (
	"log"
	"net/http"

	"github.com/farolinar/go-microservice-logger/data"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	// READ json to var
	var requestPayload JSONPayload

	_ = app.ReadJSON(w, r, &requestPayload)

	// insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	log.Printf("Logging in %v\n", event)

	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		app.ErrorJSON(w, err)
	}

	resp := JsonResponse{
		Error:   false,
		Message: "Logged",
	}

	app.WriteJSON(w, http.StatusAccepted, resp)
}
