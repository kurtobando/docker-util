package models

// LogEntry matches the structure for displaying logs in the template
// This preserves the exact same structure from the original main.go
type LogEntry struct {
	ContainerID   string
	ContainerName string
	Timestamp     string
	StreamType    string
	LogMessage    string
}

// SearchParams holds parameters for log search and filtering
type SearchParams struct {
	SearchQuery        string
	SelectedContainers []string
	CurrentPage        int
	Offset             int
}

// PaginationInfo holds pagination state information
type PaginationInfo struct {
	CurrentPage int
	HasNext     bool
	HasPrev     bool
	NextPageURL string
	PrevPageURL string
}

// Container represents a Docker container
type Container struct {
	ID    string
	Name  string
	Image string
}

// TemplateData holds all data passed to the HTML template
type TemplateData struct {
	Logs                []LogEntry
	DBPath              string
	SearchQuery         string
	ResultCount         int
	HasSearch           bool
	CurrentPage         int
	HasNext             bool
	HasPrev             bool
	NextPageURL         string
	PrevPageURL         string
	AvailableContainers []string
	SelectedContainers  []string
	HasContainerFilter  bool
}
