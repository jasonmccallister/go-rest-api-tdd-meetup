package main

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Message string `json:"message"`
}

func handler() http.HandlerFunc {
	resp := response{
		Message: "ok",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}
}

func main() {
	http.HandleFunc("/", handler())

	http.ListenAndServe(":8000", nil)
}
