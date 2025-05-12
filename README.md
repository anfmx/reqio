# 📘 HTTP Client on Go

## 📦 Overview

`reqio` is a command-line HTTP client written in Go. It supports:

- Sending single or concurrent HTTP requests
- Load and stress testing
- Performance and status reporting
- Support for HTTP methods: GET, POST, PUT, DELETE, etc.

---

## Installation

### Using `go install`:

```bash
go install github.com/anfmx/reqio@latest
```

### Manual build:

```bash
git clone https://github.com/anfmx/reqio
cd reqio
go build -o reqio .
```

## Usage

### Arguments

| Flag           | Type   | Default | Description                        |
| -------------- | ------ | ------- | ---------------------------------- |
| `-m`           | string | `"GET"` | HTTP method (e.g., GET, POST)      |
| `-b`           | bool   | `false` | Show response body                 |
| `--limit`      | int    | `0`     | Limit response body output         |
| `--rate`       | int    | `0`     | Requests per second                |
| `-f`           | bool   | `false` | Write response to a json file      |
| `--time-limit` | int    | `10`    | Time limit for execution (seconds) |
| `-d`           | string | `""`    | Request body                       |

## Examples

### Basic request

```bash
reqio https://yourlink.com/ #GET by default
```

```bash
reqio -m POST https://yourlink.com/
```

```bash
# sends a get request every second for 10 seconds, outputting and limiting the request body to 1 element and writes the result to a file
reqio -f -b --limit 1 --rate 1 --time-limit 10  https://yourlink.com/
```

### Multiple Requests

```bash
reqio https://yourlink.com/ https://youtube.com/ https://twitch.tv/
```

### Answer

```json
{
  "url": "https://www.yourlink.com/",
  "status_code": "200 OK",
  "body": null //if -b flag was not specified or there is no body in reply
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
