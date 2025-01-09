package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"os"

	"github.com/joho/godotenv"
)

type incidentAdded struct {
	Incident struct {
		Backfilled      bool   `json:"backfilled"`
		CreatedAt       string `json:"created_at"`
		ID              string `json:"id"`
		Impact          string `json:"impact"`
		IncidentUpdates []struct {
			Body       string `json:"body"`
			CreatedAt  string `json:"created_at"`
			ID         string `json:"id"`
			IncidentID string `json:"incident_id"`
			Status     string `json:"status"`
			UpdatedAt  string `json:"updated_at"`
		} `json:"incident_updates"`
		Name       string `json:"name"`
		ResolvedAt string `json:"resolved_at"`
		Status     string `json:"status"`
		UpdatedAt  string `json:"updated_at"`
		URL        string `json:"url"`
	} `json:"incident"`
	Meta struct {
		Documentation string `json:"documentation"`
		Unsubscribe   string `json:"unsubscribe"`
	} `json:"meta"`
	Page struct {
		ID                string `json:"id"`
		StatusDescription string `json:"status_description"`
		StatusIndicator   string `json:"status_indicator"`
		URL               string `json:"url"`
	} `json:"page"`
}

func handleWebHook(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var incident incidentAdded

	err := decoder.Decode(&incident)
	if err != nil {
		w.WriteHeader(400)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			return
		}
		return
	}
	fmt.Fprintf(w, incident.Incident.Name);
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/", handleWebHook)

	log.Println("Running")

	addr := os.Getenv("ADDR")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("can't run server: ", err)
	}
}