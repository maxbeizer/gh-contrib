package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestHandleGraphCommand_Basic(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"graph", testLogin}

	// Set a fixed date for testing
	fixedNow, _ := time.Parse(dateFormat, "2025-05-15")
	fixedOneMonthAgo := fixedNow.AddDate(0, -1, 0)
	fixedSince := fixedOneMonthAgo.Format(dateFormat) // "2025-04-15"
	since = fixedSince

	// Create test dates for items
	week1Date := fixedOneMonthAgo.AddDate(0, 0, 3)  // 3 days after since date (Week 1)
	week2Date := fixedOneMonthAgo.AddDate(0, 0, 10) // 10 days after since date (Week 2)
	week3Date := fixedOneMonthAgo.AddDate(0, 0, 17) // 17 days after since date (Week 3)

	mockClient.GetFunc = func(path string, response interface{}) error {
		if strings.Contains(path, "is%3Apr") {
			// PR response
			resp := GitHubResponse{
				TotalCount: 3,
				Items: []GitHubItem{
					{
						Number:    101,
						Title:     "Closed PR Week 1",
						HTMLURL:   "http://example.com/pr/101",
						State:     "closed",
						CreatedAt: week1Date.AddDate(0, 0, -1).Format(time.RFC3339),
						ClosedAt:  week1Date.Format(time.RFC3339),
					},
					{
						Number:    102,
						Title:     "Open PR Week 2",
						HTMLURL:   "http://example.com/pr/102",
						State:     "open",
						CreatedAt: week2Date.Format(time.RFC3339),
						ClosedAt:  "",
					},
					{
						Number:    103,
						Title:     "Closed PR Week 3",
						HTMLURL:   "http://example.com/pr/103",
						State:     "closed",
						CreatedAt: week3Date.AddDate(0, 0, -2).Format(time.RFC3339),
						ClosedAt:  week3Date.Format(time.RFC3339),
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		} else if strings.Contains(path, "is%3Aissue") {
			// Issue response
			resp := GitHubResponse{
				TotalCount: 3,
				Items: []GitHubItem{
					{
						Number:    201,
						Title:     "Closed Issue Week 1",
						HTMLURL:   "http://example.com/issue/201",
						State:     "closed",
						CreatedAt: week1Date.AddDate(0, 0, -1).Format(time.RFC3339),
						ClosedAt:  week1Date.Format(time.RFC3339),
					},
					{
						Number:    202,
						Title:     "Open Issue Week 2",
						HTMLURL:   "http://example.com/issue/202",
						State:     "open",
						CreatedAt: week2Date.Format(time.RFC3339),
						ClosedAt:  "",
					},
					{
						Number:    203,
						Title:     "Closed Issue Week 3",
						HTMLURL:   "http://example.com/issue/203",
						State:     "closed",
						CreatedAt: week3Date.AddDate(0, 0, -2).Format(time.RFC3339),
						ClosedAt:  week3Date.Format(time.RFC3339),
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handleGraphCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	// Check for presence of key output elements
	expectedOutputs := []string{
		"Week  1",
		"Week  2",
		"Week  3",
		"•", // Closed PR symbol
		"○", // Open PR symbol
		"■", // Closed Issue symbol
		"□", // Open Issue symbol
		"Legend:",
		"• = Closed PR  ○ = Open PR  ■ = Closed Issue  □ = Open Issue",
		"Total Contributions: 6",
		"PRs: 3 total (2 closed, 1 open)",
		"Issues: 3 total (2 closed, 1 open)",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', but it doesn't.\nOutput:\n%s", expected, stdout)
		}
	}

	// Check API calls
	if len(mockClient.GetCalls) != 2 {
		t.Errorf("Expected 2 API calls (PRs + Issues), got %d", len(mockClient.GetCalls))
	}
}

func TestHandleGraphCommand_NoPRs(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"graph", testLogin}

	// Return empty PR list but some issues
	mockClient.GetFunc = func(path string, response interface{}) error {
		if strings.Contains(path, "is%3Apr") {
			// Empty PR response
			resp := GitHubResponse{
				TotalCount: 0,
				Items:      []GitHubItem{},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		} else if strings.Contains(path, "is%3Aissue") {
			// Issue response
			resp := GitHubResponse{
				TotalCount: 2,
				Items: []GitHubItem{
					{
						Number:    201,
						Title:     "Test Issue 1",
						HTMLURL:   "http://example.com/issue/201",
						State:     "closed",
						CreatedAt: time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
						ClosedAt:  time.Now().AddDate(0, 0, -3).Format(time.RFC3339),
					},
					{
						Number:    202,
						Title:     "Test Issue 2",
						HTMLURL:   "http://example.com/issue/202",
						State:     "open",
						CreatedAt: time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handleGraphCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	// Check for presence of key output elements
	expectedOutputs := []string{
		"■", // Closed Issue symbol
		"□", // Open Issue symbol
		"PRs: 0 total (0 closed, 0 open)",
		"Issues: 2 total (1 closed, 1 open)",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', but it doesn't.\nOutput:\n%s", expected, stdout)
		}
	}

	// Should not contain PR symbols
	unexpectedOutputs := []string{
		"•", // Closed PR symbol
		"○", // Open PR symbol
	}

	for _, unexpected := range unexpectedOutputs {
		if strings.Contains(stdout, unexpected) {
			t.Errorf("Output should not contain '%s', but it does.\nOutput:\n%s", unexpected, stdout)
		}
	}
}

func TestHandleGraphCommand_NoIssues(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"graph", testLogin}

	// Return some PRs but empty issue list
	mockClient.GetFunc = func(path string, response interface{}) error {
		if strings.Contains(path, "is%3Apr") {
			resp := GitHubResponse{
				TotalCount: 2,
				Items: []GitHubItem{
					{
						Number:    101,
						Title:     "Test PR 1",
						HTMLURL:   "http://example.com/pr/101",
						State:     "closed",
						CreatedAt: time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
						ClosedAt:  time.Now().AddDate(0, 0, -3).Format(time.RFC3339),
					},
					{
						Number:    102,
						Title:     "Test PR 2",
						HTMLURL:   "http://example.com/pr/102",
						State:     "open",
						CreatedAt: time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		} else if strings.Contains(path, "is%3Aissue") {
			// Empty Issue response
			resp := GitHubResponse{
				TotalCount: 0,
				Items:      []GitHubItem{},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handleGraphCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	// Check for presence of key output elements
	expectedOutputs := []string{
		"•", // Closed PR symbol
		"○", // Open PR symbol
		"PRs: 2 total (1 closed, 1 open)",
		"Issues: 0 total (0 closed, 0 open)",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', but it doesn't.\nOutput:\n%s", expected, stdout)
		}
	}

	// Should not contain issue symbols
	unexpectedOutputs := []string{
		"■", // Closed Issue symbol
		"□", // Open Issue symbol
	}

	for _, unexpected := range unexpectedOutputs {
		if strings.Contains(stdout, unexpected) {
			t.Errorf("Output should not contain '%s', but it does.\nOutput:\n%s", unexpected, stdout)
		}
	}
}

func TestHandleGraphCommand_NoResults(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"graph", testLogin}

	// Return empty results for both PRs and Issues
	mockClient.GetFunc = func(path string, response interface{}) error {
		// Empty response for all requests
		resp := GitHubResponse{
			TotalCount: 0,
			Items:      []GitHubItem{},
		}
		data, _ := json.Marshal(resp)
		return json.Unmarshal(data, response)
	}

	stdout, stderr := captureOutput(func() {
		handleGraphCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	expectedOutput := "No contributions found for user"
	if !strings.Contains(stdout, expectedOutput) {
		t.Errorf("Expected output to contain '%s', but it doesn't.\nOutput:\n%s", expectedOutput, stdout)
	}
}

func TestHandleGraphCommand_APIError(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"graph", testLogin}

	// Simulate API error
	mockClient.GetFunc = func(path string, response interface{}) error {
		return fmt.Errorf("simulated API error")
	}

	_, stderr := captureOutput(func() {
		handleGraphCommand(testArgs, mockClient)
	})

	expectedError := "Error fetching pull requests for graph:"
	if !strings.Contains(stderr, expectedError) {
		t.Errorf("Expected stderr to contain '%s', got: %s", expectedError, stderr)
	}
}

func TestHandleGraphCommand_DateHandling(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"graph", testLogin}

	// Set a fixed date for testing
	fixedNow, _ := time.Parse(dateFormat, "2025-05-15")
	fixedOneMonthAgo := fixedNow.AddDate(0, -1, 0)
	fixedSince := fixedOneMonthAgo.Format(dateFormat) // "2025-04-15"
	since = fixedSince

	// Test with a mix of valid and invalid dates
	mockClient.GetFunc = func(path string, response interface{}) error {
		if strings.Contains(path, "is%3Apr") {
			resp := GitHubResponse{
				TotalCount: 3,
				Items: []GitHubItem{
					{
						// Valid closed_at date
						Number:    101,
						Title:     "PR with valid closed date",
						HTMLURL:   "http://example.com/pr/101",
						State:     "closed",
						CreatedAt: "2025-04-20T12:00:00Z",
						ClosedAt:  "2025-04-22T12:00:00Z",
					},
					{
						// Missing closed_at date (open PR)
						Number:    102,
						Title:     "PR with no closed date",
						HTMLURL:   "http://example.com/pr/102",
						State:     "open",
						CreatedAt: "2025-04-25T12:00:00Z",
					},
					{
						// Invalid date format (should fall back to created_at)
						Number:    103,
						Title:     "PR with invalid date",
						HTMLURL:   "http://example.com/pr/103",
						State:     "closed",
						CreatedAt: "2025-05-01T12:00:00Z",
						ClosedAt:  "invalid-date",
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		} else if strings.Contains(path, "is%3Aissue") {
			resp := GitHubResponse{
				TotalCount: 1,
				Items: []GitHubItem{
					{
						Number:    201,
						Title:     "Test Issue",
						HTMLURL:   "http://example.com/issue/201",
						State:     "closed",
						CreatedAt: "2025-05-05T12:00:00Z",
						ClosedAt:  "2025-05-07T12:00:00Z",
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handleGraphCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	// Verify weeks in output
	expectedWeeks := []string{
		"Week  1", // Apr 15-21
		"Week  2", // Apr 22-28
		"Week  3", // Apr 29-May 5
		"Week  4", // May 6-12
	}

	for _, week := range expectedWeeks {
		if !strings.Contains(stdout, week) {
			t.Errorf("Expected output to contain week '%s', but it doesn't.\nOutput:\n%s", week, stdout)
		}
	}
}

func TestHandleGraphCommand_WebURL(t *testing.T) {
	resetFlags()
	mockClient := &MockGitHubClient{}
	testLogin := "testuser"
	testArgs := []string{"graph", testLogin}

	// Mock organization function to return "testorg"
	originalOrgConfigFunc := orgConfigFunc
	orgConfigFunc = func() (string, error) {
		return "testorg", nil
	}
	defer func() {
		orgConfigFunc = originalOrgConfigFunc
	}()

	// Mock API responses with minimal data
	mockClient.GetFunc = func(path string, response interface{}) error {
		if strings.Contains(path, "is%3Apr") {
			// PR response
			resp := GitHubResponse{
				TotalCount: 1,
				Items: []GitHubItem{
					{
						Number:    101,
						Title:     "Test PR",
						HTMLURL:   "http://example.com/pr/101",
						State:     "open",
						CreatedAt: time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		} else if strings.Contains(path, "is%3Aissue") {
			// Issue response
			resp := GitHubResponse{
				TotalCount: 1,
				Items: []GitHubItem{
					{
						Number:    201,
						Title:     "Test Issue",
						HTMLURL:   "http://example.com/issue/201",
						State:     "open",
						CreatedAt: time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, response)
		}
		return fmt.Errorf("unexpected API call: %s", path)
	}

	stdout, stderr := captureOutput(func() {
		handleGraphCommand(testArgs, mockClient)
	})

	if stderr != "" {
		t.Errorf("Expected no stderr, got: %s", stderr)
	}

	// Check that the web URL is displayed
	if !strings.Contains(stdout, "View issues in GitHub:") {
		t.Errorf("Expected output to contain web URL introduction, but it doesn't.\nOutput:\n%s", stdout)
	}

	if !strings.Contains(stdout, "https://github.com/issues?q=") {
		t.Errorf("Expected output to contain GitHub issues URL, but it doesn't.\nOutput:\n%s", stdout)
	}

	if !strings.Contains(stdout, "is%3Aissue") {
		t.Errorf("Expected URL to contain URL-encoded 'is:issue', but it doesn't.\nOutput:\n%s", stdout)
	}

	if !strings.Contains(stdout, "org%3Atestorg") {
		t.Errorf("Expected URL to contain URL-encoded 'org:testorg', but it doesn't.\nOutput:\n%s", stdout)
	}

	if !strings.Contains(stdout, "author%3Atestuser") {
		t.Errorf("Expected URL to contain URL-encoded 'author:testuser', but it doesn't.\nOutput:\n%s", stdout)
	}
}
