package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LocalHtmlSite struct {
	Name      string
	Running   bool
	Port      string
	AutoStart bool
	Subdomain string
}

type LocalDockerProject struct {
	Name       string
	Status     string
	Services   []Service
	HasCompose bool
	AutoStart  bool
	Subdomain  string
	ExposePort string
	HostPort   string
}

type SiteConfig struct {
	AutoStart bool   `json:"autostart"`
	Subdomain string `json:"subdomain"`
}

type ProjectConfig struct {
	AutoStart  bool   `json:"autostart"`
	Subdomain  string `json:"subdomain"`
	ExposePort string `json:"expose_port"`
	HostPort   string `json:"host_port"`
}

type AutoStartConfig struct {
	HtmlSites      map[string]SiteConfig    `json:"html_sites"`
	DockerProjects map[string]ProjectConfig `json:"docker_projects"`
}

func LoadAutoStartConfig() AutoStartConfig {
	cfg := AutoStartConfig{HtmlSites: map[string]SiteConfig{}, DockerProjects: map[string]ProjectConfig{}}
	data, err := os.ReadFile(autoStartConfigPath())
	if err != nil {
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg
	}
	if cfg.HtmlSites == nil {
		cfg.HtmlSites = map[string]SiteConfig{}
	}
	if cfg.DockerProjects == nil {
		cfg.DockerProjects = map[string]ProjectConfig{}
	}
	return cfg
}

func SaveAutoStartConfig(cfg AutoStartConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(autoStartConfigPath(), data, 0644)
}

func containerRestartPolicy(name string) string {
	out, err := runDockerCmd(defaultTimeout, "", "inspect", name, "--format", "{{.HostConfig.RestartPolicy.Name}}")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func composeContainerIDs(dir string) []string {
	out, err := runDockerCmd(defaultTimeout, dir, "compose", "ps", "-q")
	if err != nil {
		return nil
	}
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if id := strings.TrimSpace(line); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func ToggleHtmlAutoStart(name string, enable bool) error {
	cfg := LoadAutoStartConfig()
	sc := cfg.HtmlSites[name]
	sc.AutoStart = enable
	cfg.HtmlSites[name] = sc
	if err := SaveAutoStartConfig(cfg); err != nil {
		return err
	}
	policy := "no"
	if enable {
		policy = "unless-stopped"
	}
	_, _ = runDockerCmd(defaultTimeout, "", "update", "--restart", policy, "html-"+name)
	return nil
}

func ToggleDockerAutoStart(name string, enable bool) error {
	cfg := LoadAutoStartConfig()
	pc := cfg.DockerProjects[name]
	pc.AutoStart = enable
	cfg.DockerProjects[name] = pc
	if err := SaveAutoStartConfig(cfg); err != nil {
		return err
	}
	policy := "no"
	if enable {
		policy = "unless-stopped"
	}
	dir := filepath.Join(homeVSP, "docker", name)
	if ids := composeContainerIDs(dir); len(ids) > 0 {
		args := append([]string{"update", "--restart", policy}, ids...)
		_, _ = runDockerCmd(defaultTimeout, "", args...)
	}
	return nil
}

func SetHtmlSubdomain(name, subdomain string) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid site name: %s", name)
	}
	if !isValidSubdomain(strings.TrimSpace(subdomain)) {
		return fmt.Errorf("invalid subdomain")
	}
	cfg := LoadAutoStartConfig()
	sc := cfg.HtmlSites[name]
	sc.Subdomain = strings.ToLower(strings.TrimSpace(subdomain))
	cfg.HtmlSites[name] = sc
	return SaveAutoStartConfig(cfg)
}

