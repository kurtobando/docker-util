# Product Requirements Document: docker-util Log Collector & Viewer

**Version:** 2.0
**Date:** May 2025

## 1. Introduction

`docker-util` is a command-line Go application designed to collect logs from all running Docker containers on a host machine. It streams these logs in real-time, stores them in a local SQLite database, and provides a comprehensive web interface for viewing, searching, filtering, and navigating through the collected logs. The tool is robust, with graceful shutdown capabilities, configurable options for database storage and web UI port, and advanced log management features.

## 2. Goals

-   To provide a reliable method for centralizing Docker container logs locally.
-   To offer an accessible web interface for easy viewing, searching, and filtering of collected logs.
-   To create a lightweight, self-contained utility that is easy to build and run.
-   To ensure data integrity through graceful shutdown procedures.
-   To provide efficient log navigation and filtering capabilities for large log datasets.
-   To support dynamic container discovery and filtering without manual configuration.

## 3. Functional Requirements

### 3.1. Log Collection

-   **FC1:** The application MUST be developed in Go (target version Go 1.24.0 or later).
-   **FC2:** The application MUST connect to the local Docker daemon to discover all currently *running* Docker containers.
-   **FC3:** For each running container, the application MUST stream both `stdout` and `stderr` logs in real-time.
-   **FC4:** The application MUST include timestamps with each log entry, preferably in RFC3339Nano UTC format, parsed from the Docker log stream or generated at the time of ingestion if parsing fails.

### 3.2. Log Storage

-   **FC5:** All streamed logs MUST be stored in an SQLite database.
-   **FC6:** The path and filename for the SQLite database MUST be configurable via a command-line flag (`--dbpath`).
    -   Default path: `./docker_logs.db`
-   **FC7:** The database MUST contain a table named `logs`.
-   **FC8:** The `logs` table schema MUST be:
    -   `id` (INTEGER, PRIMARY KEY, AUTOINCREMENT)
    -   `container_id` (TEXT NOT NULL) - Full Docker container ID.
    -   `container_name` (TEXT NOT NULL) - Primary name of the Docker container.
    -   `timestamp` (TEXT NOT NULL) - Log entry timestamp (RFC3339Nano UTC format).
    -   `stream_type` (TEXT NOT NULL) - e.g., "stdout" or "stderr".
    -   `log_message` (TEXT NOT NULL) - The actual log content.
-   **FC9:** The application MUST create the database and the `logs` table if they do not exist upon startup.
-   **FC10:** The application MUST create performance indexes:
    -   `idx_logs_message` on `log_message` column for search functionality.
    -   `idx_logs_container_name` on `container_name` column for filtering functionality.

### 3.3. Web User Interface (UI)

-   **FC11:** The application MUST start an HTTP web server to provide a UI for viewing logs.
-   **FC12:** The port for the web server MUST be configurable via a command-line flag (`--port`).
    -   Default port: 9123.
-   **FC13:** The web UI MUST be accessible at the root path (e.g., `http://localhost:9123/`).
-   **FC14:** The web UI MUST display a table of log entries fetched from the SQLite database.
-   **FC15:** The log table MUST display the following columns: Timestamp, Container Name, Container ID (shortened for display, e.g., first 12 characters), Stream Type, and the Log Message.
-   **FC16:** The web UI SHOULD display the latest 100 log entries by default, ordered with the newest logs first.
-   **FC17:** HTML templates for the web UI MUST be embedded within the compiled Go binary.

### 3.4. Search Functionality

-   **FC18:** The web UI MUST provide a search input field for filtering log messages.
-   **FC19:** Search functionality MUST be case-insensitive using SQL `COLLATE NOCASE`.
-   **FC20:** Search MUST filter based on `log_message` content using `LIKE` pattern matching.
-   **FC21:** Search queries MUST be preserved across pagination navigation.
-   **FC22:** The UI MUST display search state and result count information.
-   **FC23:** A "Clear" option MUST be available to reset search filters.

### 3.5. Pagination

-   **FC24:** The web UI MUST implement pagination with Next/Previous navigation.
-   **FC25:** Each page MUST display exactly 100 log entries.
-   **FC26:** Pagination MUST use OFFSET/LIMIT approach with a safety limit of 1000 pages (100K records).
-   **FC27:** Pagination MUST preserve search queries and container filters in navigation URLs.
-   **FC28:** The UI MUST indicate current page number and availability of additional pages.
-   **FC29:** Pagination controls MUST be visually disabled when not applicable (first/last page).

### 3.6. Container Filtering

