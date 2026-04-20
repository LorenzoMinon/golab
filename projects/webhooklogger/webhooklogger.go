package webhooklogger

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Webhook struct {
	ID        int
	Source    string
	Method    string
	Body      string
	Headers   string
	CreatedAt time.Time
}

type PageData struct {
	Webhooks    []Webhook
	CurrentPage int
	NextPage    int
	PrevPage    int
	HasNext     bool
	HasPrev     bool
}

func initDB() {
	connStr := "host=localhost user=golab password=golab123 dbname=golab sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("could not connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("database unreachable:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS webhooks (
			id         SERIAL PRIMARY KEY,
			source     TEXT,
			method     VARCHAR(10),
			body       TEXT,
			headers    TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatal("could not create webhooks table:", err)
	}

	log.Println("webhooklogger: connected to PostgreSQL")
}

func saveWebhook(source, method, body, headers string) error {
	_, err := db.Exec(
		"INSERT INTO webhooks (source, method, body, headers) VALUES ($1, $2, $3, $4)",
		source, method, body, headers,
	)
	return err
}

func listWebhooks(page int) ([]Webhook, bool, error) {
	const perPage = 10
	offset := (page - 1) * perPage

	rows, err := db.Query(
		"SELECT id, source, method, body, headers, created_at FROM webhooks ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		perPage+1, offset,
	)
	if err != nil {
		return nil, false, fmt.Errorf("could not fetch webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var w Webhook
		if err := rows.Scan(&w.ID, &w.Source, &w.Method, &w.Body, &w.Headers, &w.CreatedAt); err != nil {
			return nil, false, err
		}
		webhooks = append(webhooks, w)
	}

	hasNext := len(webhooks) > perPage
	if hasNext {
		webhooks = webhooks[:perPage]
	}

	return webhooks, hasNext, nil
}

func Handler() http.HandlerFunc {
	initDB()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "could not read body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			headers := ""
			for key, values := range r.Header {
				for _, v := range values {
					headers += fmt.Sprintf("%s: %s\n", key, v)
				}
			}

			source := r.URL.Query().Get("source")
			if source == "" {
				source = "unknown"
			}

			if err := saveWebhook(source, r.Method, string(body), headers); err != nil {
				http.Error(w, "could not save webhook", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "webhook received")
			return
		}

		pageStr := r.URL.Query().Get("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		webhooks, hasNext, err := listWebhooks(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.New("").Funcs(template.FuncMap{
			"formatTime": func(t time.Time) string {
				return t.Format("02 Jan 2006 15:04:05")
			},
		}).ParseFiles("projects/webhooklogger/index.html"))

		data := PageData{
			Webhooks:    webhooks,
			CurrentPage: page,
			NextPage:    page + 1,
			PrevPage:    page - 1,
			HasNext:     hasNext,
			HasPrev:     page > 1,
		}

		if err := tmpl.ExecuteTemplate(w, "webhooklogger", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