func SetDockerSubdomain(name, subdomain string) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid project name: %s", name)
	}
	if !isValidSubdomain(strings.TrimSpace(subdomain)) {
		return fmt.Errorf("invalid subdomain")
	}
	cfg := LoadAutoStartConfig()
	pc := cfg.DockerProjects[name]
	pc.Subdomain = strings.ToLower(strings.TrimSpace(subdomain))
	dir := filepath.Join(homeVSP, "docker", name)
	if pc.ExposePort == "" {
		pc.ExposePort = DetectInternalPort(dir)
	}
	cfg.DockerProjects[name] = pc
	if err := SaveAutoStartConfig(cfg); err != nil {
		return err
	}
	return GenerateDockerOverride(dir, name, pc, GetCurrentConfig().Domain)
}

func DetectInternalPort(dir string) string {
	if b, err := os.ReadFile(filepath.Join(dir, "Dockerfile")); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "EXPOSE ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return parts[1]
				}
			}
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		return "3000"
	}
	if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
		return "8000"
	}
	return "80"
}

func GenerateDockerOverride(dir, name string, pc ProjectConfig, domain string) error {
	_, err := syncDockerOverride(dir, name, pc, domain)
	return err
}

func syncDockerOverride(dir, name string, pc ProjectConfig, domain string) (bool, error) {
	serviceName := getFirstComposeService(dir)
	if serviceName == "" {
		return false, fmt.Errorf("could not detect service name from compose file")
	}
	if pc.ExposePort == "" {
		pc.ExposePort = DetectInternalPort(dir)
	}

	content, err := dockerOverrideContent(serviceName, name, pc, domain)
	if err != nil {
		return false, err
	}
	path := filepath.Join(dir, "docker-compose.override.yml")
	if content == "" {
		if _, err := os.Stat(path); err == nil {
			RemoveDockerOverride(dir)
			return true, nil
		}
		return false, nil
	}

	existing, _ := os.ReadFile(path)
	if string(existing) == content {
		return false, nil
	}
	if err := writeFileAtomic(path, []byte(content), 0644); err != nil {
		return false, err
	}
	return true, nil
}

func dockerOverrideContent(serviceName, name string, pc ProjectConfig, domain string) (string, error) {
	var sb strings.Builder
	sb.WriteString("# Generated by LocalVSP — do not edit manually\n")
	hasNetwork := shouldExposeThroughTraefik(pc.Subdomain, domain)
	if hasNetwork {
		sb.WriteString("networks:\n")
		sb.WriteString("  vsp-network:\n")
		sb.WriteString("    external: true\n")
		sb.WriteString("    name: " + traefikNetwork + "\n")
	}
	sb.WriteString("services:\n")
	sb.WriteString(fmt.Sprintf("  %s:\n", serviceName))
	if hasNetwork {
		sb.WriteString("    networks:\n")
		sb.WriteString("      - vsp-network\n")
	}
	if pc.HostPort != "" {
		sb.WriteString("    ports:\n")
		sb.WriteString(fmt.Sprintf("      - \"%s:%s\"\n", pc.HostPort, pc.ExposePort))
	}
	if hasNetwork {
		routerName := traefikRouterName("vsp", name)
		sb.WriteString("    labels:\n")
		sb.WriteString("      - \"traefik.enable=true\"\n")
		sb.WriteString(fmt.Sprintf("      - \"traefik.docker.network=%s\"\n", traefikNetwork))
		sb.WriteString(fmt.Sprintf("      - \"traefik.http.routers.%s.entrypoints=web\"\n", routerName))
		sb.WriteString(fmt.Sprintf("      - \"traefik.http.routers.%s.rule=Host(`%s.%s`)\"\n", routerName, pc.Subdomain, domain))
		sb.WriteString(fmt.Sprintf("      - \"traefik.http.services.%s.loadbalancer.server.port=%s\"\n", routerName, pc.ExposePort))
	}
	if !hasNetwork && pc.HostPort == "" {
		return "", nil
	}
	return sb.String(), nil
}

func RemoveDockerOverride(dir string) {
	_ = os.Remove(filepath.Join(dir, "docker-compose.override.yml"))
}

