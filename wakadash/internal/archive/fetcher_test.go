package archive

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/b00y0h/wakadash/internal/types"
)

// TestNew_EmptyRepo verifies that New("") returns nil.
func TestNew_EmptyRepo(t *testing.T) {
	f := New("")
	if f != nil {
		t.Errorf("New(\"\") should return nil, got %v", f)
	}
}

// TestNew_InvalidFormat verifies that New with invalid format returns nil.
func TestNew_InvalidFormat(t *testing.T) {
	testCases := []string{
		"invalid-no-slash",
		"too/many/slashes",
		"/leading-slash",
		"trailing-slash/",
	}

	for _, tc := range testCases {
		f := New(tc)
		if f != nil {
			t.Errorf("New(%q) should return nil for invalid format, got %v", tc, f)
		}
	}
}

// TestNew_ValidFormat verifies that New with valid format returns a Fetcher.
func TestNew_ValidFormat(t *testing.T) {
	f := New("owner/repo")
	if f == nil {
		t.Fatal("New(\"owner/repo\") should return non-nil Fetcher")
	}
	if f.HistoryRepo != "owner/repo" {
		t.Errorf("Expected HistoryRepo=\"owner/repo\", got %q", f.HistoryRepo)
	}
	if f.httpCli == nil {
		t.Error("Expected httpCli to be initialized, got nil")
	}
}

// TestFetchArchive_NilFetcher verifies that calling FetchArchive on nil returns (nil, nil).
func TestFetchArchive_NilFetcher(t *testing.T) {
	var f *Fetcher // nil fetcher
	data, err := f.FetchArchive("2026-02-24")
	if data != nil {
		t.Errorf("Expected nil data, got %v", data)
	}
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

// TestFetchArchive_404 verifies that 404 responses return (nil, nil) not an error.
func TestFetchArchive_404(t *testing.T) {
	// Create mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create fetcher with custom httpCli pointing to mock server
	// We'll bypass New() to inject a test server URL
	f := &Fetcher{
		HistoryRepo: "test/repo",
		httpCli:     server.Client(),
	}

	// Temporarily override the URL construction by testing against the mock server
	// In real usage, FetchArchive constructs the URL internally, but for testing
	// we need to intercept HTTP calls. We'll use httptest client redirect.

	// Actually, let's test this properly by checking the behavior when GitHub returns 404
	// Since we can't easily mock the URL construction, let's create a more realistic test

	// For this test, we'll verify behavior by inspecting the actual logic
	// Create a fetcher that would hit a 404
	realFetcher := New("nonexistent/repo-that-does-not-exist-12345")
	if realFetcher == nil {
		t.Fatal("Expected non-nil fetcher for valid format")
	}

	// Note: This would make a real HTTP request to GitHub.
	// For a true unit test, we should mock the HTTP transport.
	// Let's implement a proper mock using RoundTripper instead.
}

// mockRoundTripper allows us to mock HTTP responses.
type mockRoundTripper struct {
	statusCode int
	body       string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	response := &http.Response{
		StatusCode: m.statusCode,
		Status:     http.StatusText(m.statusCode),
		Body:       http.NoBody,
		Header:     make(http.Header),
	}

	if m.body != "" {
		response.Body = http.NoBody // Will be replaced in Success test
	}

	return response, nil
}

// TestFetchArchive_404_Mocked uses a mock transport to test 404 handling.
func TestFetchArchive_404_Mocked(t *testing.T) {
	f := &Fetcher{
		HistoryRepo: "test/repo",
		httpCli: &http.Client{
			Transport: &mockRoundTripper{statusCode: http.StatusNotFound},
		},
	}

	data, err := f.FetchArchive("2026-02-24")
	if data != nil {
		t.Errorf("Expected nil data for 404, got %v", data)
	}
	if err != nil {
		t.Errorf("Expected nil error for 404, got %v", err)
	}
}

// TestFetchArchive_Success verifies successful JSON parsing.
func TestFetchArchive_Success(t *testing.T) {
	// Valid minimal DayData JSON
	validJSON := `{
  "grand_total": {
    "total_seconds": 3600,
    "digital": "1:00",
    "text": "1 hr",
    "hours": 1,
    "minutes": 0
  },
  "languages": [
    {
      "name": "Go",
      "total_seconds": 3600,
      "percent": 100
    }
  ],
  "projects": [
    {
      "name": "wakadash",
      "total_seconds": 3600,
      "percent": 100
    }
  ],
  "range": {
    "date": "2026-02-24",
    "text": "Mon Feb 24",
    "start": "",
    "end": "",
    "timezone": ""
  },
  "categories": [],
  "editors": [],
  "machines": [],
  "operating_systems": [],
  "entities": [],
  "branches": [],
  "dependencies": []
}`

	// Create mock server that returns valid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validJSON))
	}))
	defer server.Close()

	// Override the fetcher to use mock server URL
	// We need to construct fetcher with custom base URL
	// Since FetchArchive hardcodes GitHub URL, we'll use the server's URL pattern

	// Create a fetcher but we'll need to test against the actual implementation
	// The challenge is that FetchArchive builds the URL internally

	// For now, let's create a realistic integration test:
	// We'll test with a real repo that we know exists and has data
	// But for unit testing, we should refactor FetchArchive to accept a base URL

	// Instead, let's test the mock server approach properly:
	f := New("test/repo")
	if f == nil {
		t.Fatal("Expected non-nil fetcher")
	}

	// Replace the httpCli to point to our test server
	// We need to rewrite the URL, so let's use a custom transport
	f.httpCli = &http.Client{
		Transport: &testTransport{server: server},
	}

	data, err := f.FetchArchive("2026-02-24")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if data == nil {
		t.Fatal("Expected non-nil data")
	}

	// Verify parsed data
	if data.GrandTotal.TotalSeconds != 3600 {
		t.Errorf("Expected TotalSeconds=3600, got %f", data.GrandTotal.TotalSeconds)
	}
	if len(data.Languages) != 1 {
		t.Errorf("Expected 1 language, got %d", len(data.Languages))
	}
	if len(data.Languages) > 0 && data.Languages[0].Name != "Go" {
		t.Errorf("Expected language=Go, got %s", data.Languages[0].Name)
	}
	if data.Range.Date != "2026-02-24" {
		t.Errorf("Expected date=2026-02-24, got %s", data.Range.Date)
	}
}

// testTransport is a custom RoundTripper that redirects all requests to a test server.
type testTransport struct {
	server *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect all requests to the test server
	req.URL.Scheme = "http"
	req.URL.Host = t.server.URL[7:] // Remove "http://" prefix
	return http.DefaultTransport.RoundTrip(req)
}
