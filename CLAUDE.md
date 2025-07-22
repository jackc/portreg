# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

portreg is a port registry tool written in Go that helps developers manage port assignments across multiple projects to avoid conflicts. It uses static port assignment (not dynamic) stored in a JSON registry file.

## Key Commands

The tool implements four main commands:
- `init` - Initialize the registry file at `$HOME/.portreg.json` (or custom location via command line argument)
- `assign` - Assign an unused port to a project (auto-finds next available or accepts specific port)
- `unassign` - Release a port assignment by port number
- `list` - Display all assigned ports

## Development Commands

### Building
```bash
go build -o portreg
```

### Running
```bash
go run main.go [command] [args]
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
- The `description` value under `blockedPorts` are optional.
- The `ports` value under `blockedPorts` can be a range separated by a hyphen.

### Key Implementation Considerations

1. **Port Assignment Logic**
   - Check if port is already assigned
   - Check if port is in blocked ranges
   - Auto-assignment should find the lowest available port starting from 3100
   - The default registry file created by `init` should have `blockedPorts` for the default ports for MySQL, PostgreSQL, and other common network servers used in web development.

2. **Command Structure**
   - Do not use a 3rd party CLI framework (e.g., cobra, urfave/cli) for command handling
   - Use the standard library exclusively
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
   - Storage layer should be independent from the CLI or the port choosing logic
   - Port choosing logic should be independent from the CLI or storage layer

6. **Testing**
   - Extensively test all functionality
   - Use the standard library only

## Common Development Tasks

When implementing new features or fixing bugs:
1. Read/modify the registry file using proper JSON marshaling/unmarshaling
2. Ensure backward compatibility with existing registry files
3. Add appropriate unit tests for any new functionality
4. Update the README if adding new commands or changing behavior
