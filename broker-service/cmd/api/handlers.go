package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/farolinar/go-microservice-broker/event"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := JsonResponse{
		Error:   false,
		Message: "Hit the shadow broker",
	}

	err := app.WriteJSON(w, http.StatusOK, payload)
	if err != nil {
		log.Panic(err)
	}
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.ReadJSON(w, r, &requestPayload)
	if err != nil {
		app.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.LogEventViaRabbitMQ(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)

	default:
		app.ErrorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create JSON send to auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer((jsonData)))
	if err != nil {

		app.ErrorJSON(w, err, http.StatusBadGateway)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {

		app.ErrorJSON(w, err, http.StatusInternalServerError)
		return
	}

	defer response.Body.Close()

	// makes sure get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
	} else if response.StatusCode != http.StatusAccepted {
		app.ErrorJSON(w, errors.New("error calling auth service"), http.StatusInternalServerError)
	}

	// read response.Body
	var jsonFromService JsonResponse

	// decode the json
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.ErrorJSON(w, err, http.StatusInternalServerError)
		return
	}

	if jsonFromService.Error {
		app.ErrorJSON(w, errors.New(jsonFromService.Message), http.StatusUnauthorized)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "authenticated!"
	payload.Data = jsonFromService.Data

	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceUrl := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	defer response.Body.Close()

	var jsonFromService JsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.ErrorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var jsonPayload JsonResponse
	jsonPayload.Data = jsonFromService.Data
	jsonPayload.Message = "logged!"
	jsonPayload.Error = false

	app.WriteJSON(w, http.StatusAccepted, jsonPayload)
}

func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, err := json.MarshalIndent(msg, "", "\t")
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	mailServiceUrl := "http://mail-service/send"

	request, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.ErrorJSON(w, errors.New("error calling mail service"))
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: "message sent to " + msg.To,
	}

	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) LogEventViaRabbitMQ(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: fmt.Sprintf("pushed log %s to rabbitmq", l.Name),
	}

	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(l LogPayload) error {
	emitter, err := event.NewEventEmitter(app.rabbitConn)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(l, "", "\t")

	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}
