package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// AINews represents an AI-related news article
type AINews struct {
	Title     string    `json:"title"`
	Link      string    `json:"link"`
	Source    string    `json:"source"`
	Summary   string    `json:"summary"`
	Category  string    `json:"category"`
	ScrapedAt time.Time `json:"scraped_at"`
	IsAITool  bool      `json:"is_ai_tool"`
}

// Newsletter represents the final newsletter structure
type Newsletter struct {
	GeneratedAt time.Time `json:"generated_at"`
	AITools     []AINews  `json:"ai_tools"`
	Research    []AINews  `json:"research"`
	Funding     []AINews  `json:"funding"`
	General     []AINews  `json:"general"`
	TotalCount  int       `json:"total_count"`
	DebugInfo   []string  `json:"debug_info"`
}

// AIScraper handles scraping AI news from multiple sources
type AIScraper struct {
	client       *http.Client
	aiKeywords   []string
	toolKeywords []string
	debugLogs    []string
}

// NewAIScraper creates a new AI-focused scraper
func NewAIScraper() *AIScraper {
	return &AIScraper{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		aiKeywords: []string{
			"AI", "artificial intelligence", "machine learning", "ML", "deep learning",
			"neural network", "GPT", "LLM", "ChatGPT", "OpenAI", "Claude", "Gemini",
			"automation", "algorithm", "computer vision", "NLP", "natural language",
		},
		toolKeywords: []string{
			"tool", "app", "platform", "software", "API", "service", "launch",
			"release", "beta", "startup", "product", "features", "integration",
		},
	}
}

