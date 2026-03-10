package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"
)

var startTime = time.Now()

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		uptime := int(time.Since(startTime).Seconds())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Hello Go</title>
  <style>
    body { background: #0f172a; color: #e2e8f0; font-family: monospace;
           display: flex; align-items: center; justify-content: center;
           height: 100vh; margin: 0; }
    .box { text-align: center; }
    h1 { font-size: 3rem; color: #22d3ee; margin-bottom: 0.5rem; }
    p  { color: #94a3b8; }
  </style>
</head>
<body>
  <div class="box">
    <h1>&#9632; Hello, Go!</h1>
    <p>Running on Local VSP &mdash; Go %s</p>
    <p style="margin-top:1rem;font-size:0.75rem;color:#475569">
      hostname: %s &nbsp;|&nbsp; uptime: %ds
    </p>
  </div>
</body>
</html>`, runtime.Version(), hostname, uptime)
	})

	fmt.Println("hello-go listening on port 8080")
	http.ListenAndServe("0.0.0.0:8080", nil)
}
