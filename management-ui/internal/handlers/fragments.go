package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/localvsp/management-ui/internal/docker"
)

func InfraTableHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "infra-rows", tplData(requestLang(w, r), map[string]interface{}{
		"Infra": docker.GetInfraStatus(),
		"Host":  hostFromRequest(r),
	}))
}

func LogsStreamHandler(w http.ResponseWriter, r *http.Request) {
	appName := sanitizeName(r.URL.Query().Get("app"))
	if appName == "" || appName == "." || appName == "/" {
		respondError(w, "App name is required.")
		return
	}

	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "100"
	}

	logOutput, err := docker.ComposeLogs(filepath.Join("/vsp-home/docker", appName), lines)
	if err != nil {
		respondError(w, "Could not fetch logs: "+err.Error())
		return
	}

	fmt.Fprintf(w, "<pre class=\"text-xs text-green-400 whitespace-pre-wrap break-all\">%s</pre>", template.HTMLEscapeString(logOutput))
}

func LocalHtmlSitesHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "local-html-rows", htmlRowMap(w, r))
}

func LocalDockerProjectsHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "local-docker-rows", dockerRowMap(w, r))
}

func htmlRowMap(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	return tplData(requestLang(w, r), map[string]interface{}{
		"Sites":  docker.ListHtmlSites(),
		"Host":   hostFromRequest(r),
		"Domain": docker.GetCurrentConfig().Domain,
	})
}

func dockerRowMap(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	return tplData(requestLang(w, r), map[string]interface{}{
		"Projects":     docker.ListDockerProjects(),
		"Domain":       docker.GetCurrentConfig().Domain,
		"Host":         hostFromRequest(r),
		"ActiveBuilds": getActiveBuilds(),
	})
}