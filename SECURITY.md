# Security Policy

## Supported Versions

Use this section to tell people about which versions of Wiki-Go are currently being supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

## Reporting a Vulnerability

We take the security of Wiki-Go seriously. If you believe you've found a security vulnerability, please follow these steps:

1. **Do not disclose the vulnerability publicly** or on the public issue tracker.
2. Submit your findings through our [contact form](https://leomoon.com/contact).
3. Allow time for us to review and address the vulnerability before any public disclosure.
4. We'll respond as quickly as possible to acknowledge receipt of your report.

## Security Features

Wiki-Go includes several security features:

- **Password Storage**: All passwords are hashed using bcrypt with appropriate cost factors.
- **Authentication**: Session-based authentication with secure, HTTP-only cookies and hashed token persistence.
- **TLS Support**: Built-in TLS support for encrypted connections.
- **Role-Based Access Control**: Fine-grained permissions through admin, editor, and viewer roles.
- **Path-Based Access Rules**: Granular document access control using glob patterns, access levels, and group membership.
- **Safe Markdown Rendering**: User-authored inline HTML, CSS, and JavaScript are not supported in Markdown documents or comments.
- **Trusted Markdown Extensions**: Wiki-Go extensions render through typed components that escape text, normalize identifiers, and validate URLs before producing markup.
- **SVG Upload Protection**: SVG uploads are parsed through a strict XML allowlist even when general MIME checking is disabled.
- **Content Security Policy**: Wiki-Go enforces a restrictive page policy that allows scripts only from approved application sources and blocks inline script handlers.
- **File Upload Validation**: MIME type checking protects uploaded files, with SVG safety checks remaining mandatory when general MIME checking is disabled.
- **Private Wiki Mode**: Option to require authentication for all pages.
- **Login Rate Limiting**: Built-in protection against brute force attacks by temporarily banning IP addresses after multiple failed login attempts, with exponential backoff.
- **Session Revocation**: Security-sensitive account changes revoke the affected user's active sessions in memory and persistent storage.
- **Trusted Proxy Boundary**: Forwarded client IP headers are ignored unless the direct proxy is explicitly trusted.
- **Log Record Integrity**: Control characters from untrusted values are escaped at the centralized logging boundary so one logging call cannot forge additional records or terminal output.

## Security Logging

Wiki-Go formats log messages and neutralizes carriage returns, newlines, Unicode line separators, and terminal control characters before writing each record. This boundary applies consistently to debug, informational, warning, error, and fatal logs, including formatted and wrapped errors.

Centralized neutralization protects log structure, but it does not make secrets safe to log. Callers must never include passwords, session tokens or token hashes, cookies, authorization headers, API keys, uploaded file contents, or other credential material in log messages. Non-secret request paths, usernames, filenames, network errors, and configuration values may be logged when operationally useful because control characters are escaped at the logging boundary.

## Markdown and Content Rendering Security

Wiki-Go treats Markdown as content rather than executable page markup. User-authored inline HTML, CSS, and JavaScript in Markdown documents and comments are unsupported and are not rendered as active application content. Raw HTML is omitted or displayed inertly instead of being passed through to the browser.

This restriction applies consistently to normal document pages, the editor preview, version-history previews, home-page content, comments, kanban documents, and links documents.

### Trusted Wiki-Go Extensions

Wiki-Go's built-in Markdown extensions do not depend on user-provided raw HTML. Supported features are parsed into trusted, typed nodes or escaped templates controlled by Wiki-Go.

These rendering boundaries provide the following protections:

- Element and attribute names are selected by Wiki-Go rather than supplied by Markdown authors.
- User-controlled text is HTML-escaped before insertion into trusted output.
- Identifiers and CSS class tokens generated from content are normalized.
- Links and media destinations are restricted to validated local URLs or approved URL schemes.
- Missing internal document and attachment links retain their visual `notfound` indicator through a trusted renderer.
- Render-specific state is isolated so concurrent document rendering does not share placeholder content.

Supported Markdown and Wiki-Go extensions remain available, including tables, task lists, code blocks, GitHub-style admonitions, Mermaid diagrams, kanban boards, links pages, wiki links, attachments, details blocks, infoboxes, highlighting, superscript, subscript, table-of-contents generation, heading anchors, supported media embeds, and Wiki-Go custom tags.

### Defense in Depth

The safe Markdown boundary is reinforced by an enforcing Content Security Policy. Application scripts are loaded from approved sources, inline event handlers are blocked, and embedded objects are disabled.

SVG files receive separate upload and serving protections. Uploaded SVG content is parsed using a minimal allowlist that excludes executable or externally loaded content, and SVG responses use restrictive security headers.

Administrators upgrading from an earlier release should review documents that contain custom HTML, `<style>` blocks, inline styles, scripts, or HTML-based components. These constructs are no longer supported inside Markdown and should be replaced with standard Markdown, a supported Wiki-Go extension, or administrator-controlled application styling.

## Role-Based Access Control

Wiki-Go implements a hierarchical role system combined with path-based access rules for comprehensive access management.

### User Roles

Each user is assigned one of three roles:

| Permission          |       Admin        |       Editor       |       Viewer       |
| ------------------- | :----------------: | :----------------: | :----------------: |
| View documents      | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Add documents       | :white_check_mark: | :white_check_mark: |                    |
| Edit documents      | :white_check_mark: | :white_check_mark: |                    |
| Delete documents    | :white_check_mark: | :white_check_mark: |                    |
| Move documents      | :white_check_mark: | :white_check_mark: |                    |
| Manage versions     | :white_check_mark: | :white_check_mark: |                    |
| Post comments       | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Delete comments     | :white_check_mark: |                    |                    |
| Manage users        | :white_check_mark: |                    |                    |
| Manage access rules | :white_check_mark: |                    |                    |
| Manage settings     | :white_check_mark: |                    |                    |

Roles are hierarchical. Administrators bypass access-rule restrictions for valid document identities; malformed paths and identity-resolution failures are rejected for every role.

### User Groups

Users can be assigned to one or more groups for fine-grained access control:

```yaml
users:
  - username: alice
    role: editor
    groups: [finance, hr]
  - username: bob
    role: viewer
    groups: [finance]
```

Groups work in conjunction with access rules to restrict document visibility.

### Path-Based Access Rules

Access rules define who can view specific documents or folders based on URL path patterns.

The same access rules are enforced for rendered pages and all document-scoped operations, including attachment, document, move, and version APIs. An editor role grants editing capabilities only for paths that the editor can access through public, private, or group-restricted rules. Administrators retain access to valid document identities.

Document request paths are decoded once by HTTP handling, validated, and resolved to a logical authorization identity and filesystem path together. Source, save, move, and version requests reject parent traversal, absolute filesystem paths, residual encoding, and traversal written with either forward slashes or backslashes. Homepage operations select the separate `pages/home` storage root. Creation destinations resolve existing parents before adding validated missing components, and authorization applies to the destination and its parents. Recursive move and delete operations also check descendant access.

#### Access Levels

| Who can view          |       Public       |      Private       |     Restricted     |
| --------------------- | :----------------: | :----------------: | :----------------: |
| Unauthenticated users | :white_check_mark: |                    |                    |
| Authenticated users   | :white_check_mark: | :white_check_mark: |                    |
| Group members         | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Admin users           | :white_check_mark: | :white_check_mark: | :white_check_mark: |

#### Pattern Matching

Rules use glob-style patterns:

| Pattern       | Matches                                            |
| ------------- | -------------------------------------------------- |
| `/finance/**` | `/finance`, `/finance/reports`, `/finance/2024/q1` |
| `/docs/*`     | `/docs/readme` (single level only)                 |
| `/secret`     | Exactly `/secret`                                  |

#### Rule Evaluation

1. Rules are evaluated in order (first match wins)
2. If no rule matches:
   - **Private wiki**: Authenticated users only
   - **Public wiki**: Everyone has access
3. Admins bypass rules for valid, resolved document identities; invalid paths and lookup failures are denied

#### Example Configuration

```yaml
access_rules:
  - pattern: "/finance/**"
    access: restricted
    groups: [finance, executives]
    description: "Financial documents - finance team only"
  
  - pattern: "/internal/**"
    access: private
    description: "Internal docs - any authenticated user"
  
  - pattern: "/public/**"
    access: public
    description: "Public documentation"
```

### Managing Access Rules

Access rules are managed through the **Admin Interface** under **Settings → Access Rules** tab. From there you can:

- Create new rules with pattern, access level, groups, and description
- Edit existing rules
- Delete rules
- Reorder rules via drag-and-drop (important since first match wins)

Rules are automatically saved to `config.yaml`, manual editing of the config file is not required.

## Windows Path Identity and Release Status

Wiki-Go v1.9.6 is confirmed affected by a Windows path authorization bypass. On the tested Windows filesystem, mixed-case names, discovered 8.3 short names, and trailing-dot aliases could reach restricted content under a different request spelling. Nested attachments were also affected because filesystem separators could reach slash-based access-rule matching. The unfixed development baseline at `c70e950` used the same affected access matcher, attachment handler, and path helpers as v1.9.6. This validation does not establish the earliest affected release.

The correction is in unreleased source through `97f522a`, followed by the full-suite CI and integration coverage in this change. No fixed release version is assigned here. A deployment must contain these fixes; a version label of v1.9.6 by itself does not include them.

On Windows, existing resource names resolve through filesystem handles to their stored long names before authorization. Supported mixed-case and 8.3 aliases share the canonical document's rules, comments, and history. Logical separators normalize to `/`, and Windows access-rule pattern matching is case insensitive. Group membership and rule order retain their existing semantics. Document and attachment reads and mutations use the corresponding resolved filesystem paths. Attachment file aliases cannot change their owning directory, and internal comment/history storage cannot remap to another document's namespace. SVG sanitization, Markdown attachment blocking, MIME checks, and download headers remain enforced.

Windows request validation rejects trailing-dot and trailing-space components, alternate data stream colons, device and UNC namespace forms, reserved device names, invalid filename characters, null bytes, traversal, and residual percent encoding. These checks also apply to administrators. Filesystem identity lookup failures are denied rather than falling back to the request spelling. A configured storage root may be resolved as a deployment path, but URL input cannot select an alternate filesystem namespace.

Linux lookup and access-rule matching retain case-sensitive behavior. Tests create distinct `finance` and `Finance` directories and verify that authorization and mutations remain separate. Validation covers Windows Server 2025 with working 8.3 aliases and Ubuntu 26.04 with case-sensitive fixture directories. It does not establish equivalent protection for macOS or every Linux filesystem configuration, including case-insensitive mounts. Resolution also does not make later filesystem opens or mutations atomic against concurrent replacement by a local process; protect storage from untrusted local modification.

### Native Windows Evidence

The [unfixed baseline run](https://github.com/leomoon-studios/wiki-go/actions/runs/45884552315) recorded document and attachment marker disclosure. The [post-fix API run](https://github.com/leomoon-studios/wiki-go/actions/runs/45896789605) passed the complete auth, comments, utils, handlers, and routes test packages on native Windows and Ubuntu. Windows discovered actual aliases, including `finance-quarterly-2026 -> FINANC~1`, rather than assuming a short name. The tests checked denied requests for marker absence and verified authorized/public reads and mutation behavior.

The [full-suite validation run](https://github.com/leomoon-studios/wiki-go/actions/runs/45898630279) passed `go test ./... -count=1 -v` and all 25 login-redirect, search-UI, and safe-DOM Node tests on Windows Server 2025 and Ubuntu 26.04, using Go 1.27.3 and Node 24.22.3. Windows passed 223 top-level Go tests and Ubuntu passed 212. The live HTTP integration exercised registered routes and real sessions in isolated public/private wikis: canonical and alias requests denied outsiders without marker disclosure, public and authorized reads succeeded, denied mutations preserved content, and member mutations succeeded. Both HTTP fixtures discovered and exercised `finance-quarterly-2026 -> FINANC~1`. No applicable Windows alias test was skipped; the only Windows skip was the Linux-only distinct-case control, which passed on Ubuntu.

Short-name tests call `GetShortPathName` and confirm filesystem identity with `os.SameFile`. They explicitly skip when the fixture has no 8.3 alias, without changing volume settings. A successful cross-compilation or an 8.3 test skip is not runtime evidence of short-name protection. CI preserves verbose Go logs and Node test logs as per-platform artifacts so alias discovery and skips can be inspected.

## Login Rate Limiting

Wiki-Go includes built-in protection against brute force attacks by temporarily banning IP addresses after multiple failed login attempts.

### How It Works

1. **Monitoring Failed Attempts**: The system tracks failed login attempts by IP address.
2. **Exponential Backoff**: Ban durations double with each subsequent failure, providing increasing protection against persistent attacks.
3. **Configurable Parameters**: All aspects of the rate limiting system can be customized via the admin interface.
4. **Persistence**: Ban data is stored in `data/temp/login_ban.json` and persists across application restarts.

### Default Settings

The login ban system is enabled by default with the following settings:

- **Enabled**: Yes
- **Maximum Failures**: 5 (failures before triggering a ban)
- **Window Time**: 180 seconds (time window in which failures are counted)
- **Initial Ban Duration**: 60 seconds (length of the first ban)
- **Maximum Ban Duration**: 86400 seconds (24 hours, upper limit for exponential backoff)

### User Experience

1. The first failures receive the standard "Invalid username or password" response.
2. The fifth failure within the default three-minute window triggers a one-minute ban.
3. After ban expires, next failure → 2-minute ban (doubling each time)
4. Ban durations continue doubling up to the configured maximum
5. Successful login resets all ban state for that IP address

### Configuration

Administrators can adjust the login ban settings through:

1. **Admin Interface**: Settings → Security tab
2. **Config File**: Edit the `security` section in `config.yaml`

Example `config.yaml` section:

```yaml
security:
  login_ban:
    enabled: true
    max_failures: 5
    window_seconds: 180
    initial_ban_seconds: 60
    max_ban_seconds: 86400  # 24 hours
```

### Error Messages

- Regular failed login: "Invalid username or password"
- Banned state: "Too many failed login attempts; try again later"
- When banned, the client also receives HTTP status 429 (Too Many Requests) with a "Retry-After" header

### Reverse Proxies and Client IP Addresses

Wiki-Go uses the direct network peer for login throttling by default. `X-Forwarded-For` and `X-Real-IP` are accepted only when the direct peer matches an exact IP address or CIDR network configured under `server.trusted_proxies`.

Instances that are not behind a reverse proxy should leave the list empty:

```yaml
server:
  trusted_proxies: []
```

For standalone Nginx running on the same machine and proxying to Wiki-Go through an IPv4 loopback address such as `proxy_pass http://127.0.0.1:3030`, trust only that loopback address:

```yaml
server:
  trusted_proxies:
    - "127.0.0.1"
```

Add `::1` only if the proxy connects through IPv6 loopback. Container deployments should trust the exact proxy address or a dedicated container-network CIDR rather than a broad private network.

## Session Security

Wiki-Go implements secure session management with persistence capabilities.

### Storage and Persistence

- **File-Based Storage**: Active sessions are persisted to `data/temp/sessions.json` to maintain login state across application restarts.
- **Token Hashing**: Session tokens are hashed using SHA256 before being stored on disk. This provides a critical security layer:
  - The browser holds the raw token (the "key").
  - The server stores the hashed token (the "lock").
  - If the session file is compromised, attackers only obtain the hashes, which cannot be used to authenticate requests.

### Session Lifecycle

- **Expiration**: Standard sessions expire after 24 hours. "Keep me logged in" sessions persist for 30 days.
- **Automatic Cleanup**: The system automatically purges expired sessions from both memory and disk to maintain hygiene and security.
- **Account Changes**: Password changes, administrator password resets, role demotions, group-access changes, and account deletion revoke affected sessions.
- **Password Change Experience**: A successful self-service password change clears the browser session and redirects the user to log in again after confirmation.
- **Secure Cookies**: Session tokens are transmitted using `HttpOnly`, `Secure` by default, and `SameSite=Lax` cookie protections.

## Security Recommendations

For secure deployment of Wiki-Go, we recommend:

1. **Always use HTTPS** in production environments.
2. **Set `allow_insecure_cookies: false`** (the default) to enforce secure cookies.
3. **Change the default admin password** immediately after installation.
4. **Set strong passwords** for all accounts, especially admin accounts.
5. **Enable login rate limiting** through the security settings to prevent brute force attacks.
6. **Configure access rules** for sensitive documents, use restricted access with groups for confidential content.
7. **Regularly review access rules** to ensure rule order and group assignments are correct.
8. **Regularly update** to the latest version for security patches.
9. **Use a reverse proxy** like Nginx, Caddy, or Traefik for additional security layers, and configure only its direct address under `server.trusted_proxies` when forwarded client IPs are needed.
10. **Back up your data** regularly to prevent data loss.
11. **Set appropriate file upload size limits** to prevent denial of service attacks.
12. **Regularly review user accounts and group memberships** to ensure only authorized users have access.

## Dependency Management

Wiki-Go uses Go modules for dependency management. All dependencies are vendored to ensure reproducible builds.

## Security Practices

Our security practices include:

1. Regular code review with a focus on security
2. Input validation to prevent injection attacks
3. Proper error handling to avoid information leakage
4. Use of standard libraries for cryptographic operations
5. Secure session management
6. Principle of least privilege for user roles
7. Safe-by-default Markdown rendering with trusted extension components
8. Defense-in-depth browser and attachment security policies

## Known Issues

The Windows path authorization bypass confirmed in v1.9.6 is addressed in the unreleased source described above. A fixed release version has not been assigned by this change. Filesystem configurations outside the documented validation scope require separate verification.

## Security Contact

For security concerns, please use our [contact form](https://leomoon.com/contact).
