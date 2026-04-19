package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/LorenzoMinon/golab/projects/argodash"
	"github.com/LorenzoMinon/golab/projects/linkchecker"
)

type Project struct {
	ID          string
	Name        string
	Description string
	Status      string
	Tags        []string //list
	URL         string
}

type PageData struct {
	Projects []Project
}

var projects = []Project{
	{
		ID:          "argodash",
		Name:        "ArgoDash",
		Description: "Real-time dashboard: blue dollar, weather, S&P500 and country risk — a Go server aggregating multiple APIs in parallel.",
		Status:      "live",
		Tags:        []string{"REST APIs", "Goroutines", "JSON", "HTTP Client"},
		URL:         "/projects/argodash",
	},
	{
		ID:          "linkchecker",
		Name:        "LinkChecker",
		Description: "Bulk URL checker running in parallel with timeouts, context cancellation and streaming results via SSE.",
		Status:      "live",
		Tags:        []string{"Channels", "Context", "SSE", "Concurrency"},
		URL:         "/projects/linkchecker",
	},
	{
		ID:          "rssaggregator",
		Name:        "RSS Aggregator",
		Description: "Reads N RSS feeds in parallel, deduplicates entries and serves a unified sorted feed.",
		Status:      "planned",
		Tags:        []string{"Goroutines", "Channels", "XML", "Scraping"},
		URL:         "/projects/rssaggregator",
	},
	{
		ID:          "shorturl",
		Name:        "ShortURL",
		Description: "URL shortener with PostgreSQL, click stats and configurable TTL.",
		Status:      "planned",
		Tags:        []string{"PostgreSQL", "REST API", "Middleware"},
		URL:         "/projects/shorturl",
	},
	{
		ID:          "webhooklogger",
		Name:        "WebhookLogger",
		Description: "Server that receives, logs and displays webhooks from any service with paginated history.",
		Status:      "planned",
		Tags:        []string{"HTTP", "PostgreSQL", "Real-time"},
		URL:         "/projects/webhooklogger",
	},
	{
		ID:          "pipelinevis",
		Name:        "PipelineVis",
		Description: "Describe a data pipeline in YAML, Go parses it and renders an interactive diagram.",
		Status:      "planned",
		Tags:        []string{"YAML", "Data Engineering", "Visualization"},
		URL:         "/projects/pipelinevis",
	},
}

func main() {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseGlob("web/templates/*.html"))

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/projects/argodash", argodash.Handler())
	http.Handle("/projects/linkchecker", linkchecker.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Projects: projects}
		if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("golab running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