func (s *AIScraper) addDebugLog(msg string) {
	s.debugLogs = append(s.debugLogs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	fmt.Println("🐛", msg)
}

// ScrapeHackerNews scrapes AI-related stories from Hacker News
func (s *AIScraper) ScrapeHackerNews() ([]AINews, error) {
	s.addDebugLog("Starting Hacker News scraping...")

	resp, err := s.client.Get("https://news.ycombinator.com")
	if err != nil {
		s.addDebugLog(fmt.Sprintf("❌ Failed to fetch Hacker News: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	s.addDebugLog(fmt.Sprintf("✅ Got response from Hacker News (Status: %d)", resp.StatusCode))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var news []AINews
	totalStories := 0

	doc.Find(".athing").Each(func(i int, sel *goquery.Selection) {
		totalStories++

		titleLink := sel.Find(".storylink, .titleline > a")
		title := strings.TrimSpace(titleLink.Text())
		link, exists := titleLink.Attr("href")

		if !exists || title == "" {
			return
		}

		// Debug first 5 stories
		if i < 5 {
			s.addDebugLog(fmt.Sprintf("Story %d: %s", i+1, title))
		}

		// Check if title contains AI-related keywords
		if s.isAIRelated(title) {
			s.addDebugLog(fmt.Sprintf("🎯 Found AI story: %s", title))

			if strings.HasPrefix(link, "item?") {
				link = "https://news.ycombinator.com/" + link
			} else if strings.HasPrefix(link, "/") {
				link = "https://news.ycombinator.com" + link
			}

			category := s.categorizeNews(title)

			news = append(news, AINews{
				Title:     title,
				Link:      link,
				Source:    "Hacker News",
				Summary:   s.generateSummary(title),
				Category:  category,
				ScrapedAt: time.Now(),
				IsAITool:  s.isAITool(title),
			})
		}
	})

	s.addDebugLog(fmt.Sprintf("Checked %d stories, found %d AI-related", totalStories, len(news)))
	return news, nil
}

// ScrapeTestSite creates dummy AI news data for testing
func (s *AIScraper) ScrapeTestSite() ([]AINews, error) {
	s.addDebugLog("Creating test AI news data...")

	testNews := []AINews{
		{
			Title:     "New ChatGPT Plugin Revolutionizes Code Reviews",
			Link:      "https://example.com/chatgpt-plugin",
			Source:    "Test Source",
			Summary:   "Revolutionary AI tool helps developers streamline code review processes with intelligent suggestions.",
			Category:  "tool",
			ScrapedAt: time.Now(),
			IsAITool:  true,
		},
		{
			Title:     "Google DeepMind Announces Breakthrough in AI Reasoning",
			Link:      "https://example.com/deepmind-research",
			Source:    "Test Source",
			Summary:   "Latest research shows significant improvements in AI logical reasoning capabilities.",
			Category:  "research",
			ScrapedAt: time.Now(),
			IsAITool:  false,
		},
		{
			Title:     "AI Startup Raises $50M Series A for Enterprise Automation",
			Link:      "https://example.com/ai-funding",
			Source:    "Test Source",
			Summary:   "Promising AI company secures major funding for business process automation tools.",
			Category:  "funding",
			ScrapedAt: time.Now(),
			IsAITool:  false,
		},
		{
			Title:     "Microsoft Copilot Gets Major Update with New AI Features",
			Link:      "https://example.com/copilot-update",
			Source:    "Test Source",
			Summary:   "Enhanced productivity features and improved AI assistance across Microsoft Office suite.",
			Category:  "tool",
			ScrapedAt: time.Now(),
			IsAITool:  true,
		},
		{
			Title:     "Open Source LLM Achieves GPT-4 Level Performance",
			Link:      "https://example.com/open-llm",
			Source:    "Test Source",
			Summary:   "Community-driven AI model matches commercial alternatives in benchmark tests.",
			Category:  "research",
			ScrapedAt: time.Now(),
			IsAITool:  false,
		},
	}

	s.addDebugLog(fmt.Sprintf("Generated %d test AI news articles", len(testNews)))
	return testNews, nil
}

// isAIRelated checks if content is AI-related
func (s *AIScraper) isAIRelated(text string) bool {
	lowerText := strings.ToLower(text)
	for _, keyword := range s.aiKeywords {
		if strings.Contains(lowerText, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// isAITool determines if the news is about an AI tool
func (s *AIScraper) isAITool(text string) bool {
	lowerText := strings.ToLower(text)
	hasAI := s.isAIRelated(text)
	hasTool := false

	for _, keyword := range s.toolKeywords {
		if strings.Contains(lowerText, strings.ToLower(keyword)) {
			hasTool = true
			break
		}
	}

	return hasAI && hasTool
}

// categorizeNews categorizes news into different types
func (s *AIScraper) categorizeNews(text string) string {
	lowerText := strings.ToLower(text)

	if s.isAITool(text) {
		return "tool"
	}

	fundingKeywords := []string{"funding", "investment", "raised", "series", "valuation", "ipo"}
	for _, keyword := range fundingKeywords {
		if strings.Contains(lowerText, keyword) {
			return "funding"
		}
	}

	researchKeywords := []string{"research", "study", "paper", "breakthrough", "discovery", "model"}
	for _, keyword := range researchKeywords {
		if strings.Contains(lowerText, keyword) {
			return "research"
		}
	}

	return "general"
}

// generateSummary creates a simple summary
func (s *AIScraper) generateSummary(title string) string {
	if s.isAITool(title) {
		return "New AI tool that could streamline workflows and boost productivity."
	}
	return "Latest development in artificial intelligence and machine learning."
}

// ScrapeAllSources scrapes all sources concurrently
func (s *AIScraper) ScrapeAllSources() *Newsletter {
	s.addDebugLog("🚀 Starting AI News Scraping...")

	var wg sync.WaitGroup
	var hnNews, testNews []AINews
	var hnErr, testErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		hnNews, hnErr = s.ScrapeHackerNews()
	}()

	go func() {
		defer wg.Done()
		testNews, testErr = s.ScrapeTestSite()
	}()

	wg.Wait()

	newsAll := append(hnNews, testNews...)

	if hnErr != nil {
		s.addDebugLog(fmt.Sprintf("Hacker News scraping error: %v", hnErr))
	}
	if testErr != nil {
		s.addDebugLog(fmt.Sprintf("Test site scraping error: %v", testErr))
	}

	newsByCategory := map[string][]AINews{
		"tool":    {},
		"research": {},
		"funding":  {},
		"general":  {},
	}

	for _, n := range newsAll {
		newsByCategory[n.Category] = append(newsByCategory[n.Category], n)
	}

	return &Newsletter{
		GeneratedAt: time.Now(),
		AITools:     newsByCategory["tool"],
		Research:    newsByCategory["research"],
		Funding:     newsByCategory["funding"],
		General:     newsByCategory["general"],
		TotalCount:  len(newsAll),
		DebugInfo:   s.debugLogs,
	}
}

// SaveToFile saves newsletter data as JSON to file
func (n *Newsletter) SaveToFile() error {
	file, err := os.Create("newsletter.json")
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(n)
}

// Web server handlers

func homeHandler(w http.ResponseWriter, r *http.Request) {
	scraper := NewAIScraper()
	newsletter := scraper.ScrapeAllSources()

	err := newsletter.SaveToFile()
	if err != nil {
		http.Error(w, "Failed to save newsletter file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Failed to load template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, newsletter)
	if err != nil {
		http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	scraper := NewAIScraper()
	newsletter := scraper.ScrapeAllSources()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(newsletter)
	if err != nil {
		http.Error(w, "Failed to encode JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/news", apiHandler)

	fmt.Println("Starting AI News Scraper web server on http://localhost:8085")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
