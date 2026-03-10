const http = require('http');

const PORT = process.env.PORT || 3000;

const server = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/html' });
  res.end(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Hello Node.js</title>
  <style>
    body { background: #0f172a; color: #e2e8f0; font-family: monospace;
           display: flex; align-items: center; justify-content: center;
           height: 100vh; margin: 0; }
    .box { text-align: center; }
    h1 { font-size: 3rem; color: #4ade80; margin-bottom: 0.5rem; }
    p  { color: #94a3b8; }
  </style>
</head>
<body>
  <div class="box">
    <h1>&#9632; Hello, Node.js!</h1>
    <p>Running on Local VSP &mdash; Node.js ${process.version}</p>
    <p style="margin-top:1rem;font-size:0.75rem;color:#475569">
      hostname: ${require('os').hostname()} &nbsp;|&nbsp; uptime: ${Math.floor(process.uptime())}s
    </p>
  </div>
</body>
</html>`);
});

server.listen(PORT, () => {
  console.log(`hello-nodejs listening on port ${PORT}`);
});
