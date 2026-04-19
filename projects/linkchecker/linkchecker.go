package linkchecker

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Result struct {
	URL      string
	Status   int
	Ok       bool
	Duration time.Duration
	Error    string
}

type CheckRequest struct {
	URLs []string
}

func checkURL(ctx context.Context, url string, results chan<- Result) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		results <- Result{URL: url, Ok: false, Error: "Invalid URL"}
		return
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		results <- Result{URL: url, Ok: false, Error: "unreachable", Duration: time.Since(start)}
		return
	}
	defer resp.Body.Close()

	results <- Result{
		URL:      url,
		Status:   resp.StatusCode,
		Ok:       resp.StatusCode >= 200 && resp.StatusCode < 300,
		Duration: time.Since(start),
	}
}

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "projects/linkchecker/index.html")
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		rawURLs := r.FormValue("urls")
		if rawURLs == "" {
			http.Error(w, "no URLs provided", http.StatusBadRequest)
			return
		}

		var urls []string
		for _, u := range splitLines(rawURLs) {
			if u != "" {
				urls = append(urls, u)
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		results := make(chan Result, len(urls))

		var wg sync.WaitGroup
		for _, url := range urls {
			wg.Add(1)
			go func(u string) {
				defer wg.Done()
				checkURL(ctx, u, results)
			}(url)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		w.Header().Set("Content-Type", "text/html")
		for result := range results {
			if result.Ok {
				fmt.Fprintf(w, `<div class="result ok">✅ %s — %d (%dms)</div>`, result.URL, result.Status, result.Duration.Milliseconds())
			} else {
				fmt.Fprintf(w, `<div class="result error">❌ %s — %s</div>`, result.URL, result.Error)
			}
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	lines = append(lines, strings.TrimSpace(s[start:]))
	return lines
}
