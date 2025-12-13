package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
	"wiki-go/internal/utils"
)

// BackupJob represents a backup operation
type BackupJob struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // "processing", "completed", "failed"
	Progress  int       `json:"progress"`
	TotalFiles int      `json:"totalFiles"`
	ProcessedFiles int  `json:"processedFiles"`
	CurrentFile string  `json:"currentFile"`
	Error     string    `json:"error,omitempty"`
	Filename  string    `json:"filename,omitempty"`
}

// BackupFile represents a backup file on disk
type BackupFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Date string `json:"date"`
	URL  string `json:"url"`
}

var (
	backupJobs      = make(map[string]*BackupJob)
	backupJobsMutex sync.RWMutex
)

// StartBackupHandler initiates a new backup job
func StartBackupHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Check admin role
	session := auth.GetSession(r)
	if session == nil || session.Role != config.RoleAdmin {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create backup directory if it doesn't exist
	backupDir := filepath.Join(cfg.Wiki.RootDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		http.Error(w, "Failed to create backup directory", http.StatusInternalServerError)
		return
	}

	jobID := fmt.Sprintf("%d", time.Now().UnixNano())
	filename := fmt.Sprintf("backup_%s.zip", time.Now().Format("20060102150405"))
	
	job := &BackupJob{
		ID:       jobID,
		Status:   "processing",
		Progress: 0,
		Filename: filename,
	}

	backupJobsMutex.Lock()
	backupJobs[jobID] = job
	backupJobsMutex.Unlock()

	// Start backup in background
	go performBackup(jobID, filename, cfg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"jobId":     jobID,
		"statusUrl": "/api/backup/status/" + jobID,
	})
}

func performBackup(jobID, filename string, cfg *config.Config) {
	updateJob(jobID, func(j *BackupJob) {
		j.Status = "processing"
	})

	// 1. Count files
	totalFiles := 0
	filepath.Walk(cfg.Wiki.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Skip temp, cache, and backups directories
			if info.Name() == "temp" || info.Name() == "cache" || info.Name() == "backups" {
				return filepath.SkipDir
			}
			return nil
		}
		totalFiles++
		return nil
	})

	updateJob(jobID, func(j *BackupJob) {
		j.TotalFiles = totalFiles
	})

	// 2. Create Zip
	backupPath := filepath.Join(cfg.Wiki.RootDir, "backups", filename)
	zipFile, err := os.Create(backupPath)
	if err != nil {
		failJob(jobID, "Failed to create zip file: "+err.Error())
		return
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	processed := 0
	err = filepath.Walk(cfg.Wiki.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == "temp" || info.Name() == "cache" || info.Name() == "backups" {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path for zip
		relPath, err := filepath.Rel(cfg.Wiki.RootDir, path)
		if err != nil {
			return nil
		}

		// Update progress
		processed++
		updateJob(jobID, func(j *BackupJob) {
			j.ProcessedFiles = processed
			j.CurrentFile = relPath
			if totalFiles > 0 {
				j.Progress = int(float64(processed) / float64(totalFiles) * 100)
			}
		})

		// Add to zip
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return nil
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		failJob(jobID, "Error during backup: "+err.Error())
		return
	}

	updateJob(jobID, func(j *BackupJob) {
		j.Status = "completed"
		j.Progress = 100
	})
}

func updateJob(jobID string, updater func(*BackupJob)) {
	backupJobsMutex.Lock()
	defer backupJobsMutex.Unlock()
	if job, ok := backupJobs[jobID]; ok {
		updater(job)
	}
}

func failJob(jobID, errorMsg string) {
	updateJob(jobID, func(j *BackupJob) {
		j.Status = "failed"
		j.Error = errorMsg
	})
}

// BackupStatusHandler returns the status of a backup job
func BackupStatusHandler(w http.ResponseWriter, r *http.Request) {
	jobID := strings.TrimPrefix(r.URL.Path, "/api/backup/status/")

	// Check admin role
	session := auth.GetSession(r)
	if session == nil || session.Role != config.RoleAdmin {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	backupJobsMutex.RLock()
	job, ok := backupJobs[jobID]
	backupJobsMutex.RUnlock()

	if !ok {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// ListBackupsHandler lists all available backup files
func ListBackupsHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Check admin role
	session := auth.GetSession(r)
	if session == nil || session.Role != config.RoleAdmin {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	backupDir := filepath.Join(cfg.Wiki.RootDir, "backups")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		// Return empty list if directory doesn't exist
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"backups": []BackupFile{},
		})
		return
	}

	files, err := os.ReadDir(backupDir)
	if err != nil {
		http.Error(w, "Failed to read backup directory", http.StatusInternalServerError)
		return
	}

	var backups []BackupFile
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".zip") {
			info, err := file.Info()
			if err != nil {
				continue
			}
			backups = append(backups, BackupFile{
				Name: file.Name(),
				Size: info.Size(),
				Date: info.ModTime().Format("2006-01-02 15:04:05"),
				URL:  "/api/backup/download/" + file.Name(),
			})
		}
	}

	// Sort by date descending (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Date > backups[j].Date
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backups": backups,
	})
}

// DownloadBackupHandler serves a backup file
func DownloadBackupHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	filename := strings.TrimPrefix(r.URL.Path, "/api/backup/download/")

	// Check admin role
	session := auth.GetSession(r)
	if session == nil || session.Role != config.RoleAdmin {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate filename to prevent directory traversal
	if !utils.IsValidFilename(filename) || !strings.HasSuffix(filename, ".zip") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(cfg.Wiki.RootDir, "backups", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/zip")
	http.ServeFile(w, r, filePath)
}

// DeleteBackupHandler deletes a backup file
func DeleteBackupHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	filename := strings.TrimPrefix(r.URL.Path, "/api/backup/delete/")

	// Check admin role
	session := auth.GetSession(r)
	if session == nil || session.Role != config.RoleAdmin {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate filename
	if !utils.IsValidFilename(filename) || !strings.HasSuffix(filename, ".zip") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(cfg.Wiki.RootDir, "backups", filename)
	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
