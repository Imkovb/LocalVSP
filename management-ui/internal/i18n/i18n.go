// Package i18n provides English and Dutch translations for the Local VSP
// management interface.  Language preference is stored in a "lang" cookie.
package i18n

import "net/http"

// SupportedLangs lists every language code the UI knows about.
var SupportedLangs = []string{"en", "nl"}

// LangName maps codes to human-readable names for the language picker.
var LangName = map[string]string{
	"en": "English",
	"nl": "Nederlands",
}

// LangFlag maps codes to emoji flags.
var LangFlag = map[string]string{
	"en": "🇬🇧",
	"nl": "🇳🇱",
}

// Detect reads the ?lang= query param, falling back to the "lang" cookie, then
// to "en".  If a valid lang query param is present it also sets the cookie so
// the preference is remembered.
func Detect(w http.ResponseWriter, r *http.Request) string {
	if q := r.URL.Query().Get("lang"); q != "" {
		if isValid(q) {
			http.SetCookie(w, &http.Cookie{
				Name:  "lang",
				Value: q,
				Path:  "/",
				// 1 year
				MaxAge: 60 * 60 * 24 * 365,
			})
			return q
		}
	}
	if c, err := r.Cookie("lang"); err == nil && isValid(c.Value) {
		return c.Value
	}
	return "en"
}

func isValid(lang string) bool {
	for _, l := range SupportedLangs {
		if l == lang {
			return true
		}
	}
	return false
}

// T returns the translation map for the given language.
// Unknown keys fall back to the English string; unknown languages fall back to
// English entirely.
func T(lang string) map[string]string {
	base := translations["en"]
	if lang == "en" {
		return base
	}
	target, ok := translations[lang]
	if !ok {
		return base
	}
	// Merge: start from English so any untranslated key keeps the EN string.
	merged := make(map[string]string, len(base))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range target {
		merged[k] = v
	}
	return merged
}

// ─── Translation tables ───────────────────────────────────────────────────────

