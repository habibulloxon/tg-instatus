package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"errors"

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
var db *sql.DB

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


	chatIds := getChatIds()

	for i := 0; i < len(chatIds); i++ {
		chatId := chatIds[i]
		_, err := b.SendMessage(int64(chatId), incident.Incident.Name, nil)
		if err != nil {
			return
		}
	}
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

	database, err := initDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

	db = database

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
	chatID := ctx.EffectiveChat.Id
	username := ctx.EffectiveChat.Username

	var existingChatID int64
	err := db.QueryRow("SELECT chatId FROM users WHERE chatId = ?", chatID).Scan(&existingChatID)
	
	switch {
	case err == nil:
		if _, err := ctx.EffectiveMessage.Reply(b, "Currently you are watching", nil); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	case errors.Is(err, sql.ErrNoRows):
		if _, err := db.Exec("INSERT INTO users (username, chatId) VALUES(?, ?)", username, chatID); err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
		
		if _, err := ctx.EffectiveMessage.Reply(b, "You are now watching", nil); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		
		log.Printf("New user registered - username: %s, chatID: %d", username, chatID)
	default:
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func connectToDB() (*sql.DB, error) {
	dsn := "root:@tcp(localhost:3306)/tg_instatus"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func getChatIds() []int {
	db, err := connectToDB();
	if err != nil {
		log.Fatalln("failed to connect to db: " + err.Error())
	}
	query := "SELECT chatId from users";

	rows, err := db.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close();

	var ids []int;

	for rows.Next() {
		var chatID int
		if err := rows.Scan(&chatID); err != nil {
			log.Println(err.Error())
			continue
		}
		ids = append(ids, chatID)
		fmt.Println(chatID)
	}

	return ids;
}

func initDB() (*sql.DB, error) {
    db, err := connectToDB()
    if err != nil {
        return nil, err
    }
    
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db, nil
}