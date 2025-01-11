package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	_ "github.com/go-sql-driver/mysql"
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

var b *gotgbot.Bot

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

	token := os.Getenv("API_KEY")
	if token == "" {
		panic("TOKEN environment variable is empty")
	}

	b, err = gotgbot.NewBot(token, nil)
	if err != nil {
		log.Fatalln("failed to create new bot: " + err.Error())
	}

	go func() {
		http.HandleFunc("/", handleWebHook)

		log.Println("Running")

		addr := os.Getenv("ADDR")

		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal("can't run server: ", err)
		}
	}()

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)

	dispatcher.AddHandler(handlers.NewCommand("start", start))

	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", b.User.Username)
	updater.Idle()
}

func start(b *gotgbot.Bot, ctx *ext.Context) error {
	db, err := connectToDB();
	if err != nil {
		log.Fatalln("failed to connect to db: " + err.Error())
	}
	query := "SELECT chatId from users";

	rows, err := db.Query(query)
	if err != nil {
		return logError(err)
	}
	defer rows.Close();

	chatId := ctx.EffectiveChat.Id
	userName := ctx.EffectiveChat.Username

	fmt.Println(chatId);
	fmt.Println(userName);

	return nil
}

func connectToDB() (*sql.DB, error) {
	dsn := "root:@tcp(localhost:3306)/tg_instatus"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, logError(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, logError(err)
	}

	return db, nil
}

func logError(err error) error {
	log.Println(err.Error())
	return err
}