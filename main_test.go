package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// --- Mock Implementations ---

// MockGitHubClient simulates the GitHub API client.
type MockGitHubClient struct {
	// GetFunc allows customizing the Get behavior for different paths.
	GetFunc func(path string, response interface{}) error
	// GetCalls records the paths called with Get.
	GetCalls []string
}

func (m *MockGitHubClient) Get(path string, response interface{}) error {
	m.GetCalls = append(m.GetCalls, path)
	if m.GetFunc != nil {
		return m.GetFunc(path, response)
	}
	// Default behavior: return empty response
	// You might want to return specific errors or data based on path in GetFunc
	return nil
}

// MockGraphQLClient simulates the GitHub GraphQL API client.
type MockGraphQLClient struct {
	DoFunc func(query string, variables map[string]interface{}, response interface{}) error
	DoCalls []string
}

func (m *MockGraphQLClient) Do(query string, variables map[string]interface{}, response interface{}) error {
	m.DoCalls = append(m.DoCalls, query)
	if m.DoFunc != nil {
		return m.DoFunc(query, variables, response)
	}
	return nil
}

// MockTokenFetcher simulates fetching an auth token.
type MockTokenFetcher struct {
	TokenToReturn string
	ErrorToReturn error
	FetchCount    int
}

func (m *MockTokenFetcher) FetchToken() (string, error) {
	m.FetchCount++
	return m.TokenToReturn, m.ErrorToReturn
}

// MockSummarizer simulates the AI summarization service.
type MockSummarizer struct {
	SummaryToReturn string
	ErrorToReturn   error
	SummarizeCalls  []string // Record the text passed to Summarize
}

func (m *MockSummarizer) Summarize(text string) (string, error) {
	m.SummarizeCalls = append(m.SummarizeCalls, text)
	return m.SummaryToReturn, m.ErrorToReturn
}

// --- Test Helper Functions ---

// captureOutput captures stdout and stderr during a function execution.
func captureOutput(f func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	f() // Execute the function

	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut bytes.Buffer
	var bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)
	return bufOut.String(), bufErr.String()
}

// resetFlags resets the flag package for clean tests.
func resetFlags() {
	// Resetting global flags requires care or using a library.
	// For simplicity here, we'll manually reset the global vars used by flags.
	// NOTE: This is not robust if other packages define flags.
	// Consider using specific flag sets or test setup/teardown for more complex scenarios.
	debug = false
	since = time.Now().AddDate(0, 0, -30).Format(dateFormat) // Reset to default
	bodyOnly = false
}

// --- Test Functions ---

