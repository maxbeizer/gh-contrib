package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"os/user"

	"github.com/cli/go-gh/v2/pkg/api"
	"gopkg.in/yaml.v2"
)

// --- Interfaces for Dependency Injection ---

// GitHubClient defines the methods needed to interact with the GitHub API.
type GitHubClient interface {
	Get(path string, response interface{}) error
}

// TokenFetcher defines the method needed to fetch an authentication token.
type TokenFetcher interface {
	FetchToken() (string, error)
}

// Summarizer defines the method needed to summarize text.
type Summarizer interface {
	Summarize(text string) (string, error)
}

// --- Concrete Implementations ---

// DefaultGitHubClient is the default implementation using go-gh.
type DefaultGitHubClient struct {
	client *api.RESTClient
}

func NewDefaultGitHubClient() (*DefaultGitHubClient, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("error creating default GitHub API client: %w", err)
	}
	return &DefaultGitHubClient{client: client}, nil
}

func (c *DefaultGitHubClient) Get(path string, response interface{}) error {
	return c.client.Get(path, response)
}

// GhCliTokenFetcher fetches the token using the 'gh' CLI.
type GhCliTokenFetcher struct{}

func (tf *GhCliTokenFetcher) FetchToken() (string, error) {
	cmd := exec.Command("gh", "auth", "status", "--show-token")
	tokenOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running gh auth status: %w", err)
	}

	lines := strings.Split(string(tokenOutput), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, tokenPrefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, tokenPrefix)), nil
		}
	}
	return "", fmt.Errorf("github token not found in auth status output")
}

// AzureAISummarizer uses the Azure AI endpoint for summarization.
type AzureAISummarizer struct {
	httpClient   *http.Client
	tokenFetcher TokenFetcher
	endpoint     string
	model        string
}

func NewAzureAISummarizer(httpClient *http.Client, tokenFetcher TokenFetcher) *AzureAISummarizer {
	return &AzureAISummarizer{
		httpClient:   httpClient,
		tokenFetcher: tokenFetcher,
		endpoint:     aiEndpoint,
		model:        getEffectiveModel(), // Use the effective model
	}
}

