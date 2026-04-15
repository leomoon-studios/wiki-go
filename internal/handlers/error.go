package handlers

import (
    "bytes"
    "html/template"
    "net/http"
    "strings"
    "time"

    "wiki-go/internal/auth"
    "wiki-go/internal/config"
    "wiki-go/internal/i18n"
    "wiki-go/internal/types"
    "wiki-go/internal/utils"
)

// NotFoundHandler renders a clean 404 page using a dedicated template (templates/404.html).
// All presentation logic resides in the template; this Go code only supplies data.
func NotFoundHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
    // Ensure 404 status
    w.WriteHeader(http.StatusNotFound)

    // Session / role information
    session := auth.GetSession(r)
    isAuthenticated := session != nil
    userRole := ""
    if isAuthenticated {
        userRole = session.Role
    }

    // Navigation tree
    nav, err := utils.BuildNavigation(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, cfg.Wiki.AlwaysOpenChildrenInSidebar)
    if err != nil {
        http.Error(w, "Error building navigation: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Requested path and breadcrumbs
    requestedPath := r.URL.Path
    parts := strings.Split(strings.Trim(requestedPath, "/"), "/")
    breadcrumbs := make([]types.BreadcrumbItem, 0, len(parts)+1)
    breadcrumbs = append(breadcrumbs, types.BreadcrumbItem{Title: "Home", Path: "/", IsLast: len(parts) == 0})
    current := ""
    for i, p := range parts {
        if p == "" {
            continue
        }
        if current == "" {
            current = p
        } else {
            current += "/" + p
        }
        breadcrumbs = append(breadcrumbs, types.BreadcrumbItem{Title: utils.FormatDirName(p), Path: "/" + current, IsLast: i == len(parts)-1})
    }

    // Base template data (Content will be filled afterwards)
    data := &types.PageData{
        Navigation:         nav,
        Breadcrumbs:        breadcrumbs,
        Config:             cfg,
        CurrentDir:         &types.NavItem{Title: "404 - Page Not Found", Path: requestedPath},
        Title:              "404 - Page Not Found",
        AvailableLanguages: i18n.GetAvailableLanguages(),
        IsAuthenticated:    isAuthenticated,
        UserRole:           userRole,
        LastModified:       time.Now(),
    }

    // Render the not-found specific template fragment into .Content
    tmpl, err := getTemplate()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    var buf bytes.Buffer
    if err := tmpl.ExecuteTemplate(&buf, "notfound", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    data.Content = template.HTML(buf.String())

    // Render full page using the standard renderer (base.html + data)
    renderTemplate(w, data)
}
