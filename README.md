# sachithra-web-analyzer

A simple **web page analyzer** built with Go, split into two components:

- **Client web app (UI)**: serves an HTML page where you paste a URL and submit it for analysis.
- **Analyzer API (server)**: fetches the target page, parses HTML, analyzes structure, and validates links concurrently.

---

## Prerequisites

- Go installed (recommended: Go 1.20+)
- Internet access (the analyzer fetches external pages)
- Ports available:
  - Client UI: **8090**
  - Analyzer API: **8080**

> Note: `go.mod` and `go.sum` are committed, so dependencies should be fetched automatically by Go when you run the apps.

---

## Project Structure

- `web-analyzer-client/` — UI web server (HTML form + forwards request to API)
- `web-analyzer-server/` — Analyzer API (fetch + parse + analyze + link checking)

---

## How to Run (Client + Server)

### 1) Start the client (UI)

Open a terminal and run:

```bash
cd web-analyzer-client
go run cmd/client/client.go
```

This starts the client web server on **http://localhost:8090**.

### 2) Start the analyzer server (API)

Open a second terminal and run:

```bash
cd web-analyzer-server
go run cmd/server/main.go
```

This starts the analyzer API on **http://localhost:8080**.

### 3) Use the app

Open your browser and go to:

- **http://localhost:8090/analyze**

Paste the URL of the webpage you want to analyze into the text field and click **Analyze**.

---

## Configuration

### Worker count (link checking)

The analyzer validates links using a **worker pool** (concurrent HEAD requests).  
If response time is high, tune the number of workers in:

- `config.properties`

Increase workers to speed up link validation, decrease to reduce load / throttling risk.

---

## Dependencies

External libraries used:

- `github.com/PuerkitoBio/goquery`
- `golang.org/x/net/html`

Standard library packages used include (not exhaustive):

- `net/http`, `net/url`
- `context`, `time`
- `encoding/json`
- `sync`
- `text/template`
- `log/slog`

### Installing (if you ever need to)

Normally not required (Go will pull deps automatically), but you can run:

```bash
go get github.com/PuerkitoBio/goquery
go get golang.org/x/net/html
```

---

## Architecture

### Three-tier architecture diagram

See: `three_tier_architecture.png` (located next to this README).

### Descriptive architecture diagram

```text
┌──────────────────────────┐
│        User Browser      │
│   (Web Analyzer Client)  │
└─────────────┬────────────┘
              │  HTTP GET /analyze
              ▼
┌─────────────────────────┐
│   Client Web Server      │
│   (Port 8090)            │
│                         │
│ - HTML Template UI       │
│ - URL Form Input         │
│ - Sends API Request      │
└─────────────┬───────────┘
              │  HTTP POST /analyzer
              ▼
┌─────────────────────────┐
│     Analyzer API         │
│     (Port 8080)          │
│                         │
│ - HTTP Server            │
└─────────────┬───────────┘
              ▼
┌─────────────────────────┐
│       Middleware         │
│ - Logging (slog)         │
│ - Recovery (panic guard) │
│ - Request timing         │
└─────────────┬───────────┘
              ▼
┌─────────────────────────┐
│     Handler Layer        │
│ - Validation             │
│ - Error mapping          │
│ - JSON responses         │
└─────────────┬───────────┘
              ▼
┌─────────────────────────┐
│     FetchService         │
│ - Fetch HTML page        │
│ - Parse content          │
│ - Analyze structure      │
└─────────────┬───────────┘
              ▼
┌─────────────────────────────────┐
│        HTML Processing          │
│ - getHTMLVersion()              │
│ - getTitleAndHeadings()         │
│ - getLinks()                    │
│ - getHasLogin()                 │
└─────────────┬───────────────────┘
              ▼
┌─────────────────────────────────┐
│          Worker Pool            │
│ - Configurable workers          │
│ - HEAD requests for link checks │
└─────────────┬───────────────────┘
              ▼
┌─────────────────────────┐
│    External Websites     │
│ - Link accessibility     │
│   validation (HEAD)      │
└─────────────────────────┘
```

---

## Possible Future Improvements

- Throttling (rate-limit external calls)
- Docker support
- Unit tests
- TLS between client and server
- Caching
- Request tracing
- Retry mechanism
- Health check endpoint (e.g., `/healthz`)

---

## Troubleshooting

### Port already in use
If you see an error like “address already in use”, either:
- stop the process currently using ports **8080** or **8090**, or
- change the ports in the code/config (depending on how your project is set up).

### Slow responses
If the page has many links, link checking can take time.
- Reduce/increase worker count in `config.properties`
- Try analyzing a simpler page to confirm everything is working

---