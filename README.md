# Docker Log Collector to SQLite with Web UI

This Go application connects to running Docker containers, streams their logs in real-time, and stores them in an SQLite database. It supports graceful shutdown, allows configuration of the database path, and provides a simple web interface to view the collected logs.

## Features

- Real-time log streaming from all running Docker containers.
- Log storage in an SQLite database.
- Configurable database path via `--dbpath` flag.
- Graceful shutdown on SIGINT/SIGTERM.
- **Web UI to view collected logs (served on `http://localhost:9123` by default).**

## Prerequisites

- Go version 1.24.0 or higher
- Docker installed and running

## Setup

1.  **Clone the repository (or create the project directory):**
    ```bash
    # If you have a git repo already, skip this
    # git clone <your_repo_url>
    # cd <your_project_directory>
    ```

2.  **Initialize Go Module (if not already done):**
    ```bash
    go mod init bw_util
    ```

3.  **Install Dependencies:**
    ```bash
    go get github.com/docker/docker/client
    go get github.com/mattn/go-sqlite3
    ```

## Usage

1.  **Build the application:**
    ```bash
    go build -o docker-util
    ```

2.  **Run the application:**
    ```bash
    ./docker-util [options]
    ```
    The application will:
    - Create/use an SQLite database file (defaults to `./docker_logs.db`).
    - Start collecting logs from all running containers.
    - **Start a web server on `http://localhost:9123`. Open this URL in your browser to view logs.**

    **Options:**
    *   `--dbpath <path>`: Specifies the path for the SQLite database file.
        Example: `./docker-util --dbpath /var/log/docker_app_logs.sqlite`
    *   `--port <port_number>`: Specifies the port for the web UI (defaults to 9123).
        Example: `./docker-util --port 8888`

## Database Schema

The logs are stored in a table named `logs` with the following schema:

-   `id` (INTEGER, PRIMARY KEY, AUTOINCREMENT)
-   `container_id` (TEXT) - Docker container ID (long ID)
-   `container_name` (TEXT) - Primary name of the Docker container
-   `timestamp` (TEXT) - Log entry timestamp (RFC3339Nano format, typically UTC)
-   `stream_type` (TEXT) - "stdout" or "stderr"
-   `log_message` (TEXT) - The actual log content

## Development

-   Ensure Docker daemon is accessible.
-   Run `go run main.go` for development (you can also pass flags: `go run main.go --dbpath ./custom_dev.db --port 8081`). 