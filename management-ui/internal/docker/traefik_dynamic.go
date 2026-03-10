package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type traefikRoute struct {
	RouterName  string
	Rule        string
	ServiceName string
	TargetURL   string
}

func traefikDynamicConfigPath() string {
	return "/opt/localvsp/data/traefik/routes.yml"
}

func SyncTraefikDynamicConfig() error {
	content := renderTraefikDynamicConfig(collectTraefikRoutes(GetCurrentConfig().Domain, LoadAutoStartConfig()))
	return writeFileAtomic(traefikDynamicConfigPath(), []byte(content), 0644)
}

func collectTraefikRoutes(domain string, autoCfg AutoStartConfig) []traefikRoute {
	var routes []traefikRoute

	dockerNames := make([]string, 0, len(autoCfg.DockerProjects))
	for name := range autoCfg.DockerProjects {
		dockerNames = append(dockerNames, name)
	}
	sort.Strings(dockerNames)
	for _, name := range dockerNames {
		pc := autoCfg.DockerProjects[name]
		if !shouldExposeThroughTraefik(pc.Subdomain, domain) {
			continue
		}
		dir := filepath.Join(homeVSP, "docker", name)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		if len(composeContainerIDs(dir)) == 0 {
			continue
		}
		serviceName := getFirstComposeService(dir)
		if serviceName == "" {
			continue
		}
		exposePort := pc.ExposePort
		if exposePort == "" {
			exposePort = DetectInternalPort(dir)
		}
		routerName := traefikRouterName("vsp", name)
		routes = append(routes, traefikRoute{
			RouterName:  routerName,
			Rule:        fmt.Sprintf("Host(`%s.%s`)", pc.Subdomain, domain),
			ServiceName: routerName,
			TargetURL:   fmt.Sprintf("http://%s:%s", serviceName, exposePort),
		})
	}

	htmlNames := make([]string, 0, len(autoCfg.HtmlSites))
	for name := range autoCfg.HtmlSites {
		htmlNames = append(htmlNames, name)
	}
	sort.Strings(htmlNames)
	for _, name := range htmlNames {
		sc := autoCfg.HtmlSites[name]
		if !shouldExposeThroughTraefik(sc.Subdomain, domain) {
			continue
		}
		containerName := "html-" + name
		statusOut, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", "{{.State.Status}}")
		if err != nil || strings.TrimSpace(string(statusOut)) != "running" {
			continue
		}
		routerName := traefikRouterName("html", name)
		routes = append(routes, traefikRoute{
			RouterName:  routerName,
			Rule:        fmt.Sprintf("Host(`%s.%s`)", sc.Subdomain, domain),
			ServiceName: routerName,
			TargetURL:   fmt.Sprintf("http://%s:80", containerName),
		})
	}

	return routes
}

func renderTraefikDynamicConfig(routes []traefikRoute) string {
	var sb strings.Builder
	sb.WriteString("http:\n")
	if len(routes) == 0 {
		sb.WriteString("  routers: {}\n")
		sb.WriteString("  services: {}\n")
		return sb.String()
	}

	sb.WriteString("  routers:\n")
	for _, route := range routes {
		sb.WriteString(fmt.Sprintf("    %s:\n", route.RouterName))
		sb.WriteString("      entryPoints:\n")
		sb.WriteString("        - web\n")
		sb.WriteString(fmt.Sprintf("      rule: %q\n", route.Rule))
		sb.WriteString(fmt.Sprintf("      service: %s\n", route.ServiceName))
	}

	sb.WriteString("  services:\n")
	for _, route := range routes {
		sb.WriteString(fmt.Sprintf("    %s:\n", route.ServiceName))
		sb.WriteString("      loadBalancer:\n")
		sb.WriteString("        servers:\n")
		sb.WriteString(fmt.Sprintf("          - url: %q\n", route.TargetURL))
	}

	return sb.String()
}