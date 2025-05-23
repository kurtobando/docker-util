# Product Requirements Document: docker-util Log Collector & Viewer

**Version:** 1.0
**Date:** {{CURRENT_DATE}} 

## 1. Introduction

`docker-util` is a command-line Go application designed to collect logs from all running Docker containers on a host machine. It streams these logs in real-time, stores them in a local SQLite database, and provides a simple web interface for viewing the collected logs. The tool aims to be robust, with graceful shutdown capabilities and configurable options for database storage and web UI port.

## 2. Goals

-   To provide a reliable method for centralizing Docker container logs locally.
-   To offer an accessible web interface for easy viewing and searching (future) of collected logs.
-   To create a lightweight, self-contained utility that is easy to build and run.
-   To ensure data integrity through graceful shutdown procedures.

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

### 3.3. Web User Interface (UI)

-   **FC10:** The application MUST start an HTTP web server to provide a UI for viewing logs.
-   **FC11:** The port for the web server MUST be configurable via a command-line flag (`--port`).
    -   Default port: 9123.
-   **FC12:** The web UI MUST be accessible at the root path (e.g., `http://localhost:9123/`).
-   **FC13:** The web UI MUST display a table of log entries fetched from the SQLite database.
-   **FC14:** The log table MUST display the following columns: Timestamp, Container Name, Container ID (shortened for display, e.g., first 12 characters), Stream Type, and the Log Message.
-   **FC15:** The web UI SHOULD display the latest 100 log entries by default, ordered with the newest logs first.
-   **FC16:** HTML templates for the web UI MUST be embedded within the compiled Go binary.

### 3.4. Application Behavior

-   **FC17:** The application MUST support graceful shutdown upon receiving SIGINT (Ctrl+C) or SIGTERM signals.
    -   This includes closing the database connection, stopping log collection goroutines, and shutting down the HTTP server cleanly.
-   **FC18:** The compiled application binary name SHOULD be `docker-util`.

## 4. Non-Functional Requirements

-   **NFR1 (Performance):** The application should be efficient and aim to minimize overhead on the host system and Docker daemon. Database write operations should be resilient to transient errors (e.g., using retries with backoff for database writes).
-   **NFR2 (Reliability):** The application should run continuously. Log collection and web server should operate reliably.
-   **NFR3 (Usability):** The command-line interface should be simple. The web UI should be intuitive and provide a clear view of log data.
-   **NFR4 (Maintainability):** The Go codebase should be well-structured and understandable.

## 5. UI/UX Requirements (Web Interface)

-   **UX1:** The log viewing page MUST use Tailwind CSS (via Play CDN) for styling.
-   **UX2:** The log table and its container MUST utilize the full width of the browser window, with appropriate padding for readability.
-   **UX3:** Log messages containing multiple lines or preserved whitespace MUST be displayed correctly in the web UI.
-   **UX4:** The `stream_type` (stdout/stderr) SHOULD be visually distinguishable in the log table (e.g., using color-coded badges).

## 6. Technical Specifications

-   **Language:** Go (version 1.24.0 or later)
-   **Key Go Libraries/Packages:**
    -   `github.com/docker/docker/client` (Docker API interaction)
    -   `database/sql`, `github.com/mattn/go-sqlite3` (SQLite database)
    -   `net/http` (Web server)
    -   `html/template`, `embed` (Server-side HTML templating)
    -   `flag` (Command-line argument parsing)
    -   `os/signal`, `sync`, `context` (Concurrency and graceful shutdown)
-   **Frontend Styling:** Tailwind CSS (via Play CDN)

## 7. Future Considerations (Out of Scope for v1.0)

-   Advanced log filtering in the UI (by container name, message content, date range).
-   Log rotation or archival for the SQLite database.
-   Real-time log updates in the web UI (e.g., using HTMX or WebSockets).
-   Configuration file support (in addition to command-line flags).
-   Authentication/Authorization for the web UI.
-   More detailed error reporting and diagnostics in the UI.
-   Support for monitoring remote Docker daemons.
-   Option to specify which containers to monitor (e.g., by name, label). 