func ListHtmlSites() []LocalHtmlSite {
	base := filepath.Join(homeVSP, "html")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	cfg := LoadAutoStartConfig()
	var sites []LocalHtmlSite
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sc := cfg.HtmlSites[e.Name()]
		site := LocalHtmlSite{Name: e.Name(), Subdomain: sc.Subdomain}
		containerName := "html-" + e.Name()
		statusOut, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", "{{.State.Status}}")
		if err == nil && strings.TrimSpace(string(statusOut)) == "running" {
			site.Running = true
			portOut, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", `{{(index (index .NetworkSettings.Ports "80/tcp") 0).HostPort}}`)
			if err == nil {
				site.Port = strings.TrimSpace(string(portOut))
			}
		}
		if policy := containerRestartPolicy(containerName); policy != "" {
			site.AutoStart = policy == "unless-stopped"
		} else {
			site.AutoStart = sc.AutoStart
		}
		sites = append(sites, site)
	}
	cleanupOrphanedHtmlSites(cfg, base)
	return sites
}

func ListDockerProjects() []LocalDockerProject {
	base := filepath.Join(homeVSP, "docker")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	cfg := LoadAutoStartConfig()
	var projects []LocalDockerProject
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(base, e.Name())
		pc := cfg.DockerProjects[e.Name()]
		p := LocalDockerProject{Name: e.Name(), Subdomain: pc.Subdomain, ExposePort: pc.ExposePort, HostPort: pc.HostPort}
		if _, ok := findComposeFilePath(dir); ok {
			p.HasCompose = true
			p.Services, p.Status = getComposeServicesStatus(dir)
			if p.Status == "" {
				p.Status = "stopped"
			}
			if p.HostPort == "" {
				for _, svc := range p.Services {
					if svc.Ports != "" {
						if idx := strings.Index(svc.Ports, "->"); idx > 0 {
							p.HostPort = strings.TrimSpace(svc.Ports[:idx])
						}
						break
					}
				}
			}
		} else if detectsAsKnownType(dir) {
			p.Status = "no-compose"
		} else {
			p.Status = "unknown"
		}
		if ids := composeContainerIDs(dir); len(ids) > 0 {
			p.AutoStart = containerRestartPolicy(ids[0]) == "unless-stopped"
		} else {
			p.AutoStart = pc.AutoStart
		}
		projects = append(projects, p)
	}
	cleanupOrphanedDockerProjects(cfg, base)
	return projects
}

func DeployHtmlSite(name string) (string, error) {
	containerSiteDir := filepath.Join(homeVSP, "html", name)
	if _, err := os.Stat(containerSiteDir); os.IsNotExist(err) {
		return "", fmt.Errorf("folder html/%s does not exist in /home/vsp", name)
	}
	hostSiteDir := filepath.Join(hostHomeVSP(), "html", name)
	containerName := "html-" + name
	portStr := inspectPublishedPort(containerName, "80/tcp")
	_, _ = runDockerCmd(defaultTimeout, "", "rm", "-f", containerName)
	if portStr == "" {
		port, err := FindFreePort(8200, 8299)
		if err != nil {
			return "", err
		}
		portStr = strconv.Itoa(port)
	}
	cfg := LoadAutoStartConfig()
	sc := cfg.HtmlSites[name]
	restartPolicy := "no"
	if sc.AutoStart {
		restartPolicy = "unless-stopped"
	}
	args := []string{"run", "-d", "--name", containerName, "-p", portStr + ":80", "-v", hostSiteDir + ":/usr/share/nginx/html:ro", "--restart", restartPolicy}
	if shouldExposeThroughTraefik(sc.Subdomain, GetCurrentConfig().Domain) {
		domain := GetCurrentConfig().Domain
		routerName := traefikRouterName("html", name)
		args = append(args,
			"--network", traefikNetwork,
			"--label", "traefik.enable=true",
			"--label", fmt.Sprintf("traefik.docker.network=%s", traefikNetwork),
			"--label", fmt.Sprintf("traefik.http.routers.%s.entrypoints=web", routerName),
			"--label", fmt.Sprintf("traefik.http.routers.%s.rule=Host(`%s.%s`)", routerName, sc.Subdomain, domain),
			"--label", fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=80", routerName),
		)
	}
	args = append(args, "nginx:alpine")
	out, err := runDockerCmd(defaultTimeout, "", args...)
	if err != nil {
		return "", fmt.Errorf("docker run failed: %s", strings.TrimSpace(string(out)))
	}
	if err := SyncTraefikDynamicConfig(); err != nil {
		return "", err
	}
	return portStr, nil
}

