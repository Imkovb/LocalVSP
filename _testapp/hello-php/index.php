<?php
$hostname = gethostname();
$uptime = time() - filectime('/proc/1/cmdline');
$phpVersion = phpversion();
?>
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Hello PHP</title>
  <style>
    body { background: #0f172a; color: #e2e8f0; font-family: monospace;
           display: flex; align-items: center; justify-content: center;
           height: 100vh; margin: 0; }
    .box { text-align: center; }
    h1 { font-size: 3rem; color: #a78bfa; margin-bottom: 0.5rem; }
    p  { color: #94a3b8; }
  </style>
</head>
<body>
  <div class="box">
    <h1>&#9632; Hello, PHP!</h1>
    <p>Running on Local VSP &mdash; PHP <?= $phpVersion ?></p>
    <p style="margin-top:1rem;font-size:0.75rem;color:#475569">
      hostname: <?= $hostname ?> &nbsp;|&nbsp; uptime: <?= $uptime ?>s
    </p>
  </div>
</body>
</html>
