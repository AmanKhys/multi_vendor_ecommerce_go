package utils

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func DecodeJson(w http.ResponseWriter, r *http.Request, req any) {
	var err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// Check for json.SyntaxError (invalid JSON format)
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			log.Warn("JSON syntax error at position", syntaxError.Offset)
			http.Error(w, "Invalid JSON syntax", http.StatusBadRequest)
			return
		}

		// Check for json.UnmarshalTypeError (wrong type for a field in JSON)
		if unmarshalTypeError, ok := err.(*json.UnmarshalTypeError); ok {
			log.Warn("JSON unmarshal type error:", unmarshalTypeError)
			http.Error(w, "Invalid JSON type", http.StatusBadRequest)
			return
		}

		// If the error is of a different type, handle it as a general decoding error
		log.Warn("error decoding JSON object", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}
