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
- **File Upload Validation**: MIME type checking for uploaded files (can be disabled if needed).
- **Private Wiki Mode**: Option to require authentication for all pages.
- **Login Rate Limiting**: Built-in protection against brute force attacks by temporarily banning IP addresses after multiple failed login attempts, with exponential backoff.

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

Roles are hierarchical, admins bypass all access rule restrictions and always have full access.

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
3. Admins always have access regardless of rules

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
- **Maximum Failures**: 3 (failures before triggering a ban)
- **Window Time**: 30 seconds (time window in which failures are counted)
- **Initial Ban Duration**: 60 seconds (length of the first ban)
- **Maximum Ban Duration**: 86400 seconds (24 hours, upper limit for exponential backoff)

### User Experience

1. First 3 failures → Standard error message ("Invalid username or password")
2. After 3 failures → 1-minute ban with message "Too many failed login attempts; try again later"
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
- **Secure Cookies**: Session tokens are transmitted via `HttpOnly`, `SameSite=Strict` cookies, preventing XSS and CSRF attacks.

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
9. **Use a reverse proxy** like Nginx, Caddy, or Traefik for additional security layers.
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

## Known Issues

No known security issues at this time.

## Security Contact

For security concerns, please use our [contact form](https://leomoon.com/contact).