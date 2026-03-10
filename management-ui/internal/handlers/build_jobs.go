package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/localvsp/management-ui/internal/buildlog"
	"github.com/localvsp/management-ui/internal/docker"
)

type buildJob struct {
	mu         sync.Mutex
	project    string
	lines      []string
	done       bool
	failed     bool
	errMsg     string
	subs       []chan string
	logWriter  *buildlog.Writer
	finishedAt time.Time
}

func (j *buildJob) addLine(line string) {
	j.mu.Lock()
	j.lines = append(j.lines, line)
	subs := make([]chan string, len(j.subs))
	copy(subs, j.subs)
	lw := j.logWriter
	j.mu.Unlock()

	if lw != nil {
		lw.WriteLine(line)
	}

	for _, ch := range subs {
		select {
		case ch <- line:
		default:
		}
	}
}

func (j *buildJob) finish(err error) {
	j.mu.Lock()
	j.done = true
	j.finishedAt = time.Now()
	if err != nil {
		j.failed = true
		j.errMsg = err.Error()
	}
	subs := make([]chan string, len(j.subs))
	copy(subs, j.subs)
	j.subs = nil
	lw := j.logWriter
	j.mu.Unlock()

	if lw != nil {
		lw.Finish(err)
	}
	for _, ch := range subs {
		close(ch)
	}
}

func (j *buildJob) subscribe() ([]string, chan string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	snapshot := make([]string, len(j.lines))
	copy(snapshot, j.lines)
	if j.done {
		return snapshot, nil
	}

	ch := make(chan string, 128)
	j.subs = append(j.subs, ch)
	return snapshot, ch
}

func (j *buildJob) unsubscribe(ch chan string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	for i, sub := range j.subs {
		if sub == ch {
			j.subs = append(j.subs[:i], j.subs[i+1:]...)
			return
		}
	}
}

var (
	buildJobsMu sync.Mutex
	buildJobs   = map[string]*buildJob{}
)

func startBuildJob(name string) *buildJob {
	buildJobsMu.Lock()
	if j, ok := buildJobs[name]; ok && !j.done {
		buildJobsMu.Unlock()
		return j
	}

	lw, err := buildlog.Create(name)
	if err != nil {
		log.Printf("Warning: could not create build log for %s: %v", name, err)
	}

	j := &buildJob{project: name, logWriter: lw}
	buildJobs[name] = j
	buildJobsMu.Unlock()

	go func() {
		out := make(chan string, 128)
		errCh := make(chan error, 1)
		go func() { errCh <- docker.DeployDockerProjectStream(name, out) }()
		for line := range out {
			j.addLine(line)
		}
		j.finish(<-errCh)
	}()

	return j
}

func getJob(name string) *buildJob {
	buildJobsMu.Lock()
	defer buildJobsMu.Unlock()
	return buildJobs[name]
}

func getActiveBuilds() map[string]string {
	buildJobsMu.Lock()
	defer buildJobsMu.Unlock()

	result := make(map[string]string, len(buildJobs))
	for name, j := range buildJobs {
		j.mu.Lock()
		switch {
		case !j.done:
			result[name] = "running"
		case j.failed:
			result[name] = "failed"
		default:
			result[name] = "success"
		}
		j.mu.Unlock()
	}
	return result
}

func StartBuildJobCleanup() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			buildJobsMu.Lock()
			for name, j := range buildJobs {
				j.mu.Lock()
				shouldDelete := j.done && !j.finishedAt.IsZero() && time.Since(j.finishedAt) > time.Hour
				j.mu.Unlock()
				if shouldDelete {
					delete(buildJobs, name)
				}
			}
			buildJobsMu.Unlock()
		}
	}()
}

func LocalDockerBuildStreamHandler(w http.ResponseWriter, r *http.Request) {
	name := sanitizeName(r.URL.Query().Get("name"))
	if name == "" || name == "." || name == "/" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})

	var job *buildJob
	for i := 0; i < 20; i++ {
		job = getJob(name)
		if job != nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if job == nil {
		fmt.Fprintf(w, "event: buildfailed\ndata: Build job not found\n\n")
		flusher.Flush()
		return
	}

	snapshot, ch := job.subscribe()
	for _, line := range snapshot {
		fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(line, "\n", " "))
		flusher.Flush()
	}

	if ch == nil {
		emitBuildResult(w, flusher, job)
		return
	}

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case line, ok := <-ch:
			if !ok {
				emitBuildResult(w, flusher, job)
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(line, "\n", " "))
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			job.unsubscribe(ch)
			return
		}
	}
}

func emitBuildResult(w http.ResponseWriter, flusher http.Flusher, job *buildJob) {
	job.mu.Lock()
	defer job.mu.Unlock()
	if job.failed {
		fmt.Fprintf(w, "event: buildfailed\ndata: %s\n\n", template.HTMLEscapeString(job.errMsg))
	} else {
		fmt.Fprintf(w, "event: builddone\ndata: success\n\n")
	}
	flusher.Flush()
}

func BuildsAPIHandler(w http.ResponseWriter, r *http.Request) {
	type buildEntry struct {
		Project   string `json:"project"`
		Status    string `json:"status"`
		LineCount int    `json:"line_count"`
		Error     string `json:"error,omitempty"`
	}

	var entries []buildEntry
	buildJobsMu.Lock()
	for name, j := range buildJobs {
		j.mu.Lock()
		entry := buildEntry{Project: name, LineCount: len(j.lines)}
		switch {
		case !j.done:
			entry.Status = "running"
		case j.failed:
			entry.Status = "failed"
			entry.Error = j.errMsg
		default:
			entry.Status = "success"
		}
		j.mu.Unlock()
		entries = append(entries, entry)
	}
	buildJobsMu.Unlock()

	seen := map[string]bool{}
	for _, entry := range entries {
		seen[entry.Project] = true
	}
	for project, meta := range buildlog.LatestPerProject() {
		if seen[project] {
			continue
		}
		entries = append(entries, buildEntry{
			Project:   project,
			Status:    string(meta.Status),
			LineCount: meta.LineCount,
			Error:     meta.Error,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entries)
}

func BuildLogViewHandler(w http.ResponseWriter, r *http.Request) {
	name := sanitizeName(r.URL.Query().Get("name"))
	if name == "" || name == "." || name == "/" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if j := getJob(name); j != nil {
		j.mu.Lock()
		text := strings.Join(j.lines, "\n")
		j.mu.Unlock()
		fmt.Fprintf(w, "<pre class=\"p-3 text-green-300 text-xs whitespace-pre-wrap break-all leading-5\">%s</pre>", template.HTMLEscapeString(text))
		return
	}

	metas, err := buildlog.ListForProject(name)
	if err != nil || len(metas) == 0 {
		fmt.Fprint(w, "<pre class=\"p-3 text-gray-500 text-xs\">No build log found.</pre>")
		return
	}

	text, err := buildlog.ReadLog(name, metas[0].ID)
	if err != nil {
		fmt.Fprintf(w, "<pre class=\"p-3 text-red-400 text-xs\">Error reading log: %s</pre>", template.HTMLEscapeString(err.Error()))
		return
	}

	fmt.Fprintf(w, "<pre class=\"p-3 text-green-300 text-xs whitespace-pre-wrap break-all leading-5\">%s</pre>", template.HTMLEscapeString(text))
}