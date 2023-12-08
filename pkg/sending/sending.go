package sending

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	Message string `json:"message"`
}

func SendJSONMessage(w http.ResponseWriter, message string, status int) {
	ser, err := json.Marshal(Message{message})
	if err != nil {
		w.WriteHeader(status)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(ser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func JSONMarshalAndSend(w http.ResponseWriter, obj any) {
	serialized, err := json.Marshal(obj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(serialized)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
