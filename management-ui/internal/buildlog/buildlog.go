// Package buildlog persists build output to disk so logs survive browser
// refreshes, tab closes and server restarts.
package buildlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const baseDir = "/vsp-home/.localvsp/build-logs"
const maxBuildsPerProject = 10

// Status represents the result state of a build.
type Status string

const (
	Running Status = "running"
	Success Status = "success"
	Failed  Status = "failed"
)

// Meta holds the persisted metadata for one build run.
type Meta struct {
	ID         string     `json:"id"`
	Project    string     `json:"project"`
	Status     Status     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Error      string     `json:"error,omitempty"`
	LineCount  int        `json:"line_count"`
}

// Writer handles appending lines to a build's log file and updating its meta.
type Writer struct {
	mu       sync.Mutex
	meta     Meta
	logFile  *os.File
	metaPath string
}

// Create starts a new build log. The caller must call Finish when done.
func Create(project string) (*Writer, error) {
	id := time.Now().Format("20060102-150405")
	dir := filepath.Join(baseDir, project)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %v", err)
	}

	logFile, err := os.Create(filepath.Join(dir, id+".log"))
	if err != nil {
		return nil, fmt.Errorf("create log file: %v", err)
	}

	w := &Writer{
		meta: Meta{
			ID:        id,
			Project:   project,
			Status:    Running,
			StartedAt: time.Now(),
		},
		logFile:  logFile,
		metaPath: filepath.Join(dir, id+".meta.json"),
	}
	w.saveMeta()
	return w, nil
}

// Meta returns a copy of the current metadata.
func (w *Writer) Meta() Meta {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.meta
}

// WriteLine appends a single line to the log file.
func (w *Writer) WriteLine(line string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fmt.Fprintln(w.logFile, line)
	w.meta.LineCount++
}

// Finish marks the build as complete, writes final meta, closes the file,
// and prunes old logs beyond the retention limit.
func (w *Writer) Finish(buildErr error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	w.meta.FinishedAt = &now
	if buildErr != nil {
		w.meta.Status = Failed
		w.meta.Error = buildErr.Error()
	} else {
		w.meta.Status = Success
	}
	w.saveMeta()
	w.logFile.Close()
	cleanup(w.meta.Project)
}

func (w *Writer) saveMeta() {
	data, err := json.MarshalIndent(w.meta, "", "  ")
	if err != nil {
		return
	}
	writeMetaFile(w.metaPath, data)
}

// ─── Read Functions ──────────────────────────────────────────────────────────

// ReadLog returns the full log text for a build identified by project and id.
func ReadLog(project, id string) (string, error) {
	path := filepath.Join(baseDir, project, id+".log")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListForProject returns build metas for a project, sorted newest first.
func ListForProject(project string) ([]Meta, error) {
	dir := filepath.Join(baseDir, project)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var metas []Meta
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".meta.json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var m Meta
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		metas = append(metas, m)
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].StartedAt.After(metas[j].StartedAt)
	})
	return metas, nil
}

// LatestPerProject returns the most recent Meta for each project.
func LatestPerProject() map[string]Meta {
	result := map[string]Meta{}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return result
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		metas, err := ListForProject(e.Name())
		if err != nil || len(metas) == 0 {
			continue
		}
		result[e.Name()] = metas[0]
	}
	return result
}

// RecoverStale finds builds still marked "running" on disk (orphaned by a
// server restart) and marks them as failed.
func RecoverStale() {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		metas, err := ListForProject(e.Name())
		if err != nil {
			continue
		}
		for _, m := range metas {
			if m.Status != Running {
				continue
			}
			// Mark as failed
			now := time.Now()
			m.Status = Failed
			m.Error = "Server restarted during build"
			m.FinishedAt = &now
			data, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				continue
			}
			metaPath := filepath.Join(baseDir, e.Name(), m.ID+".meta.json")
			writeMetaFile(metaPath, data)
		}
	}
}

// ─── Cleanup ─────────────────────────────────────────────────────────────────

// cleanup removes old log+meta pairs beyond the retention limit for a project.
func cleanup(project string) {
	metas, err := ListForProject(project)
	if err != nil || len(metas) <= maxBuildsPerProject {
		return
	}
	dir := filepath.Join(baseDir, project)
	for _, m := range metas[maxBuildsPerProject:] {
		os.Remove(filepath.Join(dir, m.ID+".log"))      //nolint:errcheck
		os.Remove(filepath.Join(dir, m.ID+".meta.json")) //nolint:errcheck
	}
}

func writeMetaFile(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".meta-*")
	if err != nil {
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) //nolint:errcheck
	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck
		return
	}
	if err := tmp.Chmod(0644); err != nil {
		tmp.Close() //nolint:errcheck
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}
	os.Rename(tmpPath, path) //nolint:errcheck
}
