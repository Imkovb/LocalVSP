package docker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetCurrentConfig() Config {
	cfg := Config{}
	data, err := os.ReadFile(configEnvPath())
	if err != nil {
		return cfg
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "CLOUDFLARE_TUNNEL_TOKEN":
			cfg.CFToken = parts[1]
		case "VSP_DOMAIN":
			cfg.Domain = parts[1]
		}
	}
	return cfg
}

func SaveConfig(cfg Config) error {
	if err := validateDomain(cfg.Domain); err != nil {
		return err
	}
	existing, _ := os.ReadFile(configEnvPath())
	content := mergeEnv(string(existing), map[string]string{
		"CLOUDFLARE_TUNNEL_TOKEN": cfg.CFToken,
		"VSP_DOMAIN":             strings.TrimSpace(cfg.Domain),
	})
	return writeFileAtomic(configEnvPath(), []byte(content), 0600)
}

func StartCloudflared() error {
	return runCompose("/opt/localvsp", "--profile", "cloudflare", "up", "--detach", "--no-deps", "cloudflared")
}

func StopCloudflared() error {
	return runCompose("/opt/localvsp", "stop", "cloudflared")
}

func ApplyPlatformRoutingConfig() error {
	cfg := GetCurrentConfig()

	if err := applyCoreRoutingConfig(cfg); err != nil {
		return err
	}

	if err := applyProjectRoutingConfig(cfg); err != nil {
		return err
	}

	return nil
}

func ReconcileExistingProjectRouting() error {
	return applyProjectRoutingConfig(GetCurrentConfig())
}

func applyCoreRoutingConfig(cfg Config) error {
	if strings.TrimSpace(cfg.CFToken) == "" {
		_ = StopCloudflared()
		return nil
	}

	return StartCloudflared()
}

func applyProjectRoutingConfig(cfg Config) error {
	autoCfg := LoadAutoStartConfig()
	dirty := false

	for name, projectCfg := range autoCfg.DockerProjects {
		dir := filepath.Join(homeVSP, "docker", name)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		if projectCfg.ExposePort == "" {
			projectCfg.ExposePort = DetectInternalPort(dir)
			autoCfg.DockerProjects[name] = projectCfg
			dirty = true
		}
		if _, ok := findComposeFilePath(dir); ok {
			changed, err := syncDockerOverride(dir, name, projectCfg, cfg.Domain)
			if err != nil {
				return fmt.Errorf("apply routing for project %s: %w", name, err)
			}
			if changed && len(composeContainerIDs(dir)) > 0 {
				if err := runCompose(dir, "up", "--detach", "--remove-orphans"); err != nil {
					return fmt.Errorf("refresh running project %s: %w", name, err)
				}
			}
		}
	}

	for name, siteCfg := range autoCfg.HtmlSites {
		containerName := "html-" + name
		statusOut, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", "{{.State.Status}}")
		if err != nil || strings.TrimSpace(string(statusOut)) != "running" {
			continue
		}
		if !htmlSiteNeedsRoutingRefresh(name, siteCfg, cfg.Domain) {
			continue
		}
		if _, err := DeployHtmlSite(name); err != nil {
			return fmt.Errorf("refresh running html site %s: %w", name, err)
		}
	}

	if dirty {
		if err := SaveAutoStartConfig(autoCfg); err != nil {
			return err
		}
	}
	if err := SyncTraefikDynamicConfig(); err != nil {
		return err
	}

	return nil
}