func StopHtmlSite(name string) error {
	out, err := runDockerCmd(defaultTimeout, "", "rm", "-f", "html-"+name)
	if err != nil {
		return fmt.Errorf("docker rm failed: %s", strings.TrimSpace(string(out)))
	}
	if err := SyncTraefikDynamicConfig(); err != nil {
		return err
	}
	return nil
}

func inspectPublishedPort(containerName, containerPort string) string {
	out, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", fmt.Sprintf(`{{(index (index .NetworkSettings.Ports %q) 0).HostPort}}`, containerPort))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func containerLabelValue(containerName, label string) string {
	out, err := runDockerCmd(defaultTimeout, "", "inspect", containerName, "--format", fmt.Sprintf(`{{index .Config.Labels %q}}`, label))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func htmlSiteNeedsRoutingRefresh(name string, sc SiteConfig, domain string) bool {
	containerName := "html-" + name
	if !shouldExposeThroughTraefik(sc.Subdomain, domain) {
		return containerLabelValue(containerName, "traefik.enable") == "true"
	}
	routerName := traefikRouterName("html", name)
	return containerLabelValue(containerName, "traefik.docker.network") != traefikNetwork ||
		containerLabelValue(containerName, fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)) != "web" ||
		containerLabelValue(containerName, fmt.Sprintf("traefik.http.routers.%s.rule", routerName)) != fmt.Sprintf("Host(`%s.%s`)", sc.Subdomain, domain)
}

func DeployDockerProject(name string) error {
	out := make(chan string, 64)
	errCh := make(chan error, 1)
	go func() { errCh <- DeployDockerProjectStream(name, out) }()
	for range out {
	}
	return <-errCh
}

