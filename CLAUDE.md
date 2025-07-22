# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

portreg is a port registry tool written in Go that helps developers manage port assignments across multiple projects to avoid conflicts. It uses static port assignment (not dynamic) stored in a JSON registry file.

## Key Commands

The tool implements four main commands:
- `init` - Initialize the registry file at `$HOME/.portreg.json` (or custom location via `-r` flag)
- `assign` - Assign an unused port to a project (auto-finds next available or accepts specific port via `-p` flag)
  - Description is optional via `-d` flag
  - Path defaults to current directory, can be overridden with `--path` flag
  - Output: Only the assigned port number (e.g., `3100`)
- `unassign <port>` - Release a port assignment by port number
- `list` - Display all assigned ports
  - Supports `--format json` for JSON output

## Development Commands

### Building
```bash
go build -o portreg
```

### Running
```bash
go run . [command] [args]
```

### Installing
```bash
go install
```

### Testing
```bash
go test ./...
```

## Architecture

### Project Structure
```
portreg/
├── main.go              # Entry point
├── cmd/                 # CLI commands (using Cobra)
│   ├── root.go         # Root command and global flags
│   ├── init.go         # Init command
│   ├── assign.go       # Assign command  
│   ├── unassign.go     # Unassign command
│   └── list.go         # List command
├── registry/           # Core registry package
│   ├── registry.go     # Registry type and all core logic
│   └── registry_test.go # Unit tests
└── .github/
    └── workflows/
        └── test.yml    # CI workflow

```

### Registry Storage
- Default location: `$HOME/.portreg.json`
- JSON format with structure:
  ```json
  {
    "assignments": [
      {
        "port": 8000,
        "description": "Description of project",
        "path": "/path/to/project"
      }
    ],
    "blockedPorts": [
      {
        "ports": "3000-3010",
        "description": "common Ruby on Rails ports"
      }
    ]
  }
  ```
- The `description` and `path` values under `assignments` are optional.
- The `description` value under `blockedPorts` is optional.
- The `ports` value under `blockedPorts` can be a single port or a range separated by a hyphen.

### Key Implementation Considerations

1. **Port Assignment Logic**
   - Check if port is already assigned
   - Check if port is in blocked ranges
   - Auto-assignment should find the lowest available port starting from 3100
   - The default registry file created by `init` includes `blockedPorts` for:
     - MySQL (3306)
     - PostgreSQL (5432)
     - Redis (6379)
     - Common HTTP alternative (8080)
     - MongoDB (27017)

2. **Command Structure**
   - Use cobra for command handling
   - Each command should have appropriate flags and arguments
   - Provide helpful error messages and usage information

3. **File Operations**
   - Ensure atomic writes to prevent registry corruption
   - Handle missing registry file gracefully
   - Validate JSON structure on read/write

4. **Error Handling**
   - Clear messages for port conflicts
   - Helpful suggestions (e.g., "Port 8000 is already assigned to 'project-x'. Use 'portreg list' to see all assignments.")
   - Handle filesystem permissions issues

5. **Core Functionality**
   - Core port logic and persistence logic is in the `registry` package
   - The `Registry` type provides all functionality through methods
   - It is independent from the CLI

6. **Testing**
   - Uses `github.com/stretchr/testify` for assertions
   - Comprehensive unit tests in `registry/registry_test.go`
   - CI runs tests on push/PR via GitHub Actions

## Common Development Tasks

When implementing new features or fixing bugs:
1. Read/modify the registry file using proper JSON marshaling/unmarshaling
2. Ensure backward compatibility with existing registry files
3. Add appropriate unit tests for any new functionality
4. Update the README if adding new commands or changing behavior
