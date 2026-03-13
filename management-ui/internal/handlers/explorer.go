package handlers

import (
    "encoding/json"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

func getVspHome() string {
    if v := os.Getenv("VSP_HOME"); v != "" {
        return v
    }
    if v := os.Getenv("VSP_HOST_HOME"); v != "" {
        return v
    }
    // Fallback to current dir for local testing if /vsp-home doesnt exist
    if _, err := os.Stat("/vsp-home"); err == nil {
        return "/vsp-home"
    } else {
        d, _ := os.Getwd()
        return d
    }
}

func ExplorerPageHandler(w http.ResponseWriter, r *http.Request) {
    lang := requestLang(w, r)
    renderTemplate(w, "explorer.html", tplData(lang, map[string]interface{}{
        "Host": hostFromRequest(r),
    }))
}

type FileInfo struct {
    Name  string `json:"name"`
    Path  string `json:"path"`
    IsDir bool   `json:"isDir"`
    Size  int64  `json:"size"`
}

func ExplorerAPIList(w http.ResponseWriter, r *http.Request) {
    base := getVspHome()
    relPath := r.URL.Query().Get("path")
    if relPath == "" {
        relPath = "."
    }
    relPath = filepath.Clean(relPath)
    if strings.Contains(relPath, "..") {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }

    target := filepath.Join(base, relPath)
    entries, err := os.ReadDir(target)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var files []FileInfo
    if relPath != "." && relPath != "/" {
        files = append(files, FileInfo{
            Name:  "..",
            Path:  filepath.Dir(relPath),
            IsDir: true,
        })
    }

    for _, e := range entries {
        info, err := e.Info()
        if err != nil {
            continue
        }
        files = append(files, FileInfo{
            Name:  e.Name(),
            Path:  filepath.Join(relPath, e.Name()),
            IsDir: e.IsDir(),
            Size:  info.Size(),
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}

func ExplorerAPIRead(w http.ResponseWriter, r *http.Request) {
    base := getVspHome()
    relPath := r.URL.Query().Get("path")
    relPath = filepath.Clean(relPath)
    if strings.Contains(relPath, "..") {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }

    target := filepath.Join(base, relPath)
    data, err := os.ReadFile(target)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "text/plain")
    w.Write(data)
}

func ExplorerAPISave(w http.ResponseWriter, r *http.Request) {
    base := getVspHome()
    relPath := r.URL.Query().Get("path")
    relPath = filepath.Clean(relPath)
    if strings.Contains(relPath, "..") {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }

    target := filepath.Join(base, relPath)
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    if err := os.WriteFile(target, body, 0644); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}

func ExplorerAPIUpload(w http.ResponseWriter, r *http.Request) {
    base := getVspHome()
    
    if err := r.ParseMultipartForm(100 << 20); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // In FormData, paths might be provided as form values if dropping a folder
    // But basic multiple file upload uses "files" input. 
    relPath := r.FormValue("path")
    if relPath == "" {
        relPath = "."
    }
    relPath = filepath.Clean(relPath)
    if strings.Contains(relPath, "..") {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }
    
    targetDir := filepath.Join(base, relPath)
    
    files := r.MultipartForm.File["file"]
    paths := r.MultipartForm.Value["filepath"] // Client can send relative paths for folder structure

    for i, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            continue
        }
        
        relFilePath := fileHeader.Filename
        if i < len(paths) && paths[i] != "" {
            relFilePath = paths[i]
        }
        
        cleanRelFilePath := filepath.Clean(relFilePath)
        if strings.Contains(cleanRelFilePath, "..") {
            file.Close()
            continue
        }
        
        fullPath := filepath.Join(targetDir, cleanRelFilePath)
        os.MkdirAll(filepath.Dir(fullPath), 0755)
        
        out, err := os.Create(fullPath)
        if err == nil {
            io.Copy(out, file)
            out.Close()
        }
        file.Close()
    }
    
    w.WriteHeader(http.StatusOK)
}

