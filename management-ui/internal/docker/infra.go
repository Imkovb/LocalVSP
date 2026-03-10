package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type InfraContainer struct {
	Name            string
	ServiceName     string
	Status          string
	Port            string
	Image           string
	ImageVersion    string
	LatestVersion   string
	UpdateAvailable bool
	IsBuilt         bool
	NeedsToken      bool
}

type infraDef struct {
	name           string
	service        string
	port           string
	hubRepo        string
	imageTag       string
	isBuilt        bool
	needsToken     bool
	composeProfile string
}

var infraDefs = []infraDef{
	{name: "traefik", service: "traefik", port: "8080", hubRepo: "library/traefik", imageTag: "v3.3"},
	{name: "gitea", service: "gitea", port: "3000", hubRepo: "gitea/gitea", imageTag: "1.21"},
	{name: "cloudflared", service: "cloudflared", port: "", hubRepo: "cloudflare/cloudflared", imageTag: "latest", needsToken: true, composeProfile: "cloudflare"},
	{name: "management-ui", service: "management-ui", port: "8081", isBuilt: true},
}

func GetInfraStatus() []InfraContainer {
	result := make([]InfraContainer, len(infraDefs))
	var wg sync.WaitGroup
	cfg := GetCurrentConfig()

	for i, d := range infraDefs {
		wg.Add(1)
		go func(idx int, def infraDef) {
			defer wg.Done()
			c := InfraContainer{Name: def.name, ServiceName: def.service, Port: def.port, IsBuilt: def.isBuilt, ImageVersion: def.imageTag}
			if def.needsToken && cfg.CFToken == "" {
				c.NeedsToken = true
				c.Status = "not configured"
				result[idx] = c
				return
			}

			containerName := resolveInfraContainerName(def.name, def.service)
			out, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", "{{.State.Status}}")
			if err != nil {
				c.Status = "unknown"
			} else {
				c.Status = strings.TrimSpace(string(out))
			}

			imgOut, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", "{{.Config.Image}}")
			if err == nil {
				c.Image = strings.TrimSpace(string(imgOut))
				if parts := strings.SplitN(c.Image, ":", 2); len(parts) == 2 {
					c.ImageVersion = parts[1]
				}
			}

			if !def.isBuilt && def.hubRepo != "" {
				latest, changed := checkDockerHubUpdate(def.hubRepo, def.imageTag, containerName)
				c.LatestVersion = latest
				c.UpdateAvailable = changed
			}

			result[idx] = c
		}(i, d)
	}
	wg.Wait()
	return result
}

func checkDockerHubUpdate(hubRepo, imageTag, containerName string) (string, bool) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/%s", hubRepo, imageTag)
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return imageTag, false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var payload struct{ Digest string `json:"digest"` }
	if err := json.Unmarshal(body, &payload); err != nil || payload.Digest == "" {
		return imageTag, false
	}

	localOut, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", "{{index .RepoDigests 0}}")
	if err != nil {
		return imageTag, false
	}
	localDigest := strings.TrimSpace(string(localOut))
	if idx := strings.Index(localDigest, "@"); idx != -1 {
		localDigest = localDigest[idx+1:]
	}
	if payload.Digest != localDigest {
		return imageTag + " (update available)", true
	}
	return imageTag, false
}

func InfraAction(service, action string) error {
	known := false
	isBuilt := false
	profile := ""
	for _, d := range infraDefs {
		if d.service == service {
			known = true
			isBuilt = d.isBuilt
			profile = d.composeProfile
			break
		}
	}
	if !known {
		return fmt.Errorf("unknown infrastructure service: %s", service)
	}

	composeUp := func(extraArgs ...string) error {
		args := []string{}
		if profile != "" {
			args = append(args, "--profile", profile)
		}
		args = append(args, "up", "--detach", "--no-deps")
		args = append(args, extraArgs...)
		args = append(args, service)
		return runCompose("/opt/localvsp", args...)
	}

	switch action {
	case "start":
		if profile != "" {
			return composeUp()
		}
		return runCompose("/opt/localvsp", "start", service)
	case "stop":
		return runCompose("/opt/localvsp", "stop", service)
	case "restart":
		if profile != "" {
			return composeUp()
		}
		return runCompose("/opt/localvsp", "restart", service)
	case "update":
		if isBuilt {
			return composeUp("--build")
		}
		if err := runCompose("/opt/localvsp", "pull", service); err != nil {
			return err
		}
		return composeUp()
	case "rebuild":
		return composeUp("--build", "--force-recreate")
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func resolveInfraContainerName(defaultName, service string) string {
	out, err := runDockerCmd(defaultTimeout, "", "ps", "-a", "--filter", "label=com.docker.compose.service="+service, "--format", "{{.Names}}")
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if name := strings.TrimSpace(line); name != "" {
				return name
			}
		}
	}
	return defaultName
}