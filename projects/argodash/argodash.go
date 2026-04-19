package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type DashboardData struct {
	Dollar    DollarData
	Weather   WeatherData
	SP500     SP500Data
	RiskScore RiskData
	Crypto    CryptoData
	Inflation InflationData
	News      NewsData
	FxRate    FxRateData
	GitHub    GitHubData
	ISS       ISSData
	FetchedAt string
}

type DollarData struct {
	Official float64
	Blue     float64
	Error    string
}

type WeatherData struct {
	Temp      string
	Condition string
	Error     string
}

type SP500Data struct {
	Price  string
	Change string
	Error  string
}

type RiskData struct {
	Value int
	Error string
}

type CryptoData struct {
	Bitcoin float64
	Error   string
}

type InflationData struct {
	Value float64
	Month string
	Error string
}

type NewsData struct {
	Articles []NewsArticle
	Error    string
}

type NewsArticle struct {
	Title  string
	Source string
	URL    string
}

type FxRateData struct {
	EURUSD float64
	Error  string
}

type GitHubData struct {
	Status      string
	Description string
	Error       string
}

type ISSData struct {
	Latitude  string
	Longitude string
	Error     string
}

func fetchDollar(wg *sync.WaitGroup, result *DollarData) {
	defer wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://dolarapi.com/v1/dolares")
	if err != nil {
		result.Error = "could not reach dolarapi.com"
		return
	}
	defer resp.Body.Close()

	var raw []struct {
		Casa   string  `json:"casa"`
		Compra float64 `json:"compra"`
		Venta  float64 `json:"venta"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	for _, item := range raw {
		switch item.Casa {
		case "oficial":
			result.Official = item.Venta
		case "blue":
			result.Blue = item.Venta
		}
	}
}
func fetchWeather(wg *sync.WaitGroup, result *WeatherData) {
	defer wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://wttr.in/Rosario?format=j1")
	if err != nil {
		result.Error = "could not reach wttr.in"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Current []struct {
			TempC       string `json:"temp_C"`
			WeatherDesc []struct {
				Value string `json:"value"`
			} `json:"weatherDesc"`
		} `json:"current_condition"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	if len(raw.Current) > 0 {
		result.Temp = raw.Current[0].TempC + "°C"
		if len(raw.Current[0].WeatherDesc) > 0 {
			result.Condition = raw.Current[0].WeatherDesc[0].Value
		}
	}
}
func fetchSP500(wg *sync.WaitGroup, result *SP500Data) {
	defer wg.Done()

	key := os.Getenv("ALPHA_VANTAGE_KEY")
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=SPY&apikey=%s", key)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		result.Error = "could not reach alphavantage.co"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Quote struct {
			Price  string `json:"05. price"`
			Change string `json:"10. change percent"`
		} `json:"Global Quote"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	result.Price = raw.Quote.Price
	result.Change = raw.Quote.Change
}

func fetchRisk(wg *sync.WaitGroup, result *RiskData) {
	defer wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.argentinadatos.com/v1/finanzas/indices/riesgo-pais/ultimo")
	if err != nil {
		result.Error = "could not reach argentinadatos.com"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Valor int `json:"valor"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	result.Value = raw.Valor
}

func fetchCrypto(wg *sync.WaitGroup, result *CryptoData) {
	defer wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.coinbase.com/v2/prices/BTC-USD/spot")
	if err != nil {
		result.Error = "could not reach coinbase.com"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Data struct {
			Amount string `json:"amount"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	fmt.Sscanf(raw.Data.Amount, "%f", &result.Bitcoin)
}

func fetchInflation(wg *sync.WaitGroup, result *InflationData) {
	defer wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.argentinadatos.com/v1/finanzas/indices/inflacion/ultimo")
	if err != nil {
		result.Error = "could not reach argentinadatos.com"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Valor float64 `json:"valor"`
		Fecha string  `json:"fecha"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	result.Value = raw.Valor
	result.Month = raw.Fecha
}

func fetchNews(wg *sync.WaitGroup, result *NewsData) {
	defer wg.Done()

	key := os.Getenv("NEWS_API_KEY")
	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?category=technology&language=en&pageSize=5&apiKey=%s", key)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		result.Error = "could not reach newsapi.org"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Articles []struct {
			Title  string `json:"title"`
			Source struct {
				Name string `json:"name"`
			} `json:"source"`
			URL string `json:"url"`
		} `json:"articles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	for _, a := range raw.Articles {
		result.Articles = append(result.Articles, NewsArticle{
			Title:  a.Title,
			Source: a.Source.Name,
			URL:    a.URL,
		})
	}
}

func fetchFxRate(wg *sync.WaitGroup, result *FxRateData) {
	defer wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.frankfurter.app/latest?from=EUR&to=USD")
	if err != nil {
		result.Error = "could not reach frankfurter.app"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Rates struct {
			USD float64 `json:"USD"`
		} `json:"rates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	result.EURUSD = raw.Rates.USD
}

func fetchGitHub(wg *sync.WaitGroup, result *GitHubData) {
	defer wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://www.githubstatus.com/api/v2/status.json")
	if err != nil {
		result.Error = "could not reach githubstatus.com"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Status struct {
			Indicator   string `json:"indicator"`
			Description string `json:"description"`
		} `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	result.Status = raw.Status.Indicator
	result.Description = raw.Status.Description
}

func fetchISS(wg *sync.WaitGroup, result *ISSData) {
	defer wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://api.open-notify.org/iss-now.json")
	if err != nil {
		result.Error = "could not reach open-notify.org"
		return
	}
	defer resp.Body.Close()

	var raw struct {
		Position struct {
			Latitude  string `json:"latitude"`
			Longitude string `json:"longitude"`
		} `json:"iss_position"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = "could not parse response"
		return
	}

	result.Latitude = raw.Position.Latitude
	result.Longitude = raw.Position.Longitude
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var (
			dollar    DollarData
			weather   WeatherData
			sp500     SP500Data
			risk      RiskData
			crypto    CryptoData
			inflation InflationData
			news      NewsData
			fxrate    FxRateData
			github    GitHubData
			iss       ISSData
		)

		var wg sync.WaitGroup
		wg.Add(10) // 10 goroutines

		go fetchDollar(&wg, &dollar)
		go fetchWeather(&wg, &weather)
		go fetchSP500(&wg, &sp500)
		go fetchRisk(&wg, &risk)
		go fetchCrypto(&wg, &crypto)
		go fetchInflation(&wg, &inflation)
		go fetchNews(&wg, &news)
		go fetchFxRate(&wg, &fxrate)
		go fetchGitHub(&wg, &github)
		go fetchISS(&wg, &iss)

		wg.Wait()

		data := DashboardData{
			Dollar:    dollar,
			Weather:   weather,
			SP500:     sp500,
			RiskScore: risk,
			Crypto:    crypto,
			Inflation: inflation,
			News:      news,
			FxRate:    fxrate,
			GitHub:    github,
			ISS:       iss,
			FetchedAt: time.Now().Format("15:04:05"),
		}

		tmpl := template.Must(template.ParseFiles("dashboard.html"))
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("argodash running on http://localhost:9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