func DeployDockerProjectStream(name string, out chan<- string) error {
	defer close(out)
	send := func(msg string) {
		select {
		case out <- msg:
		default:
		}
	}

	dir := filepath.Join(homeVSP, "docker", name)
	send("==> Auto-detecting project type...")
	if err := autoGenerateDockerFiles(dir, name); err != nil {
		return fmt.Errorf("auto-generation failed: %v", err)
	}
	send("==> Project files ready.")

	cfg := LoadAutoStartConfig()
	pc := cfg.DockerProjects[name]
	if pc.HostPort == "" {
		port, err := FindFreePort(8300, 8399)
		if err == nil {
			pc.HostPort = strconv.Itoa(port)
		}
	}
	if pc.ExposePort == "" {
		pc.ExposePort = DetectInternalPort(dir)
	}
	cfg.DockerProjects[name] = pc
	if err := SaveAutoStartConfig(cfg); err != nil {
		return fmt.Errorf("save project config: %v", err)
	}

	domain := GetCurrentConfig().Domain
	if pc.Subdomain != "" {
		send(fmt.Sprintf("==> Configuring subdomain: %s.%s", pc.Subdomain, domain))
	}
	if err := GenerateDockerOverride(dir, name, pc, domain); err != nil {
		return fmt.Errorf("generate docker override: %v", err)
	}
	if err := requireComposeFile(dir); err != nil {
		return err
	}

	runStream := func(label string, args ...string) error {
		send("==> " + label)
		ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "docker", args...)
		cmd.Dir = dir
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("stdout pipe: %v", err)
		}
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("stderr pipe: %v", err)
		}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start: %v", err)
		}
		var wg sync.WaitGroup
		wg.Add(2)
		reader := func(r io.Reader) {
			defer wg.Done()
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				send(scanner.Text())
			}
		}
		go reader(stdoutPipe)
		go reader(stderrPipe)
		wg.Wait()
		return cmd.Wait()
	}

	if err := runStream("docker compose build --no-cache", "compose", "build", "--no-cache"); err != nil {
		return fmt.Errorf("docker compose build failed: %v", err)
	}
	if err := runStream("docker compose up --detach --remove-orphans", "compose", "up", "--detach", "--remove-orphans"); err != nil {
		return fmt.Errorf("docker compose up failed: %v", err)
	}

	send("==> Waiting for containers to start...")
	time.Sleep(3 * time.Second)
	svcList, svcStatus := getComposeServicesStatus(dir)
	if svcStatus != "running" {
		send(fmt.Sprintf("==> WARNING: Containers are not running (status: %s)", svcStatus))
		logsOut, _ := runDockerCmd(buildTimeout, dir, "compose", "logs", "--tail", "30", "--no-color")
		if len(logsOut) > 0 {
			send("==> Container logs:")
			for _, line := range strings.Split(strings.TrimSpace(string(logsOut)), "\n") {
				send(line)
			}
		}
		return fmt.Errorf("containers exited after start (status: %s) — check logs above", svcStatus)
	}
	send(fmt.Sprintf("==> %d service(s) running", len(svcList)))
	send("==> Applying restart policy...")
	policy := "no"
	if pc.AutoStart {
		policy = "unless-stopped"
	}
	if ids := composeContainerIDs(dir); len(ids) > 0 {
		args := append([]string{"update", "--restart", policy}, ids...)
		_, _ = runDockerCmd(defaultTimeout, "", args...)
	}
	if err := SyncTraefikDynamicConfig(); err != nil {
		return err
	}
	send("==> Done!")
	return nil
}

func StopDockerProject(name string) error {
	if err := ComposeDown(filepath.Join(homeVSP, "docker", name)); err != nil {
		return err
	}
	return SyncTraefikDynamicConfig()
}

func DeleteDockerProject(name string) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid project name: %s", name)
	}
	dir := filepath.Join(homeVSP, "docker", name)
	if _, ok := findComposeFilePath(dir); ok {
		_ = runCompose(dir, "down", "--volumes", "--remove-orphans")
	}
	RemoveDockerOverride(dir)
	cfg := LoadAutoStartConfig()
	delete(cfg.DockerProjects, name)
	_ = SaveAutoStartConfig(cfg)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove project folder: %v", err)
	}
	if err := SyncTraefikDynamicConfig(); err != nil {
		return err
	}
	return nil
}

func DeleteHtmlSite(name string) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid site name: %s", name)
	}
	dir := filepath.Join(homeVSP, "html", name)
	_, _ = runDockerCmd(defaultTimeout, "", "rm", "-f", "html-"+name)
	cfg := LoadAutoStartConfig()
	delete(cfg.HtmlSites, name)
	_ = SaveAutoStartConfig(cfg)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove site folder: %v", err)
	}
	if err := SyncTraefikDynamicConfig(); err != nil {
		return err
	}
	return nil
}

func FindFreePort(minP, maxP int) (int, error) {
	out, _ := runDockerCmd(defaultTimeout, "", "ps", "--format", "{{.Ports}}")
	used := map[int]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		for _, part := range strings.Fields(line) {
			if idx := strings.Index(part, "->"); idx != -1 {
				hostSide := part[:idx]
				if ci := strings.LastIndex(hostSide, ":"); ci != -1 {
					if p, err := strconv.Atoi(hostSide[ci+1:]); err == nil {
						used[p] = true
					}
				}
			}
		}
	}
	for p := minP; p <= maxP; p++ {
		if !used[p] {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no free port available in range %d-%d", minP, maxP)
}