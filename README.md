# Docker Log Collector to SQLite with Web UI

This Go application connects to running Docker containers, streams their logs in real-time, and stores them in an SQLite database. It supports graceful shutdown, allows configuration of the database path, and provides a simple web interface to view the collected logs.

## Features

- Real-time log streaming from all running Docker containers.
- Log storage in an SQLite database.
- Configurable database path via `--dbpath` flag.
- Graceful shutdown on SIGINT/SIGTERM.
- **Web UI to view collected logs (served on `http://localhost:9123` by default).**
- **Search functionality for filtering log messages (case-insensitive).**
- **Pagination support with Next/Previous navigation (100 records per page).**
- **Dynamic container filtering with checkbox selection (automatically detects available containers).**
- **Modular, maintainable codebase with clean separation of concerns.**

## Prerequisites

- Go version 1.24.0 or higher
- Docker installed and running

## Quick Start

1.  **Clone the repository:**
    ```bash
    git clone <your_repo_url>
    cd bw-util
    ```

2.  **Build and run:**
    ```bash
    go build -o docker-util
    ./docker-util
    ```
    
    That's it! The application will:
    - Create/use an SQLite database file (defaults to `./docker_logs.db`).
    - Start collecting logs from all running containers.
    - **Start a web server on `http://localhost:9123`. Open this URL in your browser to view logs.**

## Usage

**Run the application:**
```bash
./docker-util [options]
```

**Web Interface Features:**
- **Use the search box at the top of the page to filter log messages by text content.**
- **Filter logs by specific containers using the checkboxes (automatically shows available containers).**
- **Combine search and container filtering for precise log filtering.**
- **Navigate through log pages using the Next/Previous buttons at the bottom of the table.**

**Command Line Options:**
*   `--dbpath <path>`: Specifies the path for the SQLite database file.
    Example: `./docker-util --dbpath /var/log/docker_app_logs.sqlite`
*   `--port <port_number>`: Specifies the port for the web UI (defaults to 9123).
    Example: `./docker-util --port 8888`

## Architecture

The application follows a clean, modular architecture:

```
bw-util/
├── main.go                      # Application entry point
├── internal/
│   ├── app/                     # Application orchestration
│   ├── config/                  # Configuration management
│   ├── models/                  # Data structures
│   ├── database/                # Database layer (SQLite + repository)
│   ├── docker/                  # Docker integration (client, collector, parser)
│   └── web/                     # Web interface (server, handlers, templates)
├── templates/                   # HTML templates (embedded)
└── go.mod                       # Dependencies
```

**Benefits:**
- **Maintainable**: Each module has a single responsibility
- **Testable**: Components can be unit tested independently  
- **Extensible**: Easy to add new features without affecting existing code
- **Readable**: Clear separation of concerns makes the code easy to understand

## Database Schema

The logs are stored in a table named `logs` with the following schema:

-   `id` (INTEGER, PRIMARY KEY, AUTOINCREMENT)
-   `container_id` (TEXT) - Docker container ID (long ID)
-   `container_name` (TEXT) - Primary name of the Docker container
-   `timestamp` (TEXT) - Log entry timestamp (RFC3339Nano format, typically UTC)
-   `stream_type` (TEXT) - "stdout" or "stderr"
-   `log_message` (TEXT) - The actual log content

## Development

**Prerequisites:**
- Ensure Docker daemon is accessible
- Go 1.24.0 or higher

**Run in development mode:**
```bash
go run main.go [--dbpath ./dev.db] [--port 8081]
```

**Build for production:**
```bash
go build -o docker-util
```

**Project Structure:**
The codebase is organized into logical modules under `internal/` for better maintainability and testing. Each module handles a specific aspect of the application (configuration, database, Docker integration, web interface, etc.). 