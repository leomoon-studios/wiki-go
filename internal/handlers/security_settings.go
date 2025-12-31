package handlers

import (
    "encoding/json"
    "net/http"
    "os"
    "sync"
    "wiki-go/internal/config"
)

var securityMu sync.Mutex

// SecuritySettings represents the JSON payload for security settings.
type SecuritySettings struct {
	PasswordStrength int `yaml:"passwordstrength"`
    LoginBan struct {
        Enabled           bool `json:"enabled"`
        MaxFailures       int  `json:"max_failures"`
        WindowSeconds     int  `json:"window_seconds"`
        InitialBanSeconds int  `json:"initial_ban_seconds"`
        MaxBanSeconds     int  `json:"max_ban_seconds"`
    } `json:"login_ban"`
}

// SecuritySettingsHandler handles GET (read) and POST (update) of security settings.
func SecuritySettingsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        handleGetSecurity(w, r)
    case http.MethodPost, http.MethodPut:
        handleUpdateSecurity(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func handleGetSecurity(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var resp SecuritySettings
	resp.PasswordStrength = cfg.Security.PasswordStrength
    resp.LoginBan.Enabled = cfg.Security.LoginBan.Enabled
    resp.LoginBan.MaxFailures = cfg.Security.LoginBan.MaxFailures
    resp.LoginBan.WindowSeconds = cfg.Security.LoginBan.WindowSeconds
    resp.LoginBan.InitialBanSeconds = cfg.Security.LoginBan.InitialBanSeconds
    resp.LoginBan.MaxBanSeconds = cfg.Security.LoginBan.MaxBanSeconds

    json.NewEncoder(w).Encode(resp)
}

func handleUpdateSecurity(w http.ResponseWriter, r *http.Request) {
    var req SecuritySettings
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Basic validation
    if req.LoginBan.MaxFailures <= 0 || req.LoginBan.WindowSeconds <= 0 || req.LoginBan.InitialBanSeconds <= 0 || req.LoginBan.MaxBanSeconds < req.LoginBan.InitialBanSeconds {
        http.Error(w, "Invalid values", http.StatusBadRequest)
        return
    }

    securityMu.Lock()
    defer securityMu.Unlock()

    // Update cfg in-memory
	cfg.Security.PasswordStrength = req.PasswordStrength
    cfg.Security.LoginBan.Enabled = req.LoginBan.Enabled
    cfg.Security.LoginBan.MaxFailures = req.LoginBan.MaxFailures
    cfg.Security.LoginBan.WindowSeconds = req.LoginBan.WindowSeconds
    cfg.Security.LoginBan.InitialBanSeconds = req.LoginBan.InitialBanSeconds
    cfg.Security.LoginBan.MaxBanSeconds = req.LoginBan.MaxBanSeconds

    // Persist to disk
    // Reuse SaveConfig with config.ConfigFilePath
    f, err := os.Create(config.ConfigFilePath)
    if err == nil {
        _ = config.SaveConfig(cfg, f)
        f.Close()
    }

    // Reinitialise ban list with new policy
    InitLoginBan(cfg)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(req)
}