func (s *AzureAISummarizer) Summarize(text string) (string, error) {
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": fmt.Sprintf(userPrompt, text)},
		},
		"temperature": 1.0,
		"top_p":       1.0,
		"max_tokens":  1000,
		"model":       s.model,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error creating JSON payload: %w", err)
	}

	githubToken, err := s.tokenFetcher.FetchToken()
	if err != nil {
		return "", fmt.Errorf("error retrieving GitHub token: %w", err)
	}

	req, err := http.NewRequest("POST", s.endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var aiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(responseBody, &aiResponse); err != nil {
		// Optionally log the raw response body here for debugging
		// fmt.Printf("Raw AI response: %s\n", string(responseBody))
		return "", fmt.Errorf("error parsing AI response JSON: %w", err)
	}

	if len(aiResponse.Choices) > 0 && aiResponse.Choices[0].Message.Content != "" {
		return aiResponse.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no summary content available in the AI response")
}

const (
	defaultOrg     = "github"
	dateFormat     = "2006-01-02"
	defaultModel   = "gpt-4o"
	aiEndpoint     = "https://models.inference.ai.azure.com/chat/completions"
	tokenPrefix    = "  - Token:"
	entryDelimiter = "---END-OF-ENTRY---"
	startOfEntry   = "---START-OF-ENTRY---"
	startOfPR      = "---START-OF-PR---"
	startOfIssue   = "---START-OF-ISSUE---"
	endOfPR        = "---END-OF-PR---"
	endOfIssue     = "---END-OF-ISSUE---"

	systemPrompt = `You are an expert engineering manager assistant designed to
	summarize the bodies of GitHub issues and pull requests. Your goal is to
	extract key details, provide concise summaries, and ignore irrelevant
	sections or headers such as 'Mitigation and Rollback Strategies', 'Testing',
	'Deployment Plan', and 'Approval Responsibility'. Ensure the summaries are
	actionable and easy to understand. Your responses should be in Markdown
	format without wrapping Markdown in a code fence and geared for a technical
	audience with an emphasis on readability.

  Format entries as follows:
  ## <descriptive title>
  <summary of the entry>
  ### Links
  - [Link to Artifact 1](<URL>)
  - [Link to Artifact 2](<URL>)

  <br /><br />

  For each distinct entry, provide a summary that captures the essence of the
  content, while ensuring that any links to artifacts are included. Do not
  include any headers or irrelevant sections in your summaries.`

	userPrompt = `Summarize the following text while ignoring sections with
	headers like (e.g., 'Mitigation and Rollback Strategies', 'Testing',
	'Deployment Plan', 'Approval Responsibility'), include links to all
	artifacts: %s`
)

// Structs for API responses
type GitHubItem struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	HTMLURL    string `json:"html_url"`
	State      string `json:"state"`
	Body       string `json:"body,omitempty"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

type GitHubResponse struct {
	TotalCount int          `json:"total_count"`
	Items      []GitHubItem `json:"items"`
}

// Global variables
var (
	debug     bool
	since     string
	bodyOnly  bool
	orgFlag   string
	modelFlag string // Global variable to store the value of the --model flag
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	defaultSince := time.Now().AddDate(0, 0, -30).Format(dateFormat)
	flag.StringVar(&since, "since", defaultSince, "Filter results created since the specified date (e.g., 2025-04-11)")
	flag.BoolVar(&bodyOnly, "body-only", false, "Fetch and print only the body of the pull requests")
	flag.StringVar(&orgFlag, "org", "", "Override the configured organization")
	flag.StringVar(&modelFlag, "model", "", "Override the configured or default model")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if debug {
		fmt.Println("Debug mode enabled")
		fmt.Printf("Arguments: %v\n", args)
		fmt.Printf("Using AI model: %s\n", getEffectiveModel())
	}

	ghClient, err := NewDefaultGitHubClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing GitHub client: %v\n", err)
		os.Exit(1)
	}

	tokenFetcher := &GhCliTokenFetcher{}
	httpClient := &http.Client{}
	summarizer := NewAzureAISummarizer(httpClient, tokenFetcher)

	if len(args) == 0 {
		printHelp(ghClient)
		return
	}

	cmd := args[0]
	switch cmd {
	case "pulls":
		handlePullsCommand(args, ghClient)
	case "issues":
		handleIssuesCommand(args, ghClient)
	case "all":
		handleAllCommand(args, ghClient)
	case "summarize":
		handleSummarizeCommand(args, summarizer)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printHelp(ghClient)
	}
}

func handlePullsCommand(args []string, client GitHubClient) {
	if len(args) < 2 {
		fmt.Println("Error: login argument is required")
		fmt.Println("Usage: gh-contrib pulls <login>")
		return
	}
	login := args[1]

	org, err := orgConfigFunc()
	if err != nil {
		org = defaultOrg
	}

	query := buildQuery("is:pr", login)
	searchURL := fmt.Sprintf("search/issues?q=%s", query)

	if debug {
		fmt.Printf("Calling GitHub API with URL: %s\n", searchURL)
	}

	responseItems, err := fetchAllResults(client, searchURL)
	if err != nil {
		fmt.Println("Error fetching pull requests:", err)
		return
	}

	if len(responseItems) == 0 {
		fmt.Printf("No pull requests found for user '%s' in the '%s' organization.\n", login, org)
		return
	}

	if bodyOnly {
		printBodies(responseItems, startOfPR, endOfPR)
		return
	}

	printPullRequestsAsCSV(responseItems)
}

func handleIssuesCommand(args []string, client GitHubClient) {
	if len(args) < 2 {
		fmt.Println("Error: login argument is required")
		fmt.Println("Usage: gh-contrib issues <login>")
		return
	}
	login := args[1]

	org, err := orgConfigFunc()
	if err != nil {
		org = defaultOrg
	}

	query := buildQuery("is:issue", login)
	searchURL := fmt.Sprintf("search/issues?q=%s", query)

	if debug {
		fmt.Printf("Calling GitHub API with URL: %s\n", searchURL)
	}

	responseItems, err := fetchAllResults(client, searchURL)
	if err != nil {
		fmt.Println("Error fetching issues:", err)
		return
	}

	if len(responseItems) == 0 {
		fmt.Printf("No issues found for user '%s' in the '%s' organization.\n", login, org)
		return
	}

	if bodyOnly {
		printBodies(responseItems, startOfIssue, endOfIssue)
		return
	}

	printIssuesAsCSV(responseItems)
}

func handleAllCommand(args []string, client GitHubClient) {
	if len(args) < 2 {
		fmt.Println("Error: login argument is required")
		fmt.Println("Usage: gh-contrib all <login>")
		return
	}
	login := args[1]

	prQuery := buildQuery("is:pr", login)
	prSearchURL := fmt.Sprintf("search/issues?q=%s", prQuery)
	if debug {
		fmt.Printf("Calling GitHub API for PRs with URL: %s\n", prSearchURL)
	}

	prItems, err := fetchAllResults(client, prSearchURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching pull requests: %v\n", err)
		return
	}

	issueQuery := buildQuery("is:issue", login)
	issueSearchURL := fmt.Sprintf("search/issues?q=%s", issueQuery)
	if debug {
		fmt.Printf("Calling GitHub API for issues with URL: %s\n", issueSearchURL)
	}

	issueItems, err := fetchAllResults(client, issueSearchURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching issues: %v\n", err)
		return
	}

	if bodyOnly {

		printBodies(prItems, startOfPR, endOfPR)
		printBodies(issueItems, startOfIssue, endOfIssue)
		return
	}

	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write the header row
	writer.Write([]string{"Type", "URL", "Title", "State"})

	// Write pull requests
	for _, pr := range prItems {
		writer.Write([]string{
			"Pull Request",
			pr.HTMLURL + " ",
			pr.Title,
			pr.State,
		})
	}

	// Write issues
	for _, issue := range issueItems {
		writer.Write([]string{
			"Issue",
			issue.HTMLURL + " ",
			issue.Title,
			issue.State,
		})
	}
}

func handleSummarizeCommand(args []string, summarizer Summarizer) {
	var input string
	if len(args) > 1 {
		input = args[1]
	} else {
		stdinInput, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			return
		}
		input = string(stdinInput)
	}

	entries := strings.Split(input, entryDelimiter)

	for _, entry := range entries {
		entry = strings.TrimSpace(entry) // Trim any extra whitespace
		if entry == "" {
			continue
		}

		summary, err := summarizer.Summarize(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error summarizing entry: %v\n", err)
			continue // Continue to the next entry on error
		}

		fmt.Println(summary)
	}
}

var orgConfigFunc = getOrgFromConfig // Default to the actual implementation

// Function to read the organization from the GitHub CLI config file
func getOrgFromConfig() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error getting current user: %w", err)
	}

	configPath := filepath.Join(usr.HomeDir, ".config", "gh", "config.yml")
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("error reading config file: %w", err)
	}

	var config struct {
		Extensions map[string]struct {
			Org string `yaml:"org"`
		} `yaml:"extensions"`
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return "", fmt.Errorf("error parsing config file: %w", err)
	}

	// Assuming the extension name is "gh-contrib"
	if extConfig, ok := config.Extensions["gh-contrib"]; ok {
		return extConfig.Org, nil
	}

	return "", fmt.Errorf("organization not found in config file under extensions")
}

func getEffectiveOrg() string {
	if orgFlag != "" {
		return orgFlag // Use the --org flag if provided
	}

	org, err := orgConfigFunc()
	if err != nil {
		return defaultOrg // Default to 'github' if not found
	}

	return org
}

func getEffectiveModel() string {
	if modelFlag != "" {
		return modelFlag // Use the --model flag if provided
	}
	return modelConfigFunc() // Use the configured or default model
}

func buildQuery(itemType, login string) string {
	org := getEffectiveOrg() // Use the effective organization
	query := fmt.Sprintf("%s org:%s author:%s sort:created-desc", itemType, org, login)
	if since != "" {
		query += fmt.Sprintf(" created:>%s", since)
		query = url.QueryEscape(query)
	}
	return query
}

func fetchAllResults(client GitHubClient, searchURL string) ([]GitHubItem, error) {
	var allItems []GitHubItem
	page := 1
	const maxPages = 10 // Safety break to prevent infinite loops in case of API issues

	for page <= maxPages {
		separator := "&"
		if !strings.Contains(searchURL, "?") {
			separator = "?"
		}
		paginatedURL := fmt.Sprintf("%s%spage=%d&per_page=100", searchURL, separator, page)

		if debug {
			fmt.Printf("Fetching page %d: %s\n", page, paginatedURL)
		}

		response := GitHubResponse{}

		err := client.Get(paginatedURL, &response)
		if err != nil {
			return nil, fmt.Errorf("error fetching page %d from %s: %w", page, paginatedURL, err)
		}

		if debug {
			fmt.Printf("Page %d: Found %d items (TotalCount: %d)\n", page, len(response.Items), response.TotalCount)
		}

		allItems = append(allItems, response.Items...)

		if len(response.Items) < 100 {
			break
		}

		page++
	}

	if page > maxPages {
		fmt.Fprintf(os.Stderr, "Warning: Reached maximum page limit (%d) for URL: %s\n", maxPages, searchURL)
	}

	return allItems, nil
}

func printUserInfo(client GitHubClient) {
	response := struct{ Login string }{}
	err := client.Get("user", &response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching user info: %v\n", err)
		return
	}
	fmt.Printf("running as %s\n", response.Login)
}

func printHelp(client GitHubClient) {
	fmt.Println("gh-contrib: A tool to better understand GitHub Issues and Pull Requests.")
	printUserInfo(client)
	fmt.Println("\nAvailable commands:")
	fmt.Println("  pulls <username>   - Get Pull Requests authored by <username> in the 'github' (or specified) org.")
	fmt.Println("  issues <username>  - Get Issues authored by <username> in the 'github' (or specified) org.")
	fmt.Println("  all <username>     - Get all Pull Requests and Issues by <username> in the 'github' (or specified) org.")
	fmt.Println("  summarize          - Summarize PR/Issue bodies from stdin or argument.")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
}

func printPullRequestsAsCSV(pullRequests []GitHubItem) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write the header row
	writer.Write([]string{"URL", "Title", "State"})

	// Write each pull request as a row
	for _, pr := range pullRequests {
		writer.Write([]string{
			pr.HTMLURL + " ", // Add a space after the URL intentionally to make terminal clicking easier
			pr.Title,
			pr.State,
		})
	}
}

func printIssuesAsCSV(issues []GitHubItem) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write the header row
	writer.Write([]string{"URL", "Title", "State"})

	// Write each issue as a row
	for _, issue := range issues {
		writer.Write([]string{
			issue.HTMLURL + " ", // Add a space after the URL intentionally to make terminal clicking easier
			issue.Title,
			issue.State,
		})
	}
}

func printBodies(items []GitHubItem, startMarker, endMarker string) {
	for _, item := range items {
		// Use the correct delimiter constant for consistency between entries
		fmt.Printf("%s\n%s #%d\n%s\n%s\n%s\n", startMarker, item.Title, item.Number, item.Body, endMarker, entryDelimiter)
	}
}

var modelConfigFunc = getModelFromConfig // Default to the actual implementation

func getModelFromConfig() string {
	configPath := os.Getenv("GH_CONFIG_PATH")
	if configPath == "" {
		usr, err := user.Current()
		if err != nil {
			return defaultModel // Default to 'gpt-4o' if user info is unavailable
		}
		configPath = filepath.Join(usr.HomeDir, ".config", "gh", "config.yml")
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return defaultModel // Default to 'gpt-4o' if config file is missing
	}

	var config struct {
		Extensions map[string]struct {
			Model string `yaml:"model"`
		} `yaml:"extensions"`
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return defaultModel // Default to 'gpt-4o' if parsing fails
	}

	if extConfig, ok := config.Extensions["gh-contrib"]; ok && extConfig.Model != "" {
		return extConfig.Model
	}

	return defaultModel // Default to 'gpt-4o' if model is not configured
}
