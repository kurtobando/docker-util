package web

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bw_util/internal/database"
	"bw_util/internal/models"
)

// Handler manages HTTP request handling
type Handler struct {
	repository      database.LogRepository
	templateManager *TemplateManager
	dbPath          string
}

// NewHandler creates a new HTTP handler
func NewHandler(repository database.LogRepository, templateManager *TemplateManager, dbPath string) *Handler {
	return &Handler{
		repository:      repository,
		templateManager: templateManager,
		dbPath:          dbPath,
	}
}

// ServeLogsPage handles the main logs page request
// This preserves the exact same logic from the original serveLogsPage function
func (h *Handler) ServeLogsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get search query parameter
	searchQuery := strings.TrimSpace(r.URL.Query().Get("search"))

	// Get container filter parameter (comma-separated)
	containersParam := strings.TrimSpace(r.URL.Query().Get("containers"))
	var selectedContainers []string
	if containersParam != "" {
		// Split by comma and trim spaces
		for _, container := range strings.Split(containersParam, ",") {
			if trimmed := strings.TrimSpace(container); trimmed != "" {
				selectedContainers = append(selectedContainers, trimmed)
			}
		}
	}

	// Get page parameter (default to 1)
	pageStr := r.URL.Query().Get("page")
	currentPage := 1
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage >= 1 {
			currentPage = parsedPage
		}
	}

	// Safety limit: max 1000 pages (100K records)
	if currentPage > 1000 {
		currentPage = 1000
	}

	// Calculate OFFSET
	offset := (currentPage - 1) * 100

	// Get available containers dynamically
	availableContainers, err := h.repository.GetContainers(r.Context())
	if err != nil {
		log.Printf("Error fetching available containers: %v", err)
		http.Error(w, "Failed to load container list", http.StatusInternalServerError)
		return
	}

	// Search for logs using the repository
	searchParams := &models.SearchParams{
		SearchQuery:        searchQuery,
		SelectedContainers: selectedContainers,
		CurrentPage:        currentPage,
		Offset:             offset,
	}

	logsToDisplay, hasNext, err := h.repository.Search(r.Context(), searchParams)
	if err != nil {
		log.Printf("Error querying logs from DB: %v", err)
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}

	// Determine pagination state
	hasPrev := currentPage > 1

	// Build pagination URLs
	nextPageURL := ""
	prevPageURL := ""

	if hasNext {
		nextPageURL = h.buildPaginationURL(currentPage+1, searchQuery, selectedContainers)
	}

	if hasPrev {
		prevPageURL = h.buildPaginationURL(currentPage-1, searchQuery, selectedContainers)
	}

	// Use a map for data to easily add page title or other elements if needed
	data := &models.TemplateData{
		Logs:                logsToDisplay,
		DBPath:              h.dbPath,
		SearchQuery:         searchQuery,
		ResultCount:         len(logsToDisplay),
		HasSearch:           searchQuery != "",
		CurrentPage:         currentPage,
		HasNext:             hasNext,
		HasPrev:             hasPrev,
		NextPageURL:         nextPageURL,
		PrevPageURL:         prevPageURL,
		AvailableContainers: availableContainers,
		SelectedContainers:  selectedContainers,
		HasContainerFilter:  len(selectedContainers) > 0,
	}

	err = h.templateManager.GetTemplate().ExecuteTemplate(w, "logs.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// buildPaginationURL constructs pagination URLs while preserving search parameters
// This preserves the exact same logic from the original buildPaginationURL function
func (h *Handler) buildPaginationURL(page int, searchQuery string, containers []string) string {
	urlStr := "/?page=" + strconv.Itoa(page)
	if searchQuery != "" {
		urlStr += "&search=" + url.QueryEscape(searchQuery)
	}
	if len(containers) > 0 {
		urlStr += "&containers=" + url.QueryEscape(strings.Join(containers, ","))
	}
	return urlStr
}