-   **FC30:** The web UI MUST provide dynamic container filtering via checkbox selection.
-   **FC31:** Available containers MUST be discovered automatically from the database.
-   **FC32:** Container filtering MUST support multiple container selection simultaneously.
-   **FC33:** Container filters MUST be preserved across pagination and search operations.
-   **FC34:** The UI MUST display selected container information and filter state.
-   **FC35:** Container filtering MUST work in combination with search functionality.
-   **FC36:** Container filter state MUST be maintained via URL parameters using comma-separated values.

### 3.7. Application Behavior

-   **FC37:** The application MUST support graceful shutdown upon receiving SIGINT (Ctrl+C) or SIGTERM signals.
    -   This includes closing the database connection, stopping log collection goroutines, and shutting down the HTTP server cleanly.
-   **FC38:** The compiled application binary name SHOULD be `docker-util`.

## 4. Non-Functional Requirements

-   **NFR1 (Performance):** The application should be efficient and aim to minimize overhead on the host system and Docker daemon. Database write operations should be resilient to transient errors (e.g., using retries with backoff for database writes). Search and filtering operations should be optimized using database indexes.
-   **NFR2 (Reliability):** The application should run continuously. Log collection and web server should operate reliably. Pagination should handle large datasets efficiently.
-   **NFR3 (Usability):** The command-line interface should be simple. The web UI should be intuitive and provide a clear view of log data with efficient search, filtering, and navigation capabilities.
-   **NFR4 (Maintainability):** The Go codebase should be well-structured and understandable.
-   **NFR5 (Scalability):** The application should handle large log volumes (>1GB databases) efficiently through proper indexing and pagination limits.

## 5. UI/UX Requirements (Web Interface)

-   **UX1:** The log viewing page MUST use Tailwind CSS (via Play CDN) for styling.
-   **UX2:** The log table and its container MUST utilize the full width of the browser window, with appropriate padding for readability.
-   **UX3:** Log messages containing multiple lines or preserved whitespace MUST be displayed correctly in the web UI.
-   **UX4:** The `stream_type` (stdout/stderr) SHOULD be visually distinguishable in the log table (e.g., using color-coded badges).
-   **UX5:** Search input MUST be prominently placed at the top of the page with clear placeholder text.
-   **UX6:** Container filtering checkboxes MUST be organized in a responsive grid layout (1-3 columns based on screen size).
-   **UX7:** Active filters and search state MUST be clearly indicated with visual feedback.
-   **UX8:** Pagination controls MUST be intuitive with clear Previous/Next buttons and page indicators.
-   **UX9:** The interface MUST be fully responsive and functional on mobile devices.
-   **UX10:** Filter and search state information MUST be displayed to help users understand current view context.

## 6. Technical Specifications

-   **Language:** Go (version 1.24.0 or later)
-   **Key Go Libraries/Packages:**
    -   `github.com/docker/docker/client` (Docker API interaction)
    -   `database/sql`, `github.com/mattn/go-sqlite3` (SQLite database)
    -   `net/http`, `net/url` (Web server and URL handling)
    -   `html/template`, `embed` (Server-side HTML templating)
    -   `flag` (Command-line argument parsing)
    -   `os/signal`, `sync`, `context` (Concurrency and graceful shutdown)
    -   `strconv`, `strings` (String processing for pagination and filtering)
-   **Frontend Styling:** Tailwind CSS (via Play CDN)
-   **Frontend JavaScript:** Vanilla JavaScript for form handling and dynamic behavior
-   **Database Indexes:** 
    -   `idx_logs_message` (for search performance)
    -   `idx_logs_container_name` (for filtering performance)
-   **URL Structure:** 
    -   Base: `http://localhost:9123/`
    -   With filters: `http://localhost:9123/?search=query&containers=name1,name2&page=2`

## 7. Future Considerations (Out of Scope for v2.0)

-   ~~Advanced log filtering in the UI (by container name, message content, date range).~~ ✅ **IMPLEMENTED:** Container filtering and message search implemented.
-   Advanced date range filtering and timestamp-based navigation.
-   Log rotation or archival for the SQLite database.
-   Real-time log updates in the web UI (e.g., using HTMX or WebSockets).
-   Configuration file support (in addition to command-line flags).
-   Authentication/Authorization for the web UI.
-   More detailed error reporting and diagnostics in the UI.
-   Support for monitoring remote Docker daemons.
-   ~~Option to specify which containers to monitor (e.g., by name, label).~~ ✅ **IMPLEMENTED:** Dynamic container filtering implemented.
-   Log export functionality (CSV, JSON formats).
-   Advanced search features (regex patterns, boolean operators).
-   Log highlighting and syntax highlighting for common log formats.
-   Dashboard and analytics features (log volume trends, error rates).
-   Integration with external log management systems. 