var translations = map[string]map[string]string{

	// ── English ───────────────────────────────────────────────────────────────
	"en": {
		// ---- Navigation ----
		"nav.brand":     "Local VSP",
		"nav.dashboard": "Dashboard",
		"nav.logs":      "Logs",
		"nav.settings":  "Settings",
		"nav.help":      "Help",
		"nav.gitea":     "Gitea ↗",
		"dashboard.page.title":         "Local VSP Dashboard",
		"dashboard.hero.badge":         "Server control surface",
		"dashboard.hero.title":         "Run local apps, static sites, and platform services from one dashboard.",
		"dashboard.hero.body":          "The dashboard is organized around health, deploy state, and exposure. Primary actions stay upfront; destructive and networking changes stay contained.",
		"dashboard.hero.endpoint.title": "Endpoint",
		"dashboard.hero.endpoint.body":  "Management UI on the local network",
		"dashboard.hero.state.title":    "Platform state",
		"dashboard.hero.state.metric":   "active containers",
		"dashboard.hero.state.body":     "Across Local VSP core services and deployed workloads",

		// ---- Shared ----
		"btn.refresh":   "↺ Refresh",
		"btn.deploy":    "Deploy",
		"btn.redeploy":  "Redeploy",
		"btn.stop":      "Stop",
		"btn.restart":   "Restart",
		"btn.remove":    "Remove",
		"btn.delete":    "Delete",
		"btn.logs":      "Logs",
		"btn.save":      "Save Settings",
		"btn.save_short": "Save",
		"btn.start":     "Start",
		"lbl.loading":   "Loading…",
		"lbl.working":   "working…",
		"lbl.status":    "Status",
		"lbl.actions":   "Actions",
		"lbl.autostart": "Auto-Start",
		"lbl.webaddr":   "Web Address",
		"shared.help.title": "Help",
		"shared.help.more":  "Full documentation",
		"notify.confirm.title": "Confirm action",
		"notify.confirm.cancel": "Cancel",
		"notify.confirm.confirm": "Confirm",
		"notify.confirm.type_label": "Type the name to confirm deletion",
		"notify.confirm.type_hint": "This permanently removes all files and containers for",
		"notify.confirm.type_placeholder": "Type the exact name",
		"autostart.on.short": "Auto ON",
		"autostart.off.short": "Auto OFF",
		"autostart.on.title": "Auto-start ON — click to disable",
		"autostart.off.title": "Auto-start OFF — click to enable",
		"row.path": "Path",
		"row.exposure": "Exposure",
		"row.domain": "Domain",
		"row.port": "Port",
		"row.no_address": "No public address configured yet.",
		"row.webaddr.placeholder": "web address",
		"confirm.project.stop": "Stop %s?",
		"confirm.project.delete": "Delete %s? This will stop all containers and permanently remove all files.",
		"confirm.site.stop": "Stop %s?",
		"confirm.site.delete": "Delete %s? This will stop the container and permanently remove all files.",
		"confirm.infra.stop": "Stop %s? It will become unavailable.",
		"confirm.infra.rebuild": "Rebuild %s from source? This will cause a brief outage.",
		"confirm.infra.update": "Pull latest %s image and recreate?",

		// ---- Build panel ----
		"build.panel.title":    "Build Output",
		"build.status.building": "building\u2026",
		"build.status.running": "building\u2026",
		"build.status.success": "Build complete",
		"build.status.failed":  "Build failed",
		"build.btn.viewlog":    "View Build",
		"build.lines":          "lines",

		// ---- Status badges ----
		"status.running": "Running",
		"status.stopped": "Stopped",
		"status.unknown": "Unknown",

		// ---- System info bar ----
		"sysinfo.docker":     "Docker Version",
		"sysinfo.containers": "Containers",
		"sysinfo.disk":       "Disk Usage",
		"sysinfo.load":       "System Load",
		"sysinfo.memory":     "Memory Usage",
		"sysinfo.running":    "running",
		"sysinfo.total":      "total",

		// ---- Custom Apps section ----
		"apps.title": "Custom Apps",
		"apps.desc":  "Drop a project folder (Node.js, Python, Go, PHP, Rust, HTML) or a docker-compose.yml into",
		"apps.col.project": "Project",
		"apps.row.subtitle": "Docker project workspace",
		"apps.row.building.help": "Build output is streaming in the side panel.",
		"apps.row.services.active": "services active",
		"apps.row.running.help": "Project is serving traffic.",
		"apps.row.status.partial": "Partial",
		"apps.row.partial.help": "Some services are up, but the project is not fully healthy.",
		"apps.row.status.ready": "Ready to auto-build",
		"apps.row.ready.help": "A supported project was detected and can be containerized automatically.",
		"apps.row.stopped.help": "Deploy to start the project and expose its addresses.",
		"apps.empty.title": "No application folders detected yet.",
		"apps.empty.body": "No folders found in",
		"apps.empty.hint": "Add a subfolder with source code or a docker-compose.yml via the network share.",
		"apps.help.title":  "Custom Apps — Help",
		"apps.help.body": `<p>Custom Apps are programs that run inside Docker containers on your server. They are stored in the <strong>docker</strong> network share.</p>
<h4>Supported project types (auto-detected):</h4>
<ul>
<li><strong>Node.js</strong> — Place a folder with a <code>package.json</code>. A Dockerfile and docker-compose.yml are generated automatically (port 3000).</li>
<li><strong>Python</strong> — Place a folder with a <code>requirements.txt</code>. Auto-generated with Python 3.11 (port 8000, runs <code>app.py</code>).</li>
<li><strong>Go</strong> — Place a folder with a <code>go.mod</code>. Multi-stage build with Go 1.22 (port 8080).</li>
<li><strong>PHP</strong> — Place a folder with a <code>composer.json</code>. Apache + PHP 8.3 with Composer dependencies (port 80).</li>
<li><strong>Rust</strong> — Place a folder with a <code>Cargo.toml</code>. Multi-stage build with Rust 1.77 (port 8080).</li>
<li><strong>Static HTML</strong> — Place a folder with an <code>index.html</code>. Served by Nginx (port 80).</li>
<li><strong>Custom</strong> — Provide your own <code>docker-compose.yml</code> (and optionally a <code>Dockerfile</code>).</li>
</ul>
<h4>How to deploy:</h4>
<ol>
<li>Open File Explorer and navigate to <code>\\&lt;server-ip&gt;\vsp-home\docker</code>.</li>
<li>Copy your project folder there.</li>
<li>On this page, press <strong>Deploy</strong>. A build panel slides open showing real-time terminal output.</li>
<li>After the build, the system verifies the app is actually running. If the app crashes, the build panel shows the error and the last 30 lines of container logs.</li>
<li>Once running, a clickable link appears in the <strong>Web Address</strong> column.</li>
</ol>
<h4>Buttons:</h4>
<ul>
<li><strong>Deploy</strong> — Builds and starts the app (or rebuilds if already deployed).</li>
<li><strong>Stop</strong> — Stops the app without deleting files.</li>
<li><strong>View Build</strong> — Opens the build panel to see the last build output.</li>
<li><strong>Auto-Start</strong> — When on, the app restarts automatically after a server reboot.</li>
</ul>`,

		// ---- Static Websites section ----
		"html.title": "Static Websites",
		"html.desc":  "Drop a folder of static files into",
		"html.desc2": "— served by Traefik automatically",
		"html.col.site": "Site",
		"html.row.subtitle": "Static website folder",
		"html.row.running.help": "Static site is published and reachable through its port or domain.",
		"html.row.stopped.help": "Deploy to publish the site and assign an address.",
		"html.empty.title": "No static sites detected yet.",
		"html.empty.body": "No folders found in",
		"html.empty.hint": "Add a subfolder via the network share and put your index.html inside it.",
		"html.help.title": "Static Websites — Help",
		"html.help.body": `<p>Static Websites are simple folder-based HTML sites served directly by an Nginx container — no programming required.</p>
<h4>How to publish a website:</h4>
<ol>
<li>Open File Explorer and go to <code>\\&lt;server-ip&gt;\vsp-home\html</code>.</li>
<li>Create a new folder with your website name (e.g. <code>my-site</code>).</li>
<li>Place your <code>index.html</code> and any images / CSS files inside that folder.</li>
<li>On this page, press <strong>Deploy</strong> next to the site name.</li>
<li>A port is automatically allocated (range 8200–8299) and the site goes live.</li>
</ol>
<h4>Subdomain access:</h4>
<p>You can optionally assign a subdomain (e.g. <code>my-site.yourdomain.com</code>) by typing it into the <strong>Web Address</strong> field when the site is stopped, then deploying.</p>
<h4>Buttons:</h4>
<ul>
<li><strong>Deploy</strong> — Makes the site live (or redeploys after file changes).</li>
<li><strong>Stop</strong> — Takes the site offline (files are not deleted).</li>
<li><strong>Auto-Start</strong> — When on, the site restarts automatically after a server reboot.</li>
</ul>`,

		// ---- Settings: Core Services ----
		"settings.infra.title": "Core Services",
		"settings.infra.desc":  "Core VSP services — start, stop, restart or update each component.",
		"settings.infra.col.service": "Service",
		"settings.infra.col.version": "Version",
		"settings.infra.col.port":    "Port",
		"settings.infra.not_configured": "not configured",
		"settings.infra.inactive": "inactive",
		"settings.infra.configure": "Configure",
		"settings.infra.rebuild": "Rebuild",
		"settings.infra.update": "Update",
		"settings.infra.local_build": "local build",
		"settings.infra.update_available": "update available",
		"settings.infra.up_to_date": "up to date",
		"settings.infra.help.title":  "Core Services — Help",
		"settings.infra.help.body": `<p>These are the built-in services that power the Local VSP platform itself. You normally do not need to touch these.</p>
<ul>
<li><strong>Traefik</strong> — The reverse-proxy that routes traffic to each app by domain name.</li>
<li><strong>Gitea</strong> — A self-hosted Git server where you can store your source code.</li>
<li><strong>Management UI</strong> — This web interface you are using right now.</li>
</ul>
<h4>Buttons:</h4>
<ul>
<li><strong>Restart</strong> — Quickly restarts the service without losing data.</li>
<li><strong>Update</strong> — Pulls the latest version and re-creates the container.</li>
</ul>
<p class="text-yellow-300">⚠ Restarting the Management UI will briefly disconnect your browser. Refresh after ~10 seconds.</p>`,

		// ---- Settings: Platform Settings ----
		"settings.platform.title":     "Platform Settings",
		"settings.platform.desc":      "Changes are written to /opt/localvsp/.env.",
		"settings.cf.title":           "Cloudflare Tunnel",
		"settings.cf.token.label":     "Tunnel Token",
		"settings.cf.token.hint":      "→ Networks → Tunnels → Create a tunnel. Saving activates the tunnel automatically.",
		"settings.cf.clear":           "Clear the saved tunnel token on save",
		"settings.domain.title":       "Domain",
		"settings.domain.label":       "Base Domain",
		"settings.domain.hint":        "Subdomains will be:",
		"settings.domain.clear":       "Clear the saved base domain on save",
		"settings.hero.badge":         "Platform configuration",
		"settings.hero.title":         "Manage Local VSP infrastructure and exposure settings.",
		"settings.hero.body":          "Infrastructure actions stay separate from Cloudflare and domain configuration, so service operations and routing changes are easier to reason about.",
		"settings.summary.domain.label": "Base domain",
		"settings.summary.domain.empty": "not configured",
		"settings.summary.token.label":  "Tunnel token",
		"settings.summary.token.present": "configured",
		"settings.summary.token.missing": "missing",
		"settings.effects.title":      "What this affects",
		"settings.effects.body":       "Tunnel and domain changes control how project subdomains become reachable outside your local network.",
		"settings.effects.local":      "Without a domain: local port links still work.",
		"settings.effects.domain":     "With a domain only: subdomain labels can be generated, but internet exposure still depends on your tunnel and DNS setup.",
		"settings.effects.tunnel":     "With both configured: the dashboard can expose apps at named hostnames.",
		"settings.platform.help.title": "Platform Settings — Help",
		"settings.platform.help.body": `<p>These settings configure how Local VSP connects apps and websites to the internet through Cloudflare.</p>
<h4>Cloudflare Tunnel Token:</h4>
<ol>
<li>Go to <a href="https://one.dash.cloudflare.com/" target="_blank" class="text-blue-400 underline">one.dash.cloudflare.com</a> and log in.</li>
<li>Open <strong>Networks → Tunnels</strong> and create a new tunnel.</li>
<li>Choose the <strong>Cloudflared</strong> connector type and give the tunnel a name, for example <code>LocalVSP</code>.</li>
<li>On the connector page, copy the long tunnel token string.</li>
<li>In Local VSP Settings, paste that token into <strong>Tunnel Token</strong> and save.</li>
</ol>
<h4>Base Domain:</h4>
<p>Enter the domain you own and have pointed at Cloudflare (e.g. <code>example.com</code>). Only apps and websites where you assign a subdomain will be reachable at <code>subdomain.example.com</code>. The Local VSP dashboard and other management pages stay local-only.</p>`,

		// ---- Logs page ----
		"logs.title":       "Logs",
		"logs.lines.label": "Lines:",
		"logs.autorefresh": "Auto-refresh (5s)",
		"logs.output":      "Output",
		"logs.loading":     "Loading logs...",
		"logs.back":        "← Dashboard",

		// ---- Help page ----
		"help.page.title":    "Help & Documentation",
		"help.page.subtitle": "Step-by-step guides for using Local VSP",
		"help.hero.badge":    "Operations guide",
		"help.hero.summary":  "Use this page as the operational reference for local access, deployments, build output, and Cloudflare publication. The cloud tunnel only exposes apps and sites that you deliberately assign a subdomain to.",
		"help.hero.note.title": "Important routing rule",
		"help.hero.note.body":  "Management UI, Settings, Help, Gitea, and other admin pages stay local-only. Only apps and sites with an assigned subdomain are published through Cloudflare.",
		"help.hero.public.title": "Public access",
		"help.hero.public.body":  "Set one wildcard hostname in Cloudflare: *.yourdomain.com → http://traefik:80. Then assign subdomains per app/site inside Local VSP.",
		"help.hero.admin.title": "Admin surface",
		"help.hero.admin.body":  "Use the local network addresses for Management UI and Gitea. Do not create public Cloudflare hostnames for them.",
		"help.toc":           "Table of Contents",
		"help.section.overview.title": "What is Local VSP?",
		"help.section.overview.body": `<p>Local VSP (Virtual Server Platform) is a self-hosted server management system that lets you run web applications, websites and services from a local server — without needing technical knowledge.</p>
<p class="mt-2">Everything is managed through this web interface. No command line required.</p>
<h4 class="mt-4 font-bold text-white">Key features:</h4>
<ul class="mt-2 space-y-1">
  <li><strong class="text-green-400">Auto-detection</strong> — Drop a Node.js, Python, Go, PHP, Rust or static HTML project folder on the server and Local VSP automatically generates a Dockerfile and docker-compose.yml for you.</li>
  <li><strong class="text-green-400">Build Panel</strong> — Real-time terminal output when deploying an app — streams build progress live and verifies the app is actually running.</li>
  <li><strong class="text-green-400">Custom Apps</strong> — Source-code folders you place on the server, built and run as Docker containers.</li>
  <li><strong class="text-green-400">Static Websites</strong> — Simple HTML/CSS folders published as websites.</li>
  <li><strong class="text-green-400">Core Services</strong> — Built-in platform services: Traefik (reverse proxy), Gitea (Git hosting), Cloudflare Tunnel (internet access).</li>
  <li><strong class="text-green-400">Persistent Logs</strong> — Build logs are saved to disk and survive page refreshes, tab closes and server restarts.</li>
</ul>`,

		"help.section.access.title": "Accessing the Dashboard",
		"help.section.access.body": `<ol class="space-y-2">
  <li><strong>1.</strong> Make sure you are on the same network as the server (home Wi-Fi or office LAN).</li>
  <li><strong>2.</strong> Open a web browser (Chrome, Firefox or Edge).</li>
  <li><strong>3.</strong> Type the server's address in the address bar, for example: <code class="text-green-400">http://192.168.1.100:8080</code></li>
  <li><strong>4.</strong> The Local VSP Dashboard will appear.</li>
</ol>
<p class="mt-3 text-gray-400">Tip: Ask your IT contact for the exact IP address of your server.</p>`,

		"help.section.customapps.title": "Publishing a Custom App",
		"help.section.customapps.body": `<ol class="space-y-3">
  <li><strong>Step 1.</strong> On your Windows PC, open <strong>File Explorer</strong>.</li>
  <li><strong>Step 2.</strong> In the address bar, type <code class="text-green-400">\\&lt;server-ip&gt;\vsp-home\docker</code> and press Enter.</li>
  <li><strong>Step 3.</strong> Copy or drag your application folder into this network location.</li>
  <li><strong>Step 4.</strong> Return to the Dashboard in your browser.</li>
  <li><strong>Step 5.</strong> Your app folder will appear in the <strong>Custom Apps</strong> table. Press <strong>Deploy</strong>.</li>
  <li><strong>Step 6.</strong> The <strong>Build Panel</strong> slides open from the right, showing real-time build output. You can continue using the dashboard while it builds.</li>
  <li><strong>Step 7.</strong> After building, the system waits 3 seconds and checks that your app is actually running. If the app crashed, the panel turns red and shows the last 30 lines of container logs.</li>
  <li><strong>Step 8.</strong> On success, click the link in the <strong>Web Address</strong> column to open your app.</li>
</ol>
<div class="mt-4 p-3 bg-blue-900/40 border border-blue-700 rounded text-sm">
  <strong class="text-blue-300">Auto-detection:</strong>
  <p class="mt-1 text-gray-300">You do <strong>not</strong> need a Dockerfile or docker-compose.yml for supported project types. The system detects your project automatically:</p>
  <ul class="mt-1 text-gray-300">
    <li><code>package.json</code> → Node.js (port 3000)</li>
    <li><code>requirements.txt</code> → Python (port 8000)</li>
    <li><code>go.mod</code> → Go (port 8080)</li>
    <li><code>composer.json</code> → PHP (port 80)</li>
    <li><code>Cargo.toml</code> → Rust (port 8080)</li>
    <li><code>index.html</code> → Static HTML (port 80)</li>
  </ul>
  <p class="mt-1 text-gray-300">For other languages or custom setups, provide your own <code>docker-compose.yml</code>.</p>
</div>
<div class="mt-3 p-3 bg-green-900/30 border border-green-800 rounded text-sm">
  <strong class="text-green-400">Multi-build:</strong>
  <p class="mt-1 text-gray-300">You can deploy multiple apps at the same time. Each build gets its own tab in the Build Panel. Closing the browser does not cancel a running build — reconnect later to see the result.</p>
</div>`,

		"help.section.staticsite.title": "Publishing a Static Website",
		"help.section.staticsite.body": `<ol class="space-y-3">
  <li><strong>Step 1.</strong> On your Windows PC, open <strong>File Explorer</strong>.</li>
  <li><strong>Step 2.</strong> Navigate to <code class="text-green-400">\\&lt;server-ip&gt;\vsp-home\html</code>.</li>
  <li><strong>Step 3.</strong> Create a new folder for your website (e.g. <code>company-site</code>).</li>
  <li><strong>Step 4.</strong> Copy your website files (<code>index.html</code>, images, CSS) into that folder.</li>
  <li><strong>Step 5.</strong> On the Dashboard, find your site in the <strong>Static Websites</strong> table and press <strong>Deploy</strong>.</li>
  <li><strong>Step 6.</strong> A port is automatically allocated (range 8200–8299) and the site goes live. Click the link in the <strong>Web Address</strong> column.</li>
</ol>
<div class="mt-4 p-3 bg-yellow-900/40 border border-yellow-700 rounded text-sm">
  <strong class="text-yellow-300">Important:</strong> Every static website folder must contain a file named exactly <code>index.html</code>. This is the first page visitors will see.
</div>
<div class="mt-3 p-3 bg-blue-900/40 border border-blue-700 rounded text-sm">
  <strong class="text-blue-300">Subdomain access:</strong>
  <p class="mt-1 text-gray-300">You can optionally type a subdomain into the Web Address field (when the site is stopped), then deploy. If a Cloudflare Tunnel and base domain are configured, the site will also be available at <code>subdomain.yourdomain.com</code>.</p>
</div>`,

		"help.section.logs.title": "Viewing Logs",
		"help.section.logs.body": `<p>Local VSP has two types of logs:</p>
<h4 class="font-bold text-white mt-2">Build Logs (deploy-time)</h4>
<p class="mt-1">When you press <strong>Deploy</strong>, the Build Panel shows live build output. This log is saved to disk — you can reopen it later by clicking <strong>View Build</strong> next to the app. Build logs survive page refreshes, browser closes, and even server restarts.</p>
<h4 class="font-bold text-white mt-3">Runtime Logs (container output)</h4>
<p class="mt-1">Once an app is running, you can view its live output:</p>
<ol class="space-y-2 mt-2">
  <li><strong>1.</strong> In the <strong>Custom Apps</strong> table, click the <strong>Logs</strong> link next to the app.</li>
  <li><strong>2.</strong> The log viewer opens showing the latest container output.</li>
  <li><strong>3.</strong> It auto-refreshes every 5 seconds by default.</li>
  <li><strong>4.</strong> Use the <strong>Lines</strong> dropdown to see more or fewer lines (50, 100, 200, or 500).</li>
  <li><strong>5.</strong> Toggle <strong>Auto-refresh</strong> off if you want to pause and read a specific section.</li>
</ol>`,

		"help.section.settings.title": "Configuring Settings",
		"help.section.settings.body": `<h4 class="font-bold text-white">Cloudflare Tunnel (for internet access)</h4>
<p class="mt-1">To publish apps and static sites on the internet, configure one wildcard route in the Cloudflare portal and keep all Local VSP admin pages local-only.</p>
<h4 class="font-bold text-white mt-4">Step 1: Prepare the domain in Cloudflare</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Add your domain to Cloudflare if it is not already managed there.</li>
	<li><strong>2.</strong> Confirm that the domain is active and using Cloudflare nameservers.</li>
	<li><strong>3.</strong> Open <a href="https://one.dash.cloudflare.com/" target="_blank" class="text-blue-400 underline">one.dash.cloudflare.com</a>.</li>
</ol>
<h4 class="font-bold text-white mt-4">Step 2: Create the tunnel</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Go to <strong>Networks → Tunnels</strong>.</li>
	<li><strong>2.</strong> Click <strong>Create a tunnel</strong>.</li>
	<li><strong>3.</strong> Choose <strong>Cloudflared</strong> as the connector type.</li>
	<li><strong>4.</strong> Give the tunnel a name, for example <code>LocalVSP</code>.</li>
	<li><strong>5.</strong> Copy the generated <strong>tunnel token</strong>.</li>
</ol>
<h4 class="font-bold text-white mt-4">Step 3: Configure the public hostname in Cloudflare</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> In the tunnel configuration, add a public hostname.</li>
	<li><strong>2.</strong> Use <code>*.yourdomain.com</code> as the hostname.</li>
	<li><strong>3.</strong> Set the service type to <strong>HTTP</strong>.</li>
	<li><strong>4.</strong> Set the origin service to <code>http://traefik:80</code>.</li>
	<li><strong>5.</strong> Save the hostname rule.</li>
</ol>
<div class="mt-4 rounded border border-amber-700 bg-amber-950/30 p-3 text-sm text-amber-200">
	<strong>Do not publish admin routes:</strong> do <strong>not</strong> add public hostnames for <code>manage.yourdomain.com</code>, <code>gitea.yourdomain.com</code>, or other admin pages. Local VSP is designed to keep those local-only.
</div>
<h4 class="font-bold text-white mt-4">Step 4: Save the settings in Local VSP</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Open <strong>Settings</strong> in Local VSP.</li>
	<li><strong>2.</strong> Paste the tunnel token into <strong>Tunnel Token</strong>.</li>
	<li><strong>3.</strong> Enter your base domain, for example <code>example.com</code>.</li>
	<li><strong>4.</strong> Save the settings.</li>
</ol>
<h4 class="font-bold text-white mt-4">Step 5: Publish individual apps or sites</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Set a subdomain on a Custom App or Static Website while it is stopped.</li>
	<li><strong>2.</strong> Deploy the app or site.</li>
	<li><strong>3.</strong> The app becomes reachable at <code>subdomain.example.com</code>.</li>
</ol>
<p class="mt-2 text-gray-400">Only apps and static sites with a configured subdomain are exposed through Cloudflare. The Local VSP dashboard, settings, help, and other management pages stay local-only.</p>
<h4 class="font-bold text-white mt-4">Core Services</h4>
<p>You can start, stop, restart or update individual platform services (Traefik, Gitea, Cloudflared, Management UI). This is useful after a configuration change or when an update is available.</p>
<p class="mt-1 text-gray-400">The settings page also shows version information and indicates when an update is available for pulled images.</p>`,

		"help.section.autostart.title": "Auto-Start / Keeping Apps Alive",
		"help.section.autostart.body": `<p>When <strong>Auto-Start</strong> is toggled <span class="text-green-400">on</span> for an app or website, the server will automatically start it again after a reboot or power outage.</p>
<p class="mt-2">Toggle it <span class="text-gray-400">off</span> if you want the app to stay stopped after a server restart (useful for temporary or test applications).</p>
<div class="mt-3 p-3 bg-green-900/30 border border-green-800 rounded text-sm">
  <strong class="text-green-400">Recommendation:</strong> Enable Auto-Start for all production apps and websites so nothing is accidentally lost after a power cycle.
</div>`,

		"help.section.troubleshoot.title": "Troubleshooting",
		"help.section.troubleshoot.body": `<h4 class="font-bold text-white">Deploy failed — "containers exited after start"</h4>
<p class="mt-1 text-gray-300">&rarr; The app was built successfully but crashed when it started. Open the <strong>Build Panel</strong> (click <strong>View Build</strong>) — the last 30 lines of container output are shown there. Common causes: missing environment variables, the app not listening on the expected port, or a syntax error in the code.</p>

<h4 class="font-bold text-white mt-4">App shows "Stopped" after deploy</h4>
<p class="mt-1 text-gray-300">&rarr; Press <strong>Deploy</strong> again and watch the Build Panel for errors. Make sure your app listens on the expected port: Node.js <strong>3000</strong>, Python <strong>8000</strong>, Go <strong>8080</strong>, Rust <strong>8080</strong>, PHP <strong>80</strong>. The app must bind to <code>0.0.0.0</code> (not 127.0.0.1).</p>

<h4 class="font-bold text-white mt-4">I can't see the server share in File Explorer</h4>
<p class="mt-1 text-gray-300">&rarr; Make sure your PC is on the same Wi-Fi or office network as the server. Try typing the address manually: <code class="text-green-400">\\&lt;server-ip&gt;</code></p>

<h4 class="font-bold text-white mt-4">The web address doesn't work from outside my network</h4>
<p class="mt-1 text-gray-300">&rarr; A Cloudflare Tunnel and Base Domain must be configured in Settings. See the <em>Configuring Settings</em> section above.</p>

<h4 class="font-bold text-white mt-4">Build is stuck or never finishes</h4>
<p class="mt-1 text-gray-300">&rarr; Builds have a 10-minute timeout. If the build panel shows no output, check that Docker is running on the server. You can restart the Management UI from the Settings page.</p>

<h4 class="font-bold text-white mt-4">The Management UI itself is not loading</h4>
<p class="mt-1 text-gray-300">&rarr; Check that the server is powered on. Connect via SSH or ask your IT contact to run <code>sudo docker compose -f /opt/localvsp/docker-compose.yml up -d</code> on the server.</p>`,

		// ---- Help page: Build & Deploy section ----
		"help.section.buildpanel.title": "Build & Deploy Panel",
		"help.section.buildpanel.body": `<p>When you press <strong>Deploy</strong> on a Custom App, the <strong>Build Panel</strong> slides open from the right side of the screen. It shows a live terminal view of the entire build and deploy process.</p>

<h4 class="font-bold text-white mt-3">What happens during deploy:</h4>
<ol class="space-y-2 mt-2">
  <li><strong>1.</strong> <strong>Auto-detect</strong> — If your project has a recognized trigger file (e.g. <code>package.json</code>, <code>requirements.txt</code>, <code>go.mod</code>, <code>composer.json</code>, <code>Cargo.toml</code>, or <code>index.html</code>) but no docker-compose.yml, a Dockerfile and docker-compose.yml are generated automatically.</li>
  <li><strong>2.</strong> <strong>Build</strong> — The Docker image is built from scratch (<code>docker compose build --no-cache</code>).</li>
  <li><strong>3.</strong> <strong>Start</strong> — The container is started in the background.</li>
  <li><strong>4.</strong> <strong>Health check</strong> — After 3 seconds, the system verifies the container is still running. If it crashed, the last 30 lines of container logs are captured and shown.</li>
</ol>

<h4 class="font-bold text-white mt-4">Multi-build support:</h4>
<p class="mt-1">You can deploy multiple apps at the same time. Each build gets its own tab at the top of the Build Panel. Switch between tabs to see each build's output.</p>

<h4 class="font-bold text-white mt-4">Persistent logs:</h4>
<p class="mt-1">Build output is saved to disk. Closing the browser, refreshing the page, or even restarting the server does <strong>not</strong> lose the build log. Click <strong>View Build</strong> on any app to reopen the last build output.</p>

<h4 class="font-bold text-white mt-4">Status indicators:</h4>
<ul class="mt-2 space-y-1">
  <li><span class="text-yellow-400 font-bold">&#9679;</span> <strong>Building</strong> — Build is in progress, output is streaming live.</li>
  <li><span class="text-green-400 font-bold">&#10003;</span> <strong>Success</strong> — Build completed and the app is running.</li>
  <li><span class="text-red-400 font-bold">&#10007;</span> <strong>Failed</strong> — Build or startup failed. Check the panel for error details.</li>
</ul>`,

		// ---- Help page: Developer Guide download ----
		"help.download.title":    "Developer Guide",
		"help.download.desc":     "Download the developer guide for writing and deploying apps on Local VSP.",
		"help.download.btn":      "Download vsp-skill.md",

	},

	// ── Dutch ─────────────────────────────────────────────────────────────────
	"nl": {
		// ---- Navigation ----
		"nav.brand":     "Local VSP",
		"nav.dashboard": "Dashboard",
		"nav.logs":      "Logboek",
		"nav.settings":  "Instellingen",
		"nav.help":      "Hulp",
		"nav.gitea":     "Gitea ↗",
		"dashboard.page.title":         "Local VSP Dashboard",
		"dashboard.hero.badge":         "Serverbedieningsvlak",
		"dashboard.hero.title":         "Beheer lokale apps, statische websites en platformservices vanuit één dashboard.",
		"dashboard.hero.body":          "Het dashboard is georganiseerd rond gezondheid, uitrolstatus en blootstelling. Primaire acties blijven vooraan; destructieve en netwerkacties blijven afgebakend.",
		"dashboard.hero.endpoint.title": "Endpoint",
		"dashboard.hero.endpoint.body":  "Management UI op het lokale netwerk",
		"dashboard.hero.state.title":    "Platformstatus",
		"dashboard.hero.state.metric":   "actieve containers",
		"dashboard.hero.state.body":     "Over Local VSP-kerndiensten en uitgerolde workloads",

		// ---- Shared ----
		"btn.refresh":   "↺ Vernieuwen",
		"btn.deploy":    "Uitrollen",
		"btn.redeploy":  "Opnieuw uitrollen",
		"btn.stop":      "Stoppen",
		"btn.restart":   "Herstarten",
		"btn.remove":    "Verwijderen",
		"btn.delete":    "Verwijderen",
		"btn.logs":      "Logboek",
		"btn.save":      "Instellingen opslaan",
		"btn.save_short": "Opslaan",
		"btn.start":     "Starten",
		"lbl.loading":   "Laden…",
		"lbl.working":   "bezig…",
		"lbl.status":    "Status",
		"lbl.actions":   "Acties",
		"lbl.autostart": "Automatisch starten",
		"lbl.webaddr":   "Webadres",
		"shared.help.title": "Hulp",
		"shared.help.more":  "Volledige documentatie",
		"notify.confirm.title": "Actie bevestigen",
		"notify.confirm.cancel": "Annuleren",
		"notify.confirm.confirm": "Bevestigen",
		"notify.confirm.type_label": "Typ de naam om de verwijdering te bevestigen",
		"notify.confirm.type_hint": "Dit verwijdert permanent alle bestanden en containers voor",
		"notify.confirm.type_placeholder": "Typ de exacte naam",
		"autostart.on.short": "Auto AAN",
		"autostart.off.short": "Auto UIT",
		"autostart.on.title": "Automatisch starten AAN — klik om uit te schakelen",
		"autostart.off.title": "Automatisch starten UIT — klik om in te schakelen",
		"row.path": "Pad",
		"row.exposure": "Publicatie",
		"row.domain": "Domein",
		"row.port": "Poort",
		"row.no_address": "Nog geen publiek adres geconfigureerd.",
		"row.webaddr.placeholder": "webadres",
		"confirm.project.stop": "%s stoppen?",
		"confirm.project.delete": "%s verwijderen? Dit stopt alle containers en verwijdert alle bestanden permanent.",
		"confirm.site.stop": "%s stoppen?",
		"confirm.site.delete": "%s verwijderen? Dit stopt de container en verwijdert alle bestanden permanent.",
		"confirm.infra.stop": "%s stoppen? De dienst wordt dan onbereikbaar.",
		"confirm.infra.rebuild": "%s opnieuw bouwen vanaf broncode? Dit veroorzaakt kort uitval.",
		"confirm.infra.update": "Nieuwste image voor %s ophalen en opnieuw aanmaken?",

		// ---- Build panel ----
		"build.panel.title":    "Bouwuitvoer",
		"build.status.building": "bezig met bouwen\u2026",
		"build.status.running": "bezig met bouwen\u2026",
		"build.status.success": "Bouw voltooid",
		"build.status.failed":  "Bouw mislukt",
		"build.btn.viewlog":    "Bouw bekijken",
		"build.lines":          "regels",

		// ---- Status badges ----
		"status.running": "Actief",
		"status.stopped": "Gestopt",
		"status.unknown": "Onbekend",

		// ---- System info bar ----
		"sysinfo.docker":     "Docker-versie",
		"sysinfo.containers": "Containers",
		"sysinfo.disk":       "Schijfgebruik",
		"sysinfo.load":       "Systeembelasting",
		"sysinfo.memory":     "Geheugengebruik",
		"sysinfo.running":    "actief",
		"sysinfo.total":      "totaal",

		// ---- Custom Apps section ----
		"apps.title": "Eigen applicaties",
		"apps.desc":  "Zet een projectmap (Node.js, Python, Go, PHP, Rust, HTML) of een docker-compose.yml in",
		"apps.col.project": "Project",
		"apps.row.subtitle": "Docker-projectmap",
		"apps.row.building.help": "Bouwuitvoer wordt in het zijpaneel gestreamd.",
		"apps.row.services.active": "services actief",
		"apps.row.running.help": "Project verwerkt verkeer.",
		"apps.row.status.partial": "Gedeeltelijk",
		"apps.row.partial.help": "Sommige services draaien, maar het project is nog niet volledig gezond.",
		"apps.row.status.ready": "Klaar voor auto-build",
		"apps.row.ready.help": "Er is een ondersteund project gedetecteerd dat automatisch kan worden gecontaineriseerd.",
		"apps.row.stopped.help": "Rol uit om het project te starten en adressen te publiceren.",
		"apps.empty.title": "Nog geen applicatiemappen gevonden.",
		"apps.empty.body": "Geen mappen gevonden in",
		"apps.empty.hint": "Voeg via de netwerkshare een submap toe met broncode of een docker-compose.yml.",
		"apps.help.title":  "Eigen applicaties — Hulp",
		"apps.help.body": `<p>Eigen applicaties zijn programma's die als Docker-containers op uw server draaien. Ze worden opgeslagen in de <strong>docker</strong>-netwerkshare.</p>
<h4>Ondersteunde projecttypen (autodetectie):</h4>
<ul>
<li><strong>Node.js</strong> — Plaats een map met een <code>package.json</code>. Een Dockerfile en docker-compose.yml worden automatisch aangemaakt (poort 3000).</li>
<li><strong>Python</strong> — Plaats een map met een <code>requirements.txt</code>. Automatisch aangemaakt met Python 3.11 (poort 8000, draait <code>app.py</code>).</li>
<li><strong>Go</strong> — Plaats een map met een <code>go.mod</code>. Multi-stage build met Go 1.22 (poort 8080).</li>
<li><strong>PHP</strong> — Plaats een map met een <code>composer.json</code>. Apache + PHP 8.3 met Composer-afhankelijkheden (poort 80).</li>
<li><strong>Rust</strong> — Plaats een map met een <code>Cargo.toml</code>. Multi-stage build met Rust 1.77 (poort 8080).</li>
<li><strong>Statische HTML</strong> — Plaats een map met een <code>index.html</code>. Geserveerd door Nginx (poort 80).</li>
<li><strong>Aangepast</strong> — Lever uw eigen <code>docker-compose.yml</code> (en eventueel een <code>Dockerfile</code>).</li>
</ul>
<h4>Hoe uitrollen:</h4>
<ol>
<li>Open Verkenner en navigeer naar <code>\\&lt;server-ip&gt;\vsp-home\docker</code>.</li>
<li>Kopieer uw projectmap daarheen.</li>
<li>Klik op deze pagina op <strong>Uitrollen</strong>. Het bouwpaneel schuift open met real-time terminaluitvoer.</li>
<li>Na het bouwen controleert het systeem of de app daadwerkelijk draait. Als de app crasht, toont het bouwpaneel de fout en de laatste 30 regels containeruitvoer.</li>
<li>Zodra de app actief is, verschijnt een klikbare link in de kolom <strong>Webadres</strong>.</li>
</ol>
<h4>Knoppen:</h4>
<ul>
<li><strong>Uitrollen</strong> — Bouwt en start de app (of herbouwt indien al uitgerold).</li>
<li><strong>Stoppen</strong> — Stopt de app zonder bestanden te verwijderen.</li>
<li><strong>Bouw bekijken</strong> — Opent het bouwpaneel met de laatste bouwuitvoer.</li>
<li><strong>Automatisch starten</strong> — Wanneer aan, herstart de app automatisch na een serverherstart.</li>
</ul>`,

		// ---- Static Websites section ----
		"html.title": "Statische websites",
		"html.desc":  "Zet een map met statische bestanden in",
		"html.desc2": "— automatisch geserveerd door Traefik",
		"html.col.site": "Website",
		"html.row.subtitle": "Map met statische website",
		"html.row.running.help": "De statische site is gepubliceerd en bereikbaar via poort of domein.",
		"html.row.stopped.help": "Rol uit om de site te publiceren en een adres toe te wijzen.",
		"html.empty.title": "Nog geen statische websites gevonden.",
		"html.empty.body": "Geen mappen gevonden in",
		"html.empty.hint": "Voeg via de netwerkshare een submap toe en plaats daarin uw index.html.",
		"html.help.title": "Statische websites — Hulp",
		"html.help.body": `<p>Statische websites zijn eenvoudige op mappen gebaseerde HTML-sites die via een Nginx-container worden geserveerd — geen programmeerkennis vereist.</p>
<h4>Hoe publiceert u een website:</h4>
<ol>
<li>Open Verkenner en ga naar <code>\\&lt;server-ip&gt;\vsp-home\html</code>.</li>
<li>Maak een nieuwe map aan met uw websitenaam (bijv. <code>mijn-site</code>).</li>
<li>Plaats uw <code>index.html</code> en eventuele afbeeldingen / CSS-bestanden in die map.</li>
<li>Klik op de dashboardpagina op <strong>Uitrollen</strong> naast uw sitenaam.</li>
<li>Er wordt automatisch een poort toegewezen (bereik 8200–8299) en de site gaat live.</li>
</ol>
<h4>Subdomein-toegang:</h4>
<p>U kunt optioneel een subdomein toewijzen (bijv. <code>mijn-site.uwdomein.nl</code>) door het in te typen in het veld <strong>Webadres</strong> wanneer de site gestopt is, en vervolgens uit te rollen.</p>
<h4>Knoppen:</h4>
<ul>
<li><strong>Uitrollen</strong> — Maakt de site live (of opnieuw uitrollen na bestandswijzigingen).</li>
<li><strong>Stoppen</strong> — Haalt de site offline (bestanden worden niet verwijderd).</li>
<li><strong>Automatisch starten</strong> — Wanneer aan, herstart de site automatisch na een serverherstart.</li>
</ul>`,

		// ---- Settings: Core Services ----
		"settings.infra.title": "Kerndiensten",
		"settings.infra.desc":  "Kern-VSP-services — start, stop, herstart of update elk onderdeel.",
		"settings.infra.col.service": "Service",
		"settings.infra.col.version": "Versie",
		"settings.infra.col.port":    "Poort",
		"settings.infra.not_configured": "niet geconfigureerd",
		"settings.infra.inactive": "inactief",
		"settings.infra.configure": "Configureren",
		"settings.infra.rebuild": "Opnieuw bouwen",
		"settings.infra.update": "Bijwerken",
		"settings.infra.local_build": "lokale build",
		"settings.infra.update_available": "update beschikbaar",
		"settings.infra.up_to_date": "up-to-date",
		"settings.infra.help.title":  "Kerndiensten — Hulp",
		"settings.infra.help.body": `<p>Dit zijn de ingebouwde services die het Local VSP-platform zelf aansturen. Normaal gesproken hoeft u hier niets aan te wijzigen.</p>
<ul>
<li><strong>Traefik</strong> — De reverse-proxy die het verkeer op basis van domeinnaam naar elke app stuurt.</li>
<li><strong>Gitea</strong> — Een zelf-gehoste Git-server waar u uw broncode kunt opslaan.</li>
<li><strong>Management UI</strong> — De webinterface die u nu gebruikt.</li>
</ul>
<h4>Knoppen:</h4>
<ul>
<li><strong>Herstarten</strong> — Herstart de service snel zonder gegevens te verliezen.</li>
<li><strong>Updaten</strong> — Haalt de nieuwste versie op en maakt de container opnieuw aan.</li>
</ul>
<p class="text-yellow-300">⚠ Bij herstarten van de Management UI verliest uw browser kort de verbinding. Ververs na ~10 seconden.</p>`,

		// ---- Settings: Platform Settings ----
		"settings.platform.title":     "Platforminstellingen",
		"settings.platform.desc":      "Wijzigingen worden opgeslagen in /opt/localvsp/.env.",
		"settings.cf.title":           "Cloudflare-tunnel",
		"settings.cf.token.label":     "Tunneltoken",
		"settings.cf.token.hint":      "→ Netwerken → Tunnels → Tunnel aanmaken. Opslaan activeert de tunnel automatisch.",
		"settings.cf.clear":           "Opgeslagen tunneltoken wissen bij opslaan",
		"settings.domain.title":       "Domein",
		"settings.domain.label":       "Basisdomein",
		"settings.domain.hint":        "Subdomeinen worden:",
		"settings.domain.clear":       "Opgeslagen basisdomein wissen bij opslaan",
		"settings.hero.badge":         "Platformconfiguratie",
		"settings.hero.title":         "Beheer Local VSP-infrastructuur en publicatie-instellingen.",
		"settings.hero.body":          "Infrastructuuracties staan los van Cloudflare- en domeinconfiguratie, zodat servicebediening en routingwijzigingen beter beheersbaar blijven.",
		"settings.summary.domain.label": "Basisdomein",
		"settings.summary.domain.empty": "niet geconfigureerd",
		"settings.summary.token.label":  "Tunneltoken",
		"settings.summary.token.present": "geconfigureerd",
		"settings.summary.token.missing": "ontbreekt",
		"settings.effects.title":      "Effect van deze instellingen",
		"settings.effects.body":       "Tunnel- en domeinwijzigingen bepalen hoe projectsubdomeinen buiten uw lokale netwerk bereikbaar worden.",
		"settings.effects.local":      "Zonder domein: lokale poortlinks blijven werken.",
		"settings.effects.domain":     "Alleen met een domein: subdomeinlabels kunnen worden gegenereerd, maar internettoegang hangt nog af van uw tunnel- en DNS-configuratie.",
		"settings.effects.tunnel":     "Met beide ingesteld: het dashboard kan apps publiceren op benoemde hostnamen.",
		"settings.platform.help.title": "Platforminstellingen — Hulp",
		"settings.platform.help.body": `<p>Deze instellingen bepalen hoe Local VSP apps en websites via Cloudflare met het internet verbindt.</p>
<h4>Cloudflare-tunneltoken:</h4>
<ol>
<li>Ga naar <a href="https://one.dash.cloudflare.com/" target="_blank" class="text-blue-400 underline">one.dash.cloudflare.com</a> en log in.</li>
<li>Open <strong>Netwerken → Tunnels</strong> en maak een nieuwe tunnel aan.</li>
<li>Kies <strong>Cloudflared</strong> als connectortype en geef de tunnel een naam, bijvoorbeeld <code>LocalVSP</code>.</li>
<li>Kopieer op de connectorpagina de lange tunneltokenreeks.</li>
<li>Plak die token in Local VSP onder <strong>Tunneltoken</strong> en sla op.</li>
</ol>
<h4>Basisdomein:</h4>
<p>Voer het domein in dat u bezit en op Cloudflare heeft ingesteld (bijv. <code>voorbeeld.nl</code>). Alleen apps en websites waaraan u een subdomein toewijst, worden dan bereikbaar op <code>subdomein.voorbeeld.nl</code>. Het Local VSP-dashboard en andere beheerpaden blijven lokaal.</p>`,

		// ---- Logs page ----
		"logs.title":       "Logboek",
		"logs.lines.label": "Regels:",
		"logs.autorefresh": "Automatisch vernieuwen (5s)",
		"logs.output":      "Uitvoer",
		"logs.loading":     "Logboek laden...",
		"logs.back":        "← Dashboard",

		// ---- Help page ----
		"help.page.title":    "Hulp & Documentatie",
		"help.page.subtitle": "Stap-voor-stap handleidingen voor het gebruik van Local VSP",
		"help.hero.badge":    "Beheerhandleiding",
		"help.hero.summary":  "Gebruik deze pagina als operationele referentie voor lokale toegang, uitrollen, bouwuitvoer en Cloudflare-publicatie. De cloudtunnel publiceert alleen apps en sites waarvoor u bewust een subdomein instelt.",
		"help.hero.note.title": "Belangrijke routingregel",
		"help.hero.note.body":  "Management UI, Instellingen, Hulp, Gitea en andere beheerpagina's blijven lokaal. Alleen apps en sites met een toegewezen subdomein worden via Cloudflare gepubliceerd.",
		"help.hero.public.title": "Publieke toegang",
		"help.hero.public.body":  "Stel in Cloudflare één wildcard-hostnaam in: *.uwdomein.nl → http://traefik:80. Ken daarna per app/site subdomeinen toe in Local VSP.",
		"help.hero.admin.title": "Beheeroppervlak",
		"help.hero.admin.body":  "Gebruik de lokale netwerkadressen voor Management UI en Gitea. Maak er geen publieke Cloudflare-hostnamen voor aan.",
		"help.toc":           "Inhoudsopgave",
		"help.section.overview.title": "Wat is Local VSP?",
		"help.section.overview.body": `<p>Local VSP (Virtual Server Platform) is een zelf-gehost serverbeheerssysteem waarmee u webapplicaties, websites en diensten kunt draaien op een lokale server — zonder technische kennis.</p>
<p class="mt-2">Alles wordt beheerd via deze webinterface. Geen commandoregel nodig.</p>
<h4 class="mt-4 font-bold text-white">Belangrijkste functies:</h4>
<ul class="mt-2 space-y-1">
  <li><strong class="text-green-400">Autodetectie</strong> — Plaats een Node.js-, Python-, Go-, PHP-, Rust- of statische HTML-projectmap op de server en Local VSP genereert automatisch een Dockerfile en docker-compose.yml.</li>
  <li><strong class="text-green-400">Bouwpaneel</strong> — Real-time terminaluitvoer bij het uitrollen — streamt de bouwvoortgang live en controleert of de app daadwerkelijk draait.</li>
  <li><strong class="text-green-400">Eigen applicaties</strong> — Broncodemappen die u op de server plaatst, gebouwd en gedraaid als Docker-containers.</li>
  <li><strong class="text-green-400">Statische websites</strong> — Eenvoudige HTML/CSS-mappen gepubliceerd als websites.</li>
  <li><strong class="text-green-400">Kerndiensten</strong> — Ingebouwde platformservices: Traefik (reverse proxy), Gitea (Git-hosting), Cloudflare-tunnel (internettoegang).</li>
  <li><strong class="text-green-400">Persistente logs</strong> — Bouwlogboeken worden op schijf opgeslagen en overleven paginaverversingen, tabsluitingen en serverherstarts.</li>
</ul>`,

		"help.section.access.title": "Toegang tot het dashboard",
		"help.section.access.body": `<ol class="space-y-2">
  <li><strong>1.</strong> Zorg dat u op hetzelfde netwerk zit als de server (thuis-wifi of kantoor-LAN).</li>
  <li><strong>2.</strong> Open een webbrowser (Chrome, Firefox of Edge).</li>
  <li><strong>3.</strong> Typ het adres van de server in de adresbalk, bijv.: <code class="text-green-400">http://192.168.1.100:8080</code></li>
  <li><strong>4.</strong> Het Local VSP-dashboard verschijnt.</li>
</ol>
<p class="mt-3 text-gray-400">Tip: Vraag uw IT-contactpersoon om het exacte IP-adres van uw server.</p>`,

		"help.section.customapps.title": "Een eigen applicatie publiceren",
		"help.section.customapps.body": `<ol class="space-y-3">
  <li><strong>Stap 1.</strong> Open <strong>Verkenner</strong> op uw Windows-pc.</li>
  <li><strong>Stap 2.</strong> Typ in de adresbalk <code class="text-green-400">\\&lt;server-ip&gt;\vsp-home\docker</code> en druk op Enter.</li>
  <li><strong>Stap 3.</strong> Kopieer of sleep uw applicatiemap naar deze netwerklocatie.</li>
  <li><strong>Stap 4.</strong> Ga terug naar het dashboard in uw browser.</li>
  <li><strong>Stap 5.</strong> Uw appmap verschijnt in de tabel <strong>Eigen applicaties</strong>. Klik op <strong>Uitrollen</strong>.</li>
  <li><strong>Stap 6.</strong> Het <strong>Bouwpaneel</strong> schuift open aan de rechterkant met real-time bouwuitvoer. U kunt het dashboard blijven gebruiken terwijl het bouwt.</li>
  <li><strong>Stap 7.</strong> Na het bouwen wacht het systeem 3 seconden en controleert of uw app daadwerkelijk draait. Als de app crasht, wordt het paneel rood en toont het de laatste 30 regels containeruitvoer.</li>
  <li><strong>Stap 8.</strong> Bij succes klikt u op de link in de kolom <strong>Webadres</strong> om uw app te openen.</li>
</ol>
<div class="mt-4 p-3 bg-blue-900/40 border border-blue-700 rounded text-sm">
  <strong class="text-blue-300">Autodetectie:</strong>
  <p class="mt-1 text-gray-300">U hebt <strong>geen</strong> Dockerfile of docker-compose.yml nodig voor ondersteunde projecttypen. Het systeem detecteert uw project automatisch:</p>
  <ul class="mt-1 text-gray-300">
    <li><code>package.json</code> → Node.js (poort 3000)</li>
    <li><code>requirements.txt</code> → Python (poort 8000)</li>
    <li><code>go.mod</code> → Go (poort 8080)</li>
    <li><code>composer.json</code> → PHP (poort 80)</li>
    <li><code>Cargo.toml</code> → Rust (poort 8080)</li>
    <li><code>index.html</code> → Statische HTML (poort 80)</li>
  </ul>
  <p class="mt-1 text-gray-300">Voor andere talen of aangepaste configuraties levert u uw eigen <code>docker-compose.yml</code>.</p>
</div>
<div class="mt-3 p-3 bg-green-900/30 border border-green-800 rounded text-sm">
  <strong class="text-green-400">Meerdere builds:</strong>
  <p class="mt-1 text-gray-300">U kunt meerdere apps tegelijk uitrollen. Elke build krijgt een eigen tabblad in het Bouwpaneel. Het sluiten van de browser annuleert een lopende build niet — verbind later opnieuw om het resultaat te zien.</p>
</div>`,

		"help.section.staticsite.title": "Een statische website publiceren",
		"help.section.staticsite.body": `<ol class="space-y-3">
  <li><strong>Stap 1.</strong> Open <strong>Verkenner</strong> op uw Windows-pc.</li>
  <li><strong>Stap 2.</strong> Navigeer naar <code class="text-green-400">\\&lt;server-ip&gt;\vsp-home\html</code>.</li>
  <li><strong>Stap 3.</strong> Maak een nieuwe map aan voor uw website (bijv. <code>bedrijfs-site</code>).</li>
  <li><strong>Stap 4.</strong> Kopieer uw websitebestanden (<code>index.html</code>, afbeeldingen, CSS) naar die map.</li>
  <li><strong>Stap 5.</strong> Zoek op het dashboard uw site in de tabel <strong>Statische websites</strong> en klik op <strong>Uitrollen</strong>.</li>
  <li><strong>Stap 6.</strong> Er wordt automatisch een poort toegewezen (bereik 8200–8299) en de site gaat live. Klik op de link in de kolom <strong>Webadres</strong>.</li>
</ol>
<div class="mt-4 p-3 bg-yellow-900/40 border border-yellow-700 rounded text-sm">
  <strong class="text-yellow-300">Belangrijk:</strong> Elke statische websitemap moet een bestand bevatten met de exacte naam <code>index.html</code>. Dit is de eerste pagina die bezoekers zien.
</div>
<div class="mt-3 p-3 bg-blue-900/40 border border-blue-700 rounded text-sm">
  <strong class="text-blue-300">Subdomein-toegang:</strong>
  <p class="mt-1 text-gray-300">U kunt optioneel een subdomein invullen in het veld Webadres (wanneer de site gestopt is) en vervolgens uitrollen. Als een Cloudflare-tunnel en basisdomein zijn geconfigureerd, is de site ook bereikbaar op <code>subdomein.uwdomein.nl</code>.</p>
</div>`,

		"help.section.logs.title": "Logboeken bekijken",
		"help.section.logs.body": `<p>Local VSP heeft twee soorten logboeken:</p>
<h4 class="font-bold text-white mt-2">Bouwlogboeken (tijdens uitrol)</h4>
<p class="mt-1">Wanneer u op <strong>Uitrollen</strong> klikt, toont het Bouwpaneel live bouwuitvoer. Dit logboek wordt op schijf opgeslagen — u kunt het later heropenen door op <strong>Bouw bekijken</strong> te klikken naast de app. Bouwlogboeken overleven paginaverversingen, browsersluitingen en zelfs serverherstarts.</p>
<h4 class="font-bold text-white mt-3">Runtime-logboeken (containeruitvoer)</h4>
<p class="mt-1">Zodra een app draait, kunt u de live uitvoer bekijken:</p>
<ol class="space-y-2 mt-2">
  <li><strong>1.</strong> Klik in de tabel <strong>Eigen applicaties</strong> op de link <strong>Logboek</strong> naast de app.</li>
  <li><strong>2.</strong> De logboekviewer opent met de laatste containeruitvoer.</li>
  <li><strong>3.</strong> Standaard ververst deze elke 5 seconden automatisch.</li>
  <li><strong>4.</strong> Gebruik de vervolgkeuzelijst <strong>Regels</strong> om meer of minder regels te zien (50, 100, 200 of 500).</li>
  <li><strong>5.</strong> Zet <strong>Automatisch vernieuwen</strong> uit als u de weergave wilt pauzeren om een bepaald gedeelte te lezen.</li>
</ol>`,

		"help.section.settings.title": "Instellingen configureren",
		"help.section.settings.body": `<h4 class="font-bold text-white">Cloudflare-tunnel (voor internettoegang)</h4>
<p class="mt-1">Om apps en statische websites op internet te publiceren, configureert u in het Cloudflare-portaal één wildcardroute en laat u alle Local VSP-beheerpagina's lokaal.</p>
<h4 class="font-bold text-white mt-4">Stap 1: Domein voorbereiden in Cloudflare</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Voeg uw domein toe aan Cloudflare als het daar nog niet wordt beheerd.</li>
	<li><strong>2.</strong> Controleer dat het domein actief is en de Cloudflare-nameservers gebruikt.</li>
	<li><strong>3.</strong> Open <a href="https://one.dash.cloudflare.com/" target="_blank" class="text-blue-400 underline">one.dash.cloudflare.com</a>.</li>
</ol>
<h4 class="font-bold text-white mt-4">Stap 2: Tunnel aanmaken</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Ga naar <strong>Netwerken → Tunnels</strong>.</li>
	<li><strong>2.</strong> Klik op <strong>Tunnel aanmaken</strong>.</li>
	<li><strong>3.</strong> Kies <strong>Cloudflared</strong> als connectortype.</li>
	<li><strong>4.</strong> Geef de tunnel een naam, bijvoorbeeld <code>LocalVSP</code>.</li>
	<li><strong>5.</strong> Kopieer het gegenereerde <strong>tunneltoken</strong>.</li>
</ol>
<h4 class="font-bold text-white mt-4">Stap 3: Publieke hostnaam configureren in Cloudflare</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Voeg in de tunnelconfiguratie een publieke hostnaam toe.</li>
	<li><strong>2.</strong> Gebruik <code>*.uwdomein.nl</code> als hostnaam.</li>
	<li><strong>3.</strong> Stel het servicetype in op <strong>HTTP</strong>.</li>
	<li><strong>4.</strong> Stel de originservice in op <code>http://traefik:80</code>.</li>
	<li><strong>5.</strong> Sla de hostnaamregel op.</li>
</ol>
<div class="mt-4 rounded border border-amber-700 bg-amber-950/30 p-3 text-sm text-amber-200">
	<strong>Publiceer geen beheerpaden:</strong> maak <strong>geen</strong> publieke hostnamen aan voor <code>manage.uwdomein.nl</code>, <code>gitea.uwdomein.nl</code> of andere beheerpagina's. Local VSP is ontworpen om die lokaal te houden.
</div>
<h4 class="font-bold text-white mt-4">Stap 4: Instellingen opslaan in Local VSP</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Open <strong>Instellingen</strong> in Local VSP.</li>
	<li><strong>2.</strong> Plak het tunneltoken in <strong>Tunneltoken</strong>.</li>
	<li><strong>3.</strong> Voer uw basisdomein in, bijvoorbeeld <code>voorbeeld.nl</code>.</li>
	<li><strong>4.</strong> Sla de instellingen op.</li>
</ol>
<h4 class="font-bold text-white mt-4">Stap 5: Individuele apps of sites publiceren</h4>
<ol class="space-y-2 mt-2">
	<li><strong>1.</strong> Stel een subdomein in op een Eigen applicatie of Statische website terwijl die gestopt is.</li>
	<li><strong>2.</strong> Rol de app of site uit.</li>
	<li><strong>3.</strong> De app wordt bereikbaar op <code>subdomein.voorbeeld.nl</code>.</li>
</ol>
<p class="mt-2 text-gray-400">Alleen apps en statische sites met een geconfigureerd subdomein worden via Cloudflare gepubliceerd. Het Local VSP-dashboard, instellingen, help en andere beheerpagina's blijven lokaal.</p>
<h4 class="font-bold text-white mt-4">Kerndiensten</h4>
<p>U kunt individuele platformservices starten, stoppen, herstarten of updaten (Traefik, Gitea, Cloudflared, Management UI). Dit is handig na een configuratiewijziging of wanneer een update beschikbaar is.</p>
<p class="mt-1 text-gray-400">De instellingenpagina toont ook versie-informatie en geeft aan wanneer een update beschikbaar is voor externe images.</p>`,

		"help.section.autostart.title": "Automatisch starten / Apps actief houden",
		"help.section.autostart.body": `<p>Wanneer <strong>Automatisch starten</strong> is ingeschakeld (<span class="text-green-400">aan</span>) voor een app of website, start de server deze automatisch opnieuw na een herstart of stroomstoring.</p>
<p class="mt-2">Schakel het <span class="text-gray-400">uit</span> als u wilt dat de app gestopt blijft na een serverherstart (handig voor tijdelijke of testapplicaties).</p>
<div class="mt-3 p-3 bg-green-900/30 border border-green-800 rounded text-sm">
  <strong class="text-green-400">Aanbeveling:</strong> Schakel Automatisch starten in voor alle productie-apps en -websites, zodat niets per ongeluk verloren gaat na een stroomstoring.
</div>`,

		"help.section.troubleshoot.title": "Probleemoplossing",
		"help.section.troubleshoot.body": `<h4 class="font-bold text-white">Uitrol mislukt — "containers gestopt na start"</h4>
<p class="mt-1 text-gray-300">&rarr; De app is succesvol gebouwd maar crashte bij het opstarten. Open het <strong>Bouwpaneel</strong> (klik op <strong>Bouw bekijken</strong>) — de laatste 30 regels containeruitvoer worden daar getoond. Veelvoorkomende oorzaken: ontbrekende omgevingsvariabelen, de app luistert niet op de verwachte poort, of een syntaxfout in de code.</p>

<h4 class="font-bold text-white mt-4">App toont "Gestopt" na uitrol</h4>
<p class="mt-1 text-gray-300">&rarr; Klik opnieuw op <strong>Uitrollen</strong> en bekijk het Bouwpaneel op fouten. Zorg dat uw app luistert op de verwachte poort: Node.js <strong>3000</strong>, Python <strong>8000</strong>, Go <strong>8080</strong>, Rust <strong>8080</strong>, PHP <strong>80</strong>. De app moet binden aan <code>0.0.0.0</code> (niet 127.0.0.1).</p>

<h4 class="font-bold text-white mt-4">Ik zie de servermap niet in Verkenner</h4>
<p class="mt-1 text-gray-300">&rarr; Controleer of uw pc op hetzelfde wifi of kantoornetwerk zit als de server. Probeer het adres handmatig in te typen: <code class="text-green-400">\\&lt;server-ip&gt;</code></p>

<h4 class="font-bold text-white mt-4">Het webadres werkt niet buiten mijn netwerk</h4>
<p class="mt-1 text-gray-300">&rarr; Er moeten een Cloudflare-tunnel en basisdomein geconfigureerd zijn in de Instellingen. Zie het gedeelte <em>Instellingen configureren</em> hierboven.</p>

<h4 class="font-bold text-white mt-4">Build loopt vast of wordt nooit afgerond</h4>
<p class="mt-1 text-gray-300">&rarr; Builds hebben een time-out van 10 minuten. Als het bouwpaneel geen uitvoer toont, controleer of Docker draait op de server. U kunt de Management UI herstarten via de Instellingenpagina.</p>

<h4 class="font-bold text-white mt-4">De Management UI laadt zelf niet</h4>
<p class="mt-1 text-gray-300">&rarr; Controleer of de server aan staat. Verbind via SSH of vraag uw IT-contactpersoon om op de server het commando <code>sudo docker compose -f /opt/localvsp/docker-compose.yml up -d</code> uit te voeren.</p>`,

		// ---- Help page: Build & Deploy section ----
		"help.section.buildpanel.title": "Bouw- en uitrolpaneel",
		"help.section.buildpanel.body": `<p>Wanneer u op <strong>Uitrollen</strong> klikt bij een Eigen applicatie, schuift het <strong>Bouwpaneel</strong> open aan de rechterkant van het scherm. Het toont een live terminalweergave van het volledige bouw- en uitrolproces.</p>

<h4 class="font-bold text-white mt-3">Wat er gebeurt tijdens uitrol:</h4>
<ol class="space-y-2 mt-2">
  <li><strong>1.</strong> <strong>Autodetectie</strong> — Als uw project een herkend triggerbestand heeft (bijv. <code>package.json</code>, <code>requirements.txt</code>, <code>go.mod</code>, <code>composer.json</code>, <code>Cargo.toml</code> of <code>index.html</code>) maar geen docker-compose.yml, worden een Dockerfile en docker-compose.yml automatisch gegenereerd.</li>
  <li><strong>2.</strong> <strong>Bouwen</strong> — De Docker-image wordt vanaf nul opgebouwd (<code>docker compose build --no-cache</code>).</li>
  <li><strong>3.</strong> <strong>Starten</strong> — De container wordt op de achtergrond gestart.</li>
  <li><strong>4.</strong> <strong>Gezondheidscontrole</strong> — Na 3 seconden controleert het systeem of de container nog draait. Als deze is gecrasht, worden de laatste 30 regels containerlogboek vastgelegd en getoond.</li>
</ol>

<h4 class="font-bold text-white mt-4">Meerdere builds:</h4>
<p class="mt-1">U kunt meerdere apps tegelijk uitrollen. Elke build krijgt een eigen tabblad bovenaan het Bouwpaneel. Schakel tussen tabbladen om de uitvoer van elke build te bekijken.</p>

<h4 class="font-bold text-white mt-4">Persistente logboeken:</h4>
<p class="mt-1">Bouwuitvoer wordt op schijf opgeslagen. Het sluiten van de browser, vernieuwen van de pagina of zelfs herstarten van de server verliest het bouwlogboek <strong>niet</strong>. Klik op <strong>Bouw bekijken</strong> bij een app om de laatste bouwuitvoer te heropenen.</p>

<h4 class="font-bold text-white mt-4">Statusindicatoren:</h4>
<ul class="mt-2 space-y-1">
  <li><span class="text-yellow-400 font-bold">&#9679;</span> <strong>Bezig met bouwen</strong> — Build is bezig, uitvoer wordt live gestreamd.</li>
  <li><span class="text-green-400 font-bold">&#10003;</span> <strong>Gelukt</strong> — Build voltooid en de app draait.</li>
  <li><span class="text-red-400 font-bold">&#10007;</span> <strong>Mislukt</strong> — Build of opstart mislukt. Bekijk het paneel voor foutdetails.</li>
</ul>`,

		// ---- Help page: Developer Guide download ----
		"help.download.title":    "Ontwikkelaarshandleiding",
		"help.download.desc":     "Download de ontwikkelaarshandleiding voor het schrijven en uitrollen van apps op Local VSP.",
		"help.download.btn":      "Download vsp-skill.md",
	},
}
