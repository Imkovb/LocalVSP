import http.server
import os
import platform
import time

PORT = int(os.environ.get("PORT", 3000))
START_TIME = time.time()


class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        uptime = int(time.time() - START_TIME)
        hostname = platform.node()
        py_version = platform.python_version()

        html = f"""<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Hello Python</title>
  <style>
    body {{ background: #0f172a; color: #e2e8f0; font-family: monospace;
           display: flex; align-items: center; justify-content: center;
           height: 100vh; margin: 0; }}
    .box {{ text-align: center; }}
    h1 {{ font-size: 3rem; color: #60a5fa; margin-bottom: 0.5rem; }}
    p  {{ color: #94a3b8; }}
  </style>
</head>
<body>
  <div class="box">
    <h1>&#9632; Hello, Python!</h1>
    <p>Running on Local VSP &mdash; Python {py_version}</p>
    <p style="margin-top:1rem;font-size:0.75rem;color:#475569">
      hostname: {hostname} &nbsp;|&nbsp; uptime: {uptime}s
    </p>
  </div>
</body>
</html>"""

        body = html.encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        print(f"{self.address_string()} - {format % args}", flush=True)


if __name__ == "__main__":
    server = http.server.HTTPServer(("0.0.0.0", PORT), Handler)
    print(f"hello-python listening on port {PORT}", flush=True)
    server.serve_forever()
