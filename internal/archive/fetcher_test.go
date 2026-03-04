package archive

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	// Implementation validates exactly one slash exists
	testCases := []string{
		"invalid-no-slash",
		"too/many/slashes",
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

// testTransport is a custom RoundTripper that redirects all requests to a test server
// and optionally captures the request URL for assertions.
type testTransport struct {
	server     *httptest.Server
	capturedURL string
}

func (tt *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Capture the original URL path for assertions
	tt.capturedURL = req.URL.String()
	// Redirect all requests to the test server
	req.URL.Scheme = "http"
	req.URL.Host = tt.server.URL[7:] // Remove "http://" prefix
	return http.DefaultTransport.RoundTrip(req)
}

// mockRoundTripper allows us to mock HTTP responses without a test server.
type mockRoundTripper struct {
	statusCode  int
	body        string
	capturedURL string
	method      string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.capturedURL = req.URL.String()
	m.method = req.Method

	response := &http.Response{
		StatusCode: m.statusCode,
		Status:     http.StatusText(m.statusCode),
		Body:       http.NoBody,
		Header:     make(http.Header),
	}

	if m.body != "" {
		response.Body = io_NopCloserFromString(m.body)
	}

	return response, nil
}

// io_NopCloserFromString creates an io.ReadCloser from a string.
func io_NopCloserFromString(s string) *readCloser {
	return &readCloser{reader: strings.NewReader(s)}
}

type readCloser struct {
	reader *strings.Reader
}

func (rc *readCloser) Read(p []byte) (int, error) {
	return rc.reader.Read(p)
}

func (rc *readCloser) Close() error {
	return nil
}

// validSummaryJSON is a SummaryResponse-wrapped JSON matching wakasync's format.
const validSummaryJSON = `{
  "data": [{
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
  }],
  "cumulative_total": { "digital": "1:00", "seconds": 3600, "text": "1 hr" },
  "daily_average": { "days_including_holidays": 1, "days_minus_holidays": 1, "holidays": 0, "seconds": 3600, "text": "1 hr" },
  "start": "2026-02-24",
  "end": "2026-02-24"
}`

// TestFetchArchive_Success verifies successful JSON parsing with SummaryResponse wrapper.
func TestFetchArchive_Success(t *testing.T) {
	// Create mock server that returns SummaryResponse-wrapped JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validSummaryJSON))
	}))
	defer server.Close()

	f := New("test/repo")
	if f == nil {
		t.Fatal("Expected non-nil fetcher")
	}

	// Replace the httpCli to point to our test server
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

	// Verify parsed data extracted from SummaryResponse wrapper
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

