package rssaggregator

import (
	"encoding/xml"
	"html/template"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Feed struct {
	Channel struct {
		Title string `xml:"title"`
		Items []Item `xml:"item"`
	} `xml:"channel"`
}

type Item struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	Source  string
}

type Result struct {
	Items []Item
	Error string
}

func fetchFeed(url string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		results <- Result{Error: "could not reach " + url}
		return
	}
	defer resp.Body.Close()

	var feed Feed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		results <- Result{Error: "could not parse feed: " + url}
		return
	}

	for i := range feed.Channel.Items {
		feed.Channel.Items[i].Source = feed.Channel.Title
	}

	results <- Result{Items: feed.Channel.Items}
}

func Handler() http.HandlerFunc {
	feeds := []string{
		"https://www.theverge.com/rss/index.xml",
		"https://feeds.arstechnica.com/arstechnica/index",
		"https://hnrss.org/frontpage",
		"https://feeds.feedburner.com/TechCrunch",
		"https://www.wired.com/feed/rss",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		results := make(chan Result, len(feeds))

		var wg sync.WaitGroup
		for _, url := range feeds {
			wg.Add(1)
			go fetchFeed(url, results, &wg)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		var allItems []Item
		for result := range results {
			if result.Error == "" {
				allItems = append(allItems, result.Items...)
			}
		}

		sort.Slice(allItems, func(i, j int) bool {
			ti, _ := time.Parse(time.RFC1123Z, allItems[i].PubDate)
			tj, _ := time.Parse(time.RFC1123Z, allItems[j].PubDate)
			return ti.After(tj)
		})

		seen := make(map[string]bool)
		var unique []Item
		for _, item := range allItems {
			if !seen[item.Link] {
				seen[item.Link] = true
				unique = append(unique, item)
			}
		}

		tmpl, err := template.ParseFiles("projects/rssaggregator/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, unique)
	}
}
