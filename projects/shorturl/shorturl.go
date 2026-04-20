package shorturl

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"

	_ "github.com/lib/pq" // driver
)

type URLEntry struct {
	Code      string
	Original  string
	Clicks    int
	CreatedAt string
}

type PageData struct {
	URLs    []URLEntry
	BaseURL string
	Error   string
	Success string
}

var db *sql.DB

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

	log.Println("shorturl: connected to PostgreSQL")
}

func generateCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}

func createURL(original string) (string, error) {
	code := generateCode()

	_, err := db.Exec(
		"INSERT INTO urls (code, original) VALUES ($1, $2)",
		code, original,
	)
	if err != nil {
		return "", fmt.Errorf("could not save URL: %w", err)
	}

	return code, nil
}

func getURL(code string) (string, error) {
	var original string
	err := db.QueryRow(
		"SELECT original FROM urls WHERE code = $1",
		code,
	).Scan(&original)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("code not found")
	}
	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}

	_, err = db.Exec(
		"UPDATE urls SET clicks = clicks + 1 WHERE code = $1",
		code,
	)
	if err != nil {
		log.Println("could not update clicks:", err)
	}

	return original, nil
}

func listURLs() ([]URLEntry, error) {
	rows, err := db.Query(
		"SELECT code, original, clicks, created_at FROM urls ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("could not fetch URLs: %w", err)
	}
	defer rows.Close()

	var urls []URLEntry
	for rows.Next() {
		var u URLEntry
		if err := rows.Scan(&u.Code, &u.Original, &u.Clicks, &u.CreatedAt); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}

	return urls, nil
}

func Handler() http.HandlerFunc {
	initDB()

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/projects/shorturl")

		if path != "" && path != "/" {
			code := strings.TrimPrefix(path, "/")
			original, err := getURL(code)
			if err != nil {
				http.Error(w, "link not found", http.StatusNotFound)
				return
			}
			http.Redirect(w, r, original, http.StatusMovedPermanently)
			return
		}

		tmpl := template.Must(template.ParseFiles("projects/shorturl/index.html"))

		if r.Method == http.MethodPost {
			original := r.FormValue("url")
			if !strings.HasPrefix(original, "http") {
				original = "https://" + original
			}

			code, err := createURL(original)

			urls, _ := listURLs()
			data := PageData{
				URLs:    urls,
				BaseURL: "http://localhost:8080/projects/shorturl/",
			}

			if err != nil {
				data.Error = "could not shorten URL"
			} else {
				data.Success = code
			}

			tmpl.Execute(w, data)
			return
		}

		urls, err := listURLs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, PageData{
			URLs:    urls,
			BaseURL: "http://localhost:8080/projects/shorturl/",
		})
	}
}
