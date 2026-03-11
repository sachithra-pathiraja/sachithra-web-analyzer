# sachithra-web-analyzer

How to start the server and client

1. Open a new terminal
2. Run the following commands in order

cd web-analyzer-client 
go run cmd/client/client.go

3. Open another terminal
4. Run the following commands in order

cd web-analyzer-server 
go run cmd/server/main.go

5. Go to your browser and load the following URL

http://localhost:8090/analyze

6. Paste the URL of the webpage you need to analyze on the only textfield in the page and press Analyze
7. Since the go.mod and go.sum files commited to the repo you may not need to install any of the libraries. but all the used libraries are as following. you can use following commands in a case you need to install external library.

go get <ex library>

"bytes"
"context"
"encoding/json"
"io"
"log/slog"
"net/http"
"os"
"os/signal"
"syscall"
"text/template"
"time"
"encoding/json"
"errors"
"github.com/PuerkitoBio/goquery"
"golang.org/x/net/html"
"net/url"
"strings"
"sync"


8. If your response time is very high you can adjust the number of workers used in analyzing links in the config.properties file.

Things I can add later

1. Throttling - I can add throttling to manage the rate of api calls.
2. Intigrate Docker.
3. Adding unit testing.
4. TLS handshake between client and the server
5. Caching mechanism
6. Add request tracing
7. Retry mechanism
8. Health check endpoint


Architectural Diagram

Three tier arcitecture diagram: three_tier_architecture.png (This can be found in the same level as README.md in this repo)

Discriptive architecture diagram: 

                                     ┌──────────────────────────┐
                   │        User Browser      │
                   │   (Web Analyzer Client)  │
                   └─────────────┬────────────┘
                                 │
                                 │ HTTP POST /analyze
                                 ▼
                    ┌─────────────────────────┐
                    │     Client Web Server   │
                    │        (Port 8090)      │
                    │                         │
                    │  - HTML Template UI     │
                    │  - URL Form Input       │
                    │  - Sends API Request    │
                    └─────────────┬───────────┘
                                  │
                                  │ HTTP POST /analyzer
                                  ▼
                    ┌─────────────────────────┐
                    │      Analyzer API       │
                    │        (Port 8080)      │
                    │                         │
                    │      HTTP Server        │
                    └─────────────┬───────────┘
                                  │
                                  ▼
                    ┌─────────────────────────┐
                    │       Middleware        │
                    │                         │
                    │  - Logging Middleware   │
                    │    (slog request logs)  │
                    │                         │
                    │  - Recovery Middleware  │
                    │    (panic protection)   │
                    │                         │
                    │  - Request Timing       │
                    └─────────────┬───────────┘
                                  │
                                  ▼
                    ┌─────────────────────────┐
                    │       Handler Layer     │
                    │                         │
                    │  - Request Validation   │
                    │  - Error Mapping        │
                    │  - JSON Responses       │
                    └─────────────┬───────────┘
                                  │
                                  ▼
                    ┌─────────────────────────┐
                    │      FetchService       │
                    │                         │
                    │  Responsibilities:      │
                    │  - Fetch HTML page      │
                    │  - Parse content        │
                    │  - Analyze structure    │
                    │                         │
                    │  Uses:                  │
                    │  - http.Client          │
                    │  - slog Logger          │
                    └─────────────┬───────────┘
                                  │
                                  ▼
                ┌─────────────────────────────────┐
                │        HTML Processing          │
                │                                 │
                │  getHTMLVersion()               │
                │  getTitleAndHeadings()          │
                │  getLinks()                     │
                │  getHasLogin()                  │
                └─────────────┬───────────────────┘
                              │
                              ▼
              ┌─────────────────────────────────┐
              │         Worker Pool              │
              │                                 │
              │ Configurable Workers            │
              │ (from config.properties)        │
              │                                 │
              │  Worker 1 ── HEAD request ──►   │
              │  Worker 2 ── HEAD request ──►   │
              │  Worker 3 ── HEAD request ──►   │
              │  Worker N ── HEAD request ──►   │
              └─────────────┬───────────────────┘
                            │
                            ▼
                ┌─────────────────────────┐
                │     External Websites   │
                │                         │
                │  Link accessibility     │
                │  validation (HEAD)      │
                └─────────────────────────┘