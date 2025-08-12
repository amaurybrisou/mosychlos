package newsapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_NewClient(t *testing.T) {
	client := NewClient("test-api-key")

	if client.apiKey != "test-api-key" {
		t.Errorf("expected apiKey to be 'test-api-key', got %s", client.apiKey)
	}

	if client.baseURL != "https://newsapi.org/v2" {
		t.Errorf("expected baseURL to be 'https://newsapi.org/v2', got %s", client.baseURL)
	}

	if client.http == nil {
		t.Error("expected http client to be initialized")
	}

	if client.http.Timeout != 30*time.Second {
		t.Errorf("expected timeout to be 30s, got %v", client.http.Timeout)
	}
}

func TestClient_GetTopHeadlines(t *testing.T) {
	mockResponse := `{
		"status": "ok",
		"totalResults": 2,
		"articles": [
			{
				"source": {"name": "Test Source 1"},
				"title": "Test Article 1",
				"publishedAt": "2023-01-01T10:00:00Z",
				"url": "https://example.com/article1"
			},
			{
				"source": {"name": "Test Source 2"},
				"title": "Test Article 2",
				"publishedAt": "2023-01-02T11:00:00Z",
				"url": "https://example.com/article2"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/top-headlines" {
			t.Errorf("expected path '/top-headlines', got %s", r.URL.Path)
		}

		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key header 'test-key', got %s", r.Header.Get("X-API-Key"))
		}

		// Check query parameters
		if r.URL.Query().Get("country") != "us" {
			t.Errorf("expected country parameter 'us', got %s", r.URL.Query().Get("country"))
		}

		if r.URL.Query().Get("category") != "business" {
			t.Errorf("expected category parameter 'business', got %s", r.URL.Query().Get("category"))
		}

		if r.URL.Query().Get("pageSize") != "10" {
			t.Errorf("expected pageSize parameter '10', got %s", r.URL.Query().Get("pageSize"))
		}

		if r.URL.Query().Get("language") != "en" {
			t.Errorf("expected language parameter 'en', got %s", r.URL.Query().Get("language"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.baseURL = server.URL

	params := TopHeadlinesParams{
		Country:  "us",
		Category: "business",
		PageSize: 10,
		Language: "en",
	}

	response, err := client.GetTopHeadlines(context.Background(), params)
	if err != nil {
		t.Fatalf("GetTopHeadlines failed: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", response.Status)
	}

	if response.TotalResults != 2 {
		t.Errorf("expected total results 2, got %d", response.TotalResults)
	}

	if len(response.Articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(response.Articles))
	}

	if response.Articles[0].Title != "Test Article 1" {
		t.Errorf("expected first article title 'Test Article 1', got %s", response.Articles[0].Title)
	}

	if response.Articles[0].Source.Name != "Test Source 1" {
		t.Errorf("expected first article source 'Test Source 1', got %s", response.Articles[0].Source.Name)
	}
}

func TestClient_GetEverything(t *testing.T) {
	mockResponse := `{
		"status": "ok",
		"totalResults": 1,
		"articles": [
			{
				"source": {"name": "Everything Source"},
				"title": "Everything Article",
				"publishedAt": "2023-01-01T12:00:00Z",
				"url": "https://example.com/everything"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/everything" {
			t.Errorf("expected path '/everything', got %s", r.URL.Path)
		}

		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key header 'test-key', got %s", r.Header.Get("X-API-Key"))
		}

		// Check query parameters
		if r.URL.Query().Get("q") != "stocks" {
			t.Errorf("expected query parameter 'stocks', got %s", r.URL.Query().Get("q"))
		}

		if r.URL.Query().Get("pageSize") != "20" {
			t.Errorf("expected pageSize parameter '20', got %s", r.URL.Query().Get("pageSize"))
		}

		if r.URL.Query().Get("language") != "en" {
			t.Errorf("expected language parameter 'en', got %s", r.URL.Query().Get("language"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.baseURL = server.URL

	params := EverythingParams{
		Query:    "stocks",
		PageSize: 20,
		Language: "en",
	}

	response, err := client.GetEverything(context.Background(), params)
	if err != nil {
		t.Fatalf("GetEverything failed: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", response.Status)
	}

	if response.TotalResults != 1 {
		t.Errorf("expected total results 1, got %d", response.TotalResults)
	}

	if len(response.Articles) != 1 {
		t.Errorf("expected 1 article, got %d", len(response.Articles))
	}

	if response.Articles[0].Title != "Everything Article" {
		t.Errorf("expected article title 'Everything Article', got %s", response.Articles[0].Title)
	}
}

func TestNewsAPIResponse_ToNewsData(t *testing.T) {
	response := &NewsAPIResponse{
		Status:       "ok",
		TotalResults: 1,
		Articles: []NewsArticle{
			{
				Source: struct {
					Name string `json:"name"`
				}{Name: "Test Source"},
				Title:       "Test Title",
				PublishedAt: "2023-01-01T10:00:00Z",
				URL:         "https://example.com/test",
			},
		},
	}

	newsData := response.ToNewsData()

	if len(newsData.Articles) != 1 {
		t.Errorf("expected 1 article, got %d", len(newsData.Articles))
	}

	article := newsData.Articles[0]
	if article.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %s", article.Title)
	}

	if article.Source != "Test Source" {
		t.Errorf("expected source 'Test Source', got %s", article.Source)
	}

	if article.URL != "https://example.com/test" {
		t.Errorf("expected URL 'https://example.com/test', got %s", article.URL)
	}

	expectedTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	if !article.PublishedAt.Equal(expectedTime) {
		t.Errorf("expected time %v, got %v", expectedTime, article.PublishedAt)
	}

	// Check that LastUpdated is set
	if newsData.LastUpdated.IsZero() {
		t.Error("expected LastUpdated to be set")
	}
}

func TestClient_HTTPErrors(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		method     string
	}{
		{"top headlines 401", http.StatusUnauthorized, "top-headlines"},
		{"top headlines 500", http.StatusInternalServerError, "top-headlines"},
		{"everything 401", http.StatusUnauthorized, "everything"},
		{"everything 500", http.StatusInternalServerError, "everything"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.statusCode)
			}))
			defer server.Close()

			client := NewClient("test-key")
			client.baseURL = server.URL

			var err error
			if c.method == "top-headlines" {
				_, err = client.GetTopHeadlines(context.Background(), TopHeadlinesParams{})
			} else {
				_, err = client.GetEverything(context.Background(), EverythingParams{})
			}

			if err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}
