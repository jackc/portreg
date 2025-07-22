# portreg

`portreg` is a tool to track what ports have been assigned to what services and projects. It is primarily useful for developers who have many projects that may be running concurrently. In these cases where multiple databases and web servers may be running at the same time it is useful to have a way to assign ports to projects and to remember the assignment.

This differs from tools that dynamically find an open port as `portreg` is designed to be used once when the project is first set up on a developer machine. Dynamically assigning a port is often insufficient because the port may be specified in a way where dynamic assignment is inconvenient. e.g. when the port is needed by server and a client.

## Installation

```
$ go install github.com/jackc/portreg@latest
```

## Usage

### init

The `init` command initializes a `portreg` registry file.

```
$ portreg init
```

Options:

* `registry` - override path to port registry file

### assign

The `assign` command is used to assign an unused port. It records the assigned port and prints it to `stdout`.

```
$ portreg assign
12345
```

Options:

* `port` - specific port to assign
* `description` - description of project or service the port is assigned to
* `path` - path to project the port is assigned to
* `registry` - override path to port registry file

### unassign

The `unassign` command is used to unassign a port.

```
$ portreg unassign 12345
```

Options:

* `registry` - override path to port registry file

### list

The `list` command is used to list all assigned ports.

```
$ portreg list
PORT  DESCRIPTION                               PATH
----  -----------                               ----
3100  My service                                /Users/jack/dev/foo
3103  My service 2                              /Users/jack/dev/bar
```

Options:

* `registry` - override path to port registry file

## Registry

The registry file is stored by default in `$HOME/.portreg.json`.

Example file:

```json
{
  "assignments": [
    {
      "port": 5678,
      "description": "description",
      "path": "/path/to/project"
    },
    {
      "port": 5689
    }
  ],
  "blockedPorts": [
    {
      "ports": "3000-3010",
      "description": "common Ruby on Rails ports"
    }
    {
      "port": "5432",
      "description": "default PostgreSQL port"
    }
  ]
}
```
