package handlers

import "net/http"

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", DashboardHandler)
	mux.HandleFunc("/logs", LogsPageHandler)
	mux.HandleFunc("/settings", SettingsPageHandler)
	mux.HandleFunc("/help", HelpHandler)

	mux.HandleFunc("/infra/action", InfraActionHandler)

	mux.HandleFunc("/api/logs", LogsStreamHandler)
	mux.HandleFunc("/api/infra", InfraTableHandler)
	mux.HandleFunc("/api/local-html", LocalHtmlSitesHandler)
	mux.HandleFunc("/api/local-docker", LocalDockerProjectsHandler)
	mux.HandleFunc("/api/builds", BuildsAPIHandler)
	mux.HandleFunc("/api/build-log", BuildLogViewHandler)

	mux.HandleFunc("/local/docker/build-stream", LocalDockerBuildStreamHandler)
	mux.HandleFunc("/local/html/action", LocalHtmlActionHandler)
	mux.HandleFunc("/local/docker/action", LocalDockerActionHandler)
	mux.HandleFunc("/local/html/autostart", LocalHtmlAutoStartHandler)
	mux.HandleFunc("/local/docker/autostart", LocalDockerAutoStartHandler)
	mux.HandleFunc("/local/html/subdomain", LocalHtmlSubdomainHandler)
	mux.HandleFunc("/local/docker/subdomain", LocalDockerSubdomainHandler)
	mux.HandleFunc("/settings/save", SaveSettingsHandler)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	return mux
}