// TestFetchArchive_EmptyDataArray confirms that when SummaryResponse has "data": [],
// FetchArchive returns (nil, nil).
func TestFetchArchive_EmptyDataArray(t *testing.T) {
	emptyDataJSON := `{
  "data": [],
  "cumulative_total": { "digital": "0:00", "seconds": 0, "text": "0 secs" },
  "daily_average": { "days_including_holidays": 0, "days_minus_holidays": 0, "holidays": 0, "seconds": 0, "text": "0 secs" },
  "start": "2026-02-24",
  "end": "2026-02-24"
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(emptyDataJSON))
	}))
	defer server.Close()

	f := New("test/repo")
	if f == nil {
		t.Fatal("Expected non-nil fetcher")
	}
	f.httpCli = &http.Client{
		Transport: &testTransport{server: server},
	}

	data, err := f.FetchArchive("2026-02-24")
	if data != nil {
		t.Errorf("Expected nil data for empty data array, got %v", data)
	}
	if err != nil {
		t.Errorf("Expected nil error for empty data array, got %v", err)
	}
}

// TestFetchArchive_URLPattern verifies the URL constructed by FetchArchive uses
// the correct wakasync path format: data/YYYY/MM/DD/summary.json
func TestFetchArchive_URLPattern(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validSummaryJSON))
	}))
	defer server.Close()

	transport := &testTransport{server: server}
	f := New("owner/repo")
	if f == nil {
		t.Fatal("Expected non-nil fetcher")
	}
	f.httpCli = &http.Client{Transport: transport}

	_, err := f.FetchArchive("2026-02-24")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the captured URL uses the new path format
	expectedPath := "/owner/repo/main/data/2026/02/24/summary.json"
	if !strings.Contains(transport.capturedURL, expectedPath) {
		t.Errorf("URL should contain %q, got %q", expectedPath, transport.capturedURL)
	}

	// Verify it does NOT use the old format
	oldPath := "/data/2026-02-24.json"
	if strings.Contains(transport.capturedURL, oldPath) {
		t.Errorf("URL should NOT contain old format %q, got %q", oldPath, transport.capturedURL)
	}
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

// TestFetchArchive_InvalidDate verifies that a malformed date returns an error.
func TestFetchArchive_InvalidDate(t *testing.T) {
	f := &Fetcher{
		HistoryRepo: "test/repo",
		httpCli: &http.Client{
			Transport: &mockRoundTripper{statusCode: http.StatusOK},
		},
	}

	testCases := []struct {
		name string
		date string
	}{
		{"no dashes", "20260224"},
		{"single dash", "2026-0224"},
		{"too many parts", "2026-02-24-extra"},
		{"empty string", ""},
		{"words", "bad-date"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := f.FetchArchive(tc.date)
			if data != nil {
				t.Errorf("Expected nil data for invalid date %q, got %v", tc.date, data)
			}
			// Most cases should return error (except "bad-date" which has 3 parts but is invalid)
			// The split on "-" produces different part counts for each case
			if tc.date == "bad-date" {
				// "bad-date" splits into ["bad", "date"] = 2 parts, should error
				if err == nil {
					t.Errorf("Expected error for invalid date %q, got nil", tc.date)
				}
			} else if err == nil && tc.date != "" {
				// Non-empty dates with wrong number of parts should error
				// (empty string also has wrong part count: [""] = 1 part)
			}
			// For dates that don't split into 3 parts, we expect an error
			parts := strings.Split(tc.date, "-")
			if len(parts) != 3 && err == nil {
				t.Errorf("Expected error for date %q with %d parts, got nil", tc.date, len(parts))
			}
		})
	}
}

// TestCheckAccess_NilFetcher verifies CheckAccess() on nil fetcher returns nil.
func TestCheckAccess_NilFetcher(t *testing.T) {
	var f *Fetcher
	err := f.CheckAccess()
	if err != nil {
		t.Errorf("Expected nil error for nil fetcher, got %v", err)
	}
}

// TestCheckAccess_Accessible verifies CheckAccess returns nil for accessible repos.
func TestCheckAccess_Accessible(t *testing.T) {
	mock := &mockRoundTripper{statusCode: http.StatusOK}
	f := &Fetcher{
		HistoryRepo: "test/repo",
		httpCli: &http.Client{
			Transport: mock,
		},
	}

	err := f.CheckAccess()
	if err != nil {
		t.Errorf("Expected nil error for accessible repo, got %v", err)
	}

	// Verify it used HEAD method
	if mock.method != "HEAD" {
		t.Errorf("Expected HEAD request, got %s", mock.method)
	}
}

// TestCheckAccess_NotFound verifies CheckAccess returns a helpful error for 404.
func TestCheckAccess_NotFound(t *testing.T) {
	f := &Fetcher{
		HistoryRepo: "test/repo",
		httpCli: &http.Client{
			Transport: &mockRoundTripper{statusCode: http.StatusNotFound},
		},
	}

	err := f.CheckAccess()
	if err == nil {
		t.Fatal("Expected error for 404 repo, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "not accessible") {
		t.Errorf("Error should mention 'not accessible', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "private repo") {
		t.Errorf("Error should mention 'private repo', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "test/repo") {
		t.Errorf("Error should include repo name 'test/repo', got: %s", errMsg)
	}
}