func TestHandlePullsCommand_CSV(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"pulls", testLogin}

	// Mock the API response
	mockClient.GetFunc = func(path string, response interface{}) error {
		// Match percent-encoded PR search URL
		if strings.Contains(path, "search/issues?q=") && strings.Contains(path, "is%3Apr") && strings.Contains(path, "author%3Atestuser") && strings.Contains(path, "page=1") {
			resp := GitHubResponse{
				TotalCount: 1,
				Items: []GitHubItem{
					{Number: 123, Title: "Test PR", HTMLURL: "http://example.com/pr/123", State: "open", Repository: struct {
						Name string `json:"name"`
					}{"test-repo"}},
				},
			}
			// Simulate JSON marshaling and unmarshaling
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handlePullsCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedHeader := "URL,Title,State"
	expectedRow := "http://example.com/pr/123 ,Test PR,open"

	if !strings.Contains(stdout, expectedHeader) {
		t.Errorf("Expected stdout to contain header '%s', got: %s", expectedHeader, stdout)
	}
	if !strings.Contains(stdout, expectedRow) {
		t.Errorf("Expected stdout to contain row '%s', got: %s", expectedRow, stdout)
	}
	if len(mockClient.GetCalls) != 1 {
		t.Errorf("Expected 1 API call, got %d", len(mockClient.GetCalls))
	}
}

func TestHandlePullsCommand_BodyOnly(t *testing.T) {
	resetFlags()
	bodyOnly = true // Set the flag for this test
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"pulls", testLogin}

	mockClient.GetFunc = func(path string, response interface{}) error {
		// Match percent-encoded PR search URL
		if strings.Contains(path, "search/issues?q=") && strings.Contains(path, "is%3Apr") && strings.Contains(path, "author%3Atestuser") && strings.Contains(path, "page=1") {
			resp := GitHubResponse{
				TotalCount: 1,
				Items: []GitHubItem{
					{Number: 123, Title: "Test PR", Body: "This is the body.", HTMLURL: "http://example.com/pr/123", State: "open", Repository: struct {
						Name string `json:"name"`
					}{"test-repo"}},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handlePullsCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedOutput := fmt.Sprintf("%s\n%s #%d\n%s\n%s\n%s\n", startOfPR, "Test PR", 123, "This is the body.", endOfPR, entryDelimiter)

	if stdout != expectedOutput {
		t.Errorf("Expected stdout to be:\n%s\nGot:\n%s", expectedOutput, stdout)
	}
	if len(mockClient.GetCalls) != 1 {
		t.Errorf("Expected 1 API call, got %d", len(mockClient.GetCalls))
	}
}

func TestHandleIssuesCommand_CSV(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"issues", testLogin}

	mockClient.GetFunc = func(path string, response interface{}) error {
		// Match percent-encoded Issue search URL
		if strings.Contains(path, "search/issues?q=") && strings.Contains(path, "is%3Aissue") && strings.Contains(path, "author%3Atestuser") && strings.Contains(path, "page=1") {
			resp := GitHubResponse{
				TotalCount: 1,
				Items: []GitHubItem{
					{Number: 456, Title: "Test Issue", HTMLURL: "http://example.com/issue/456", State: "closed", Repository: struct {
						Name string `json:"name"`
					}{"another-repo"}},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handleIssuesCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedHeader := "URL,Title,State"
	expectedRow := "http://example.com/issue/456 ,Test Issue,closed"

	if !strings.Contains(stdout, expectedHeader) {
		t.Errorf("Expected stdout to contain header '%s', got: %s", expectedHeader, stdout)
	}
	if !strings.Contains(stdout, expectedRow) {
		t.Errorf("Expected stdout to contain row '%s', got: %s", expectedRow, stdout)
	}
	if len(mockClient.GetCalls) != 1 {
		t.Errorf("Expected 1 API call, got %d", len(mockClient.GetCalls))
	}
}

func TestHandleAllCommand_CSV(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	mockGQLClient := &MockGraphQLClient{}
	testLogin := "testuser"
	testArgs := []string{"all", testLogin}

	mockClient.GetFunc = func(path string, response interface{}) error {
		var items []GitHubItem
		// Match percent-encoded PR URL
		if strings.Contains(path, "search/issues?q=") && strings.Contains(path, "is%3Apr") && strings.Contains(path, "author%3Atestuser") && strings.Contains(path, "page=1") {
			items = []GitHubItem{
				{Number: 123, Title: "Test PR", HTMLURL: "http://example.com/pr/123", State: "open", Repository: struct {
					Name string `json:"name"`
				}{"test-repo"}},
			}
		} else if strings.Contains(path, "search/issues?q=") && strings.Contains(path, "is%3Apr") && strings.Contains(path, "reviewed-by%3Atestuser") && strings.Contains(path, "page=1") {
			items = []GitHubItem{
				{Number: 789, Title: "Reviewed PR", HTMLURL: "http://example.com/pr/789", State: "closed", Repository: struct {
					Name string `json:"name"`
				}{"review-repo"}},
			}
		} else if strings.Contains(path, "search/issues?q=") && strings.Contains(path, "is%3Aissue") && strings.Contains(path, "author%3Atestuser") && strings.Contains(path, "page=1") {
			items = []GitHubItem{
				{Number: 456, Title: "Test Issue", HTMLURL: "http://example.com/issue/456", State: "closed", Repository: struct {
					Name string `json:"name"`
				}{"another-repo"}},
			}
		} else {
			return fmt.Errorf("unexpected API call: %s", path)
		}

		resp := GitHubResponse{TotalCount: len(items), Items: items}
		data, _ := json.Marshal(resp)
		return json.Unmarshal(data, response)
	}

	// Mock GraphQL client returns one discussion
	mockGQLClient.DoFunc = func(query string, variables map[string]interface{}, response interface{}) error {
		resp := response.(*DiscussionSearchResponse)
		resp.Search.Nodes = []struct {
			Title     string `json:"title"`
			URL       string `json:"url"`
			Body      string `json:"body"`
			Number    int    `json:"number"`
			CreatedAt string `json:"createdAt"`
			ClosedAt  string `json:"closedAt"`
			Closed    bool   `json:"closed"`
		}{
			{Title: "Test Discussion", URL: "http://example.com/discussion/1", Number: 1, Closed: false, CreatedAt: "2025-01-20T00:00:00Z"},
		}
		resp.Search.PageInfo.HasNextPage = false
		return nil
	}

	stdout, stderr := captureOutput(func() {
		handleAllCommand(testArgs, mockClient, mockGQLClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedHeader := "Type,URL,Title,State"
	expectedPRRow := "Pull Request,http://example.com/pr/123 ,Test PR,open"
	expectedReviewRow := "Review,http://example.com/pr/789 ,Reviewed PR,closed"
	expectedIssueRow := "Issue,http://example.com/issue/456 ,Test Issue,closed"
	expectedDiscussionRow := "Discussion,http://example.com/discussion/1 ,Test Discussion,open"

	if !strings.Contains(stdout, expectedHeader) {
		t.Errorf("Expected stdout to contain header '%s', got: %s", expectedHeader, stdout)
	}
	if !strings.Contains(stdout, expectedPRRow) {
		t.Errorf("Expected stdout to contain PR row '%s', got: %s", expectedPRRow, stdout)
	}
	if !strings.Contains(stdout, expectedReviewRow) {
		t.Errorf("Expected stdout to contain Review row '%s', got: %s", expectedReviewRow, stdout)
	}
	if !strings.Contains(stdout, expectedIssueRow) {
		t.Errorf("Expected stdout to contain Issue row '%s', got: %s", expectedIssueRow, stdout)
	}
	if !strings.Contains(stdout, expectedDiscussionRow) {
		t.Errorf("Expected stdout to contain Discussion row '%s', got: %s", expectedDiscussionRow, stdout)
	}
	if len(mockClient.GetCalls) != 3 { // One for PRs, one for Reviews, one for Issues
		t.Errorf("Expected 3 API calls, got %d", len(mockClient.GetCalls))
	}
}

func TestHandleReviewsCommand_CSV(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"reviews", testLogin}

	mockClient.GetFunc = func(path string, response interface{}) error {
		if strings.Contains(path, "search/issues?q=") && strings.Contains(path, "is%3Apr") && strings.Contains(path, "reviewed-by%3Atestuser") && strings.Contains(path, "page=1") {
			resp := GitHubResponse{
				TotalCount: 1,
				Items: []GitHubItem{
					{Number: 789, Title: "Reviewed PR", HTMLURL: "http://example.com/pr/789", State: "closed", Repository: struct {
						Name string `json:"name"`
					}{"review-repo"}},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handleReviewsCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedHeader := "URL,Title,State"
	expectedRow := "http://example.com/pr/789 ,Reviewed PR,closed"

	if !strings.Contains(stdout, expectedHeader) {
		t.Errorf("Expected stdout to contain header '%s', got: %s", expectedHeader, stdout)
	}
	if !strings.Contains(stdout, expectedRow) {
		t.Errorf("Expected stdout to contain row '%s', got: %s", expectedRow, stdout)
	}
	if len(mockClient.GetCalls) != 1 {
		t.Errorf("Expected 1 API call, got %d", len(mockClient.GetCalls))
	}
}

func TestHandleReviewsCommand_BodyOnly(t *testing.T) {
	resetFlags()
	bodyOnly = true
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"reviews", testLogin}

	mockClient.GetFunc = func(path string, response interface{}) error {
		if strings.Contains(path, "search/issues?q=") && strings.Contains(path, "is%3Apr") && strings.Contains(path, "reviewed-by%3Atestuser") && strings.Contains(path, "page=1") {
			resp := GitHubResponse{
				TotalCount: 1,
				Items: []GitHubItem{
					{Number: 789, Title: "Reviewed PR", Body: "Review body.", HTMLURL: "http://example.com/pr/789", State: "closed", Repository: struct {
						Name string `json:"name"`
					}{"review-repo"}},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handleReviewsCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedOutput := fmt.Sprintf("%s\n%s #%d\n%s\n%s\n%s\n", startOfReview, "Reviewed PR", 789, "Review body.", endOfReview, entryDelimiter)

	if stdout != expectedOutput {
		t.Errorf("Expected stdout to be:\n%s\nGot:\n%s", expectedOutput, stdout)
	}
}

func TestBuildReviewQuery(t *testing.T) {
	resetFlags()
	testLogin := "testuser"

	originalOrgConfigFunc := orgConfigFunc
	orgConfigFunc = func() (string, error) {
		return "github", nil
	}
	defer func() { orgConfigFunc = originalOrgConfigFunc }()

	since = "2025-01-15"
	expected := "is%3Apr+org%3Agithub+reviewed-by%3Atestuser+sort%3Acreated-desc+created%3A%3E2025-01-15"
	actual := buildReviewQuery(testLogin)
	if actual != expected {
		t.Errorf("Expected query '%s', got '%s'", expected, actual)
	}
}

func TestDeduplicateItems(t *testing.T) {
	existing := []GitHubItem{
		{HTMLURL: "http://example.com/pr/1"},
		{HTMLURL: "http://example.com/pr/2"},
	}
	candidates := []GitHubItem{
		{HTMLURL: "http://example.com/pr/2", Title: "Duplicate"},
		{HTMLURL: "http://example.com/pr/3", Title: "Unique"},
	}

	result := deduplicateItems(existing, candidates)

	if len(result) != 1 {
		t.Fatalf("Expected 1 item after dedup, got %d", len(result))
	}
	if result[0].HTMLURL != "http://example.com/pr/3" {
		t.Errorf("Expected unique item, got %s", result[0].HTMLURL)
	}
}

func TestHandleSummarizeCommand(t *testing.T) {
	resetFlags()
	mockSummarizer := &MockSummarizer{
		SummaryToReturn: "This is the summary.",
	}
	testArgs := []string{"summarize"} // Input will come from stdin

	// Prepare stdin
	inputBody := "Some text to summarize" + entryDelimiter + "Another piece of text"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.WriteString(inputBody)
	w.Close()
	defer func() { os.Stdin = oldStdin }() // Restore stdin

	stdout, stderr := captureOutput(func() {
		handleSummarizeCommand(testArgs, mockSummarizer, false)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedOutput := `This is the summary.
This is the summary.
`
	if stdout != expectedOutput {
		t.Errorf("Expected stdout:\n%s\nGot:\n%s", expectedOutput, stdout)
	}

	if len(mockSummarizer.SummarizeCalls) != 2 {
		t.Errorf("Expected Summarize to be called 2 times, got %d", len(mockSummarizer.SummarizeCalls))
	}
	if mockSummarizer.SummarizeCalls[0] != "Some text to summarize" {
		t.Errorf("Expected first summarize call with 'Some text to summarize', got '%s'", mockSummarizer.SummarizeCalls[0])
	}
	if mockSummarizer.SummarizeCalls[1] != "Another piece of text" {
		t.Errorf("Expected second summarize call with 'Another piece of text', got '%s'", mockSummarizer.SummarizeCalls[1])
	}
}

func TestBuildQuery(t *testing.T) {
	resetFlags()
	testLogin := "testuser"

	// Mock orgConfigFunc to return 'github'
	originalOrgConfigFunc := orgConfigFunc
	orgConfigFunc = func() (string, error) {
		return "github", nil
	}
	defer func() { orgConfigFunc = originalOrgConfigFunc }()

	// Test default query includes since filter
	expected := fmt.Sprintf("is%%3Apr+org%%3Agithub+author%%3Atestuser+sort%%3Acreated-desc+created%%3A%%3E%s", since)
	actual := buildQuery("is:pr", testLogin)
	if actual != expected {
		t.Errorf("Expected query '%s', got '%s'", expected, actual)
	}

	// Test with custom since flag
	since = "2025-01-15"
	expectedSince := "is%3Aissue+org%3Agithub+author%3Atestuser+sort%3Acreated-desc+created%3A%3E2025-01-15"
	actualSince := buildQuery("is:issue", testLogin)
	if actualSince != expectedSince {
		t.Errorf("Expected query '%s', got '%s'", expectedSince, actualSince)
	}
}

func TestBuildWebURL(t *testing.T) {
	resetFlags()
	testLogin := "testuser"

	// Mock orgConfigFunc to return 'testorg'
	originalOrgConfigFunc := orgConfigFunc
	orgConfigFunc = func() (string, error) {
		return "testorg", nil
	}
	defer func() { orgConfigFunc = originalOrgConfigFunc }()

	// Mock timeNowFunc to return a fixed date for predictable tests
	originalTimeNowFunc := timeNowFunc
	testDate := time.Date(2025, 1, 30, 12, 0, 0, 0, time.UTC)
	timeNowFunc = func() time.Time {
		return testDate
	}
	defer func() { timeNowFunc = originalTimeNowFunc }()

	// Test basic URL without since filter and without item type filter (the new behavior)
	since = ""
	expected := "https://github.com/issues?q=org%3Atestorg+author%3Atestuser+sort%3Aupdated-desc"
	actual := buildWebURL("", testLogin)
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}

	// Test URL with since filter and without item type filter - now uses date range
	since = "2025-01-15"
	expectedWithSince := "https://github.com/issues?q=org%3Atestorg+author%3Atestuser+sort%3Aupdated-desc+created%3A2025-01-15..2025-01-30"
	actualWithSince := buildWebURL("", testLogin)
	if actualWithSince != expectedWithSince {
		t.Errorf("Expected URL '%s', got '%s'", expectedWithSince, actualWithSince)
	}

	// Test with issue type (for backward compatibility)
	since = ""
	expectedIssue := "https://github.com/issues?q=is%3Aissue+org%3Atestorg+author%3Atestuser+sort%3Aupdated-desc"
	actualIssue := buildWebURL("is:issue", testLogin)
	if actualIssue != expectedIssue {
		t.Errorf("Expected URL '%s', got '%s'", expectedIssue, actualIssue)
	}

	// Test with different item type
	since = ""
	expectedPR := "https://github.com/issues?q=is%3Apr+org%3Atestorg+author%3Atestuser+sort%3Aupdated-desc"
	actualPR := buildWebURL("is:pr", testLogin)
	if actualPR != expectedPR {
		t.Errorf("Expected URL '%s', got '%s'", expectedPR, actualPR)
	}

	// Test date range format with item type
	since = "2025-01-20"
	expectedPRWithRange := "https://github.com/issues?q=is%3Apr+org%3Atestorg+author%3Atestuser+sort%3Aupdated-desc+created%3A2025-01-20..2025-01-30"
	actualPRWithRange := buildWebURL("is:pr", testLogin)
	if actualPRWithRange != expectedPRWithRange {
		t.Errorf("Expected URL '%s', got '%s'", expectedPRWithRange, actualPRWithRange)
	}
}

func TestGetEffectiveOrg(t *testing.T) {
	resetFlags()

	t.Run("OrgFlagOverridesConfig", func(t *testing.T) {
		orgFlag = "test-org-flag"
		defer func() { orgFlag = "" }() // Reset orgFlag after the test

		org := getEffectiveOrg()
		if org != "test-org-flag" {
			t.Errorf("Expected org 'test-org-flag', got '%s'", org)
		}
	})

	t.Run("ConfigOrgUsedWhenNoFlag", func(t *testing.T) {
		orgFlag = "" // Ensure no flag is set
		originalOrgConfigFunc := orgConfigFunc
		orgConfigFunc = func() (string, error) {
			return "test-config-org", nil
		}
		defer func() { orgConfigFunc = originalOrgConfigFunc }()

		org := getEffectiveOrg()
		if org != "test-config-org" {
			t.Errorf("Expected org 'test-config-org', got '%s'", org)
		}
	})

	t.Run("DefaultOrgUsedWhenNoFlagOrConfig", func(t *testing.T) {
		orgFlag = "" // Ensure no flag is set
		originalOrgConfigFunc := orgConfigFunc
		orgConfigFunc = func() (string, error) {
			return "", fmt.Errorf("no org configured")
		}
		defer func() { orgConfigFunc = originalOrgConfigFunc }()

		org := getEffectiveOrg()
		if org != defaultOrg {
			t.Errorf("Expected default org '%s', got '%s'", defaultOrg, org)
		}
	})
}

func TestGetModelFromConfig(t *testing.T) {
	t.Run("ModelConfigured", func(t *testing.T) {
		// Mock the config file with a model configured
		mockConfig := `extensions:
  gh-contrib:
    model: test-model`
		mockConfigPath := "mock_config.yml"
		if err := os.WriteFile(mockConfigPath, []byte(mockConfig), 0644); err != nil {
			t.Fatalf("Failed to write mock config file: %v", err)
		}
		defer os.Remove(mockConfigPath)

		// Temporarily override the config path
		originalPath := os.Getenv("GH_CONFIG_PATH")
		os.Setenv("GH_CONFIG_PATH", mockConfigPath)
		defer os.Setenv("GH_CONFIG_PATH", originalPath)

		model := getModelFromConfig()
		if model != "test-model" {
			t.Errorf("Expected model 'test-model', got '%s'", model)
		}
	})

	t.Run("ModelNotConfigured", func(t *testing.T) {
		// Mock the config file without a model configured
		mockConfig := `extensions:
  gh-contrib: {}`
		mockConfigPath := "mock_config.yml"
		if err := os.WriteFile(mockConfigPath, []byte(mockConfig), 0644); err != nil {
			t.Fatalf("Failed to write mock config file: %v", err)
		}
		defer os.Remove(mockConfigPath)

		// Temporarily override the config path
		originalPath := os.Getenv("GH_CONFIG_PATH")
		os.Setenv("GH_CONFIG_PATH", mockConfigPath)
		defer os.Setenv("GH_CONFIG_PATH", originalPath)

		model := getModelFromConfig()
		if model != defaultModel {
			t.Errorf("Expected default model '%s', got '%s'", defaultModel, model)
		}
	})

	t.Run("ConfigFileMissing", func(t *testing.T) {
		// Temporarily override the config path to a non-existent file
		originalPath := os.Getenv("GH_CONFIG_PATH")
		os.Setenv("GH_CONFIG_PATH", "non_existent_config.yml")
		defer os.Setenv("GH_CONFIG_PATH", originalPath)

		model := getModelFromConfig()
		if model != defaultModel {
			t.Errorf("Expected default model '%s', got '%s'", defaultModel, model)
		}
	})
}

func TestGetEffectiveModel(t *testing.T) {
	t.Run("ModelFlagOverridesConfig", func(t *testing.T) {
		modelFlag = "test-model-flag"
		defer func() { modelFlag = "" }() // Reset modelFlag after the test

		model := getEffectiveModel()
		if model != "test-model-flag" {
			t.Errorf("Expected model 'test-model-flag', got '%s'", model)
		}
	})

	t.Run("ConfigModelUsedWhenNoFlag", func(t *testing.T) {
		modelFlag = "" // Ensure no flag is set
		originalModelConfigFunc := modelConfigFunc
		modelConfigFunc = func() string {
			return "test-config-model"
		}
		defer func() { modelConfigFunc = originalModelConfigFunc }()

		// Ensure the mock is applied before calling getEffectiveModel
		model := getEffectiveModel()
		if model != "test-config-model" {
			t.Errorf("Expected model 'test-config-model', got '%s'", model)
		}
	})

	t.Run("DefaultModelUsedWhenNoFlagOrConfig", func(t *testing.T) {
		modelFlag = "" // Ensure no flag is set
		originalModelConfigFunc := modelConfigFunc
		modelConfigFunc = func() string {
			return defaultModel
		}
		defer func() { modelConfigFunc = originalModelConfigFunc }()

		model := getEffectiveModel()
		if model != defaultModel {
			t.Errorf("Expected default model '%s', got '%s'", defaultModel, model)
		}
	})
}

func TestBuildPrompt(t *testing.T) {
	text := "Some contribution text"
	result := BuildPrompt(text)

	// Verify system prompt is included
	if !strings.Contains(result, "System:") {
		t.Error("Expected prompt to contain 'System:' prefix")
	}
	if !strings.Contains(result, systemPrompt) {
		t.Error("Expected prompt to contain the system prompt")
	}

	// Verify user prompt is included with the text
	if !strings.Contains(result, "User:") {
		t.Error("Expected prompt to contain 'User:' prefix")
	}
	if !strings.Contains(result, text) {
		t.Errorf("Expected prompt to contain the input text '%s'", text)
	}
}

func TestHandleSummarizeCommand_PromptOnly(t *testing.T) {
	resetFlags()
	mockSummarizer := &MockSummarizer{
		SummaryToReturn: "This should not appear.",
	}
	testArgs := []string{"summarize"} // Input will come from stdin

	// Prepare stdin
	inputBody := "Some text to summarize"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.WriteString(inputBody)
	w.Close()
	defer func() { os.Stdin = oldStdin }() // Restore stdin

	stdout, stderr := captureOutput(func() {
		handleSummarizeCommand(testArgs, mockSummarizer, true)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	// Should contain the system prompt and user prompt
	if !strings.Contains(stdout, "System:") {
		t.Error("Expected stdout to contain 'System:' prefix")
	}
	if !strings.Contains(stdout, systemPrompt) {
		t.Error("Expected stdout to contain the system prompt")
	}
	if !strings.Contains(stdout, "User:") {
		t.Error("Expected stdout to contain 'User:' prefix")
	}
	if !strings.Contains(stdout, inputBody) {
		t.Errorf("Expected stdout to contain input text '%s'", inputBody)
	}

	// Summarizer should NOT have been called
	if len(mockSummarizer.SummarizeCalls) != 0 {
		t.Errorf("Expected Summarize to NOT be called, but it was called %d times", len(mockSummarizer.SummarizeCalls))
	}
}

func TestHandleSummarizeCommand_PromptOnlyMultipleEntries(t *testing.T) {
	resetFlags()
	mockSummarizer := &MockSummarizer{
		SummaryToReturn: "This should not appear.",
	}
	testArgs := []string{"summarize"} // Input will come from stdin

	// Prepare stdin with multiple entries
	inputBody := "First entry" + entryDelimiter + "Second entry"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.WriteString(inputBody)
	w.Close()
	defer func() { os.Stdin = oldStdin }() // Restore stdin

	stdout, stderr := captureOutput(func() {
		handleSummarizeCommand(testArgs, mockSummarizer, true)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	// Should contain both entries in the prompt output
	if !strings.Contains(stdout, "First entry") {
		t.Error("Expected stdout to contain 'First entry'")
	}
	if !strings.Contains(stdout, "Second entry") {
		t.Error("Expected stdout to contain 'Second entry'")
	}

	// Count the number of "System:" occurrences (should be 2, one per entry)
	systemCount := strings.Count(stdout, "System:")
	if systemCount != 2 {
		t.Errorf("Expected 2 'System:' occurrences, got %d", systemCount)
	}

	// Summarizer should NOT have been called
	if len(mockSummarizer.SummarizeCalls) != 0 {
		t.Errorf("Expected Summarize to NOT be called, but it was called %d times", len(mockSummarizer.SummarizeCalls))
	}
}

func TestHandleDiscussionsCommand_CSV(t *testing.T) {
	resetFlags()
	mockGQLClient := &MockGraphQLClient{}
	testLogin := "testuser"
	testArgs := []string{"discussions", testLogin}

	mockGQLClient.DoFunc = func(query string, variables map[string]interface{}, response interface{}) error {
		resp := response.(*DiscussionSearchResponse)
		resp.Search.Nodes = []struct {
			Title     string `json:"title"`
			URL       string `json:"url"`
			Body      string `json:"body"`
			Number    int    `json:"number"`
			CreatedAt string `json:"createdAt"`
			ClosedAt  string `json:"closedAt"`
			Closed    bool   `json:"closed"`
		}{
			{Title: "Test Discussion", URL: "http://example.com/discussion/1", Number: 1, Closed: false, CreatedAt: "2025-01-20T00:00:00Z"},
		}
		resp.Search.PageInfo.HasNextPage = false
		return nil
	}

	stdout, stderr := captureOutput(func() {
		handleDiscussionsCommand(testArgs, mockGQLClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedHeader := "URL,Title,State"
	expectedRow := "http://example.com/discussion/1 ,Test Discussion,open"

	if !strings.Contains(stdout, expectedHeader) {
		t.Errorf("Expected stdout to contain header '%s', got: %s", expectedHeader, stdout)
	}
	if !strings.Contains(stdout, expectedRow) {
		t.Errorf("Expected stdout to contain row '%s', got: %s", expectedRow, stdout)
	}
	if len(mockGQLClient.DoCalls) != 1 {
		t.Errorf("Expected 1 GraphQL call, got %d", len(mockGQLClient.DoCalls))
	}
}

func TestHandleDiscussionsCommand_BodyOnly(t *testing.T) {
	resetFlags()
	bodyOnly = true
	mockGQLClient := &MockGraphQLClient{}
	testLogin := "testuser"
	testArgs := []string{"discussions", testLogin}

	mockGQLClient.DoFunc = func(query string, variables map[string]interface{}, response interface{}) error {
		resp := response.(*DiscussionSearchResponse)
		resp.Search.Nodes = []struct {
			Title     string `json:"title"`
			URL       string `json:"url"`
			Body      string `json:"body"`
			Number    int    `json:"number"`
			CreatedAt string `json:"createdAt"`
			ClosedAt  string `json:"closedAt"`
			Closed    bool   `json:"closed"`
		}{
			{Title: "Test Discussion", URL: "http://example.com/discussion/1", Body: "Discussion body.", Number: 1, Closed: false, CreatedAt: "2025-01-20T00:00:00Z"},
		}
		resp.Search.PageInfo.HasNextPage = false
		return nil
	}

	stdout, stderr := captureOutput(func() {
		handleDiscussionsCommand(testArgs, mockGQLClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedOutput := fmt.Sprintf("%s\n%s #%d\n%s\n%s\n%s\n", startOfDiscussion, "Test Discussion", 1, "Discussion body.", endOfDiscussion, entryDelimiter)

	if stdout != expectedOutput {
		t.Errorf("Expected stdout to be:\n%s\nGot:\n%s", expectedOutput, stdout)
	}
}

func TestFetchDiscussions(t *testing.T) {
	resetFlags()
	mockGQLClient := &MockGraphQLClient{}

	mockGQLClient.DoFunc = func(query string, variables map[string]interface{}, response interface{}) error {
		resp := response.(*DiscussionSearchResponse)
		resp.Search.Nodes = []struct {
			Title     string `json:"title"`
			URL       string `json:"url"`
			Body      string `json:"body"`
			Number    int    `json:"number"`
			CreatedAt string `json:"createdAt"`
			ClosedAt  string `json:"closedAt"`
			Closed    bool   `json:"closed"`
		}{
			{Title: "Open Discussion", URL: "http://example.com/d/1", Number: 1, Closed: false, CreatedAt: "2025-01-20T00:00:00Z"},
			{Title: "Closed Discussion", URL: "http://example.com/d/2", Number: 2, Closed: true, CreatedAt: "2025-01-15T00:00:00Z", ClosedAt: "2025-01-18T00:00:00Z"},
		}
		resp.Search.PageInfo.HasNextPage = false
		return nil
	}

	items, err := fetchDiscussions(mockGQLClient, "testuser", "github", "2025-01-01")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	if items[0].State != "open" {
		t.Errorf("Expected first item state 'open', got '%s'", items[0].State)
	}
	if items[1].State != "closed" {
		t.Errorf("Expected second item state 'closed', got '%s'", items[1].State)
	}
	if items[0].Title != "Open Discussion" {
		t.Errorf("Expected first item title 'Open Discussion', got '%s'", items[0].Title)
	}
}

// Add more tests for edge cases, error handling, pagination in fetchAllResults, etc.
