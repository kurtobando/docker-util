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
- **Comprehensive unit testing with 55%+ code coverage.**
- **Professional build system with Makefile for development workflow.**

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
    make build
    ./docker-util
    ```
    
    Or run directly:
    ```bash
    make run
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

## Development

**Prerequisites:**
- Ensure Docker daemon is accessible
- Go 1.24.0 or higher

**Available Make Targets:**
```bash
make help                 # Show all available targets
make test                 # Run unit tests
make test-coverage        # Run tests with coverage report
make test-coverage-open   # Run coverage and open report in browser
make test-race            # Run tests with race detection
make test-verbose         # Run tests with verbose output
make build                # Build the application
make run                  # Run the application
make clean-coverage       # Clean coverage files
```

**Development Workflow:**
```bash
# Run tests during development
make test

# Check test coverage
make test-coverage

# Build for testing
make build

# Run in development mode
make run
```

**Testing:**
The project includes comprehensive unit tests with 55%+ code coverage:
```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run tests for specific package
make test-package PKG=internal/app

# Run tests with race detection
make test-race
```

## Architecture

The application follows a clean, modular architecture with comprehensive testing:

```
bw-util/
├── main.go                      # Application entry point (26 lines)
├── main_test.go                 # Main package tests
├── Makefile                     # Build system and development workflow
├── internal/
│   ├── app/                     # Application orchestration
│   │   ├── app.go              # App struct with Initialize/Run methods
│   │   └── app_test.go         # App constructor & validation tests
│   ├── config/                  # Configuration management
│   │   ├── config.go           # Command-line flag parsing
│   │   └── config_test.go      # Config tests (100% coverage)
│   ├── models/                  # Data structures
│   │   ├── log.go              # LogEntry and SearchParams models
│   │   └── models_test.go      # Model validation tests
│   ├── database/                # Database layer (SQLite + repository)
│   │   ├── sqlite.go           # Database connection and schema
│   │   ├── repository.go       # Data access layer
│   │   ├── sqlite_test.go      # Database tests
│   │   └── repository_test.go  # Repository tests (76% coverage)
│   ├── docker/                  # Docker integration
│   │   ├── client.go           # Docker API client
│   │   ├── collector.go        # Log collection orchestration
│   │   ├── parser.go           # Log parsing and timestamp handling
│   │   ├── client_test.go      # Docker client tests
│   │   ├── collector_test.go   # Collector tests
│   │   └── parser_test.go      # Parser tests (23% coverage)
│   ├── web/                     # Web interface
│   │   ├── server.go           # HTTP server with graceful shutdown
│   │   ├── handlers.go         # Request handlers and pagination
│   │   ├── templates.go        # Template management
│   │   ├── server_test.go      # Server tests
│   │   ├── handlers_test.go    # Handler tests (93% coverage)
│   │   └── templates_test.go   # Template tests
│   └── testutil/                # Testing utilities
│       ├── mocks.go            # Mock implementations
│       ├── fixtures.go         # Test data fixtures
│       ├── database.go         # Test database helpers
│       └── *_test.go           # Test utility tests (73% coverage)
├── templates/                   # HTML templates (embedded)
│   └── logs.html               # Main web interface template
└── go.mod                       # Dependencies
```

**Benefits:**
- **Maintainable**: Each module has a single responsibility with comprehensive tests
- **Testable**: 55%+ code coverage with unit tests for all components
- **Reliable**: Professional build system with race detection and coverage reporting
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

**Performance Indexes:**
- `idx_logs_message` - Optimizes search functionality
- `idx_logs_container_name` - Optimizes container filtering

## Testing

The project maintains high code quality through comprehensive testing:

**Test Coverage by Module:**
- `internal/config`: 100% coverage
- `internal/web`: 93% coverage  
- `internal/database`: 76% coverage
- `internal/testutil`: 73% coverage
- `internal/docker`: 23% coverage
- **Overall**: 55%+ coverage

**Test Types:**
- **Unit Tests**: Test individual functions and methods in isolation
- **Integration Tests**: Test component interactions with mocked dependencies
- **Validation Tests**: Test struct validation and error handling
- **Mock Tests**: Use comprehensive mock system for external dependencies

**Running Tests:**
```bash
# Basic test run
make test

# With coverage report
make test-coverage

# With race detection
make test-race

# Verbose output
make test-verbose

# Open coverage report in browser
make test-coverage-open
``` 