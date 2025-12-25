# Install

## Important Configuration Note with Non-SSL Setups

If you're running LeoMoon Wiki-Go without SSL/HTTPS and experiencing login issues, you need to set `allow_insecure_cookies: true` in your `config.yaml` file. This is because:

1. By default, LeoMoon Wiki-Go sets the "Secure" flag on cookies for security
2. Browsers reject "Secure" cookies on non-HTTPS connections
3. This prevents login from working properly on HTTP-only setups

> **Security Note**: Only use this setting in development or in trusted internal networks. For public-facing wikis, always use HTTPS.

## Native TLS Configuration

In `data/config.yaml` set:

```yaml
server:
  host: 0.0.0.0
  port: 443            # container listens on 443
  allow_insecure_cookies: false
  ssl: true            # enable built-in HTTPS
  ssl_cert: /path/to/certificate.crt
  ssl_key:  /path/to/private.key
```

If `ssl: false` (default) the app serves plain HTTP on `port` (8080 by default) and you can run it behind a reverse proxy instead.

The Docker image published by GitHub exposes **both** 8080 and 443 so you can choose either scenario at runtime (see below).

---

## Native TLS Configuration

In `data/config.yaml` set:

```yaml
server:
  host: 0.0.0.0
  port: 443            # container listens on 443
  allow_insecure_cookies: false
  ssl: true            # enable built-in HTTPS
  ssl_cert: /path/to/certificate.crt
  ssl_key:  /path/to/private.key
```

If `ssl: false` (default) the app serves plain HTTP on `port` (8080 by default) and you can run it behind a reverse proxy instead.

The Docker image published by GitHub exposes **both** 8080 and 443 so you can choose either scenario at runtime.

---

## Docker (quick test)

```bash
# Pull the latest image
docker pull leomoonstudios/wiki-go

# Run with default configuration
docker run -d \
  --name wiki-go \
  -p 8080:8080 \
  -v "$(pwd)/data:/wiki/data" \
  leomoonstudios/wiki-go
```

## Docker Compose

### Option 1 – Plain HTTP (port 8080)

Use the supplied `docker-compose-http.yml`:

```bash
docker-compose -f docker-compose-http.yml up -d
```

This starts Wiki-Go on http://localhost:8080. Ideal when you terminate TLS at a reverse-proxy (Nginx/Traefik/Caddy). Remember to set `allow_insecure_cookies: true` in `data/config.yaml` if the proxy–>container hop is plain HTTP.

<details>
<summary>Nginx reverse-proxy configuration (click to expand)</summary>

```nginx
server {
    listen 80;
    server_name wiki.example.com;

    # Redirect all HTTP to HTTPS (assuming you use Let's Encrypt on 443)
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name wiki.example.com;

    ssl_certificate     /etc/letsencrypt/live/wiki.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/wiki.example.com/privkey.pem;

    # --- proxy to Wiki-Go container running on HTTP (port 8080) ---
    location / {
        proxy_pass http://wiki-go:8080;

        # Recommended headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
    }
}
```

Compose example for the Nginx service:

```yaml
  nginx:
    image: nginx:alpine
    container_name: wiki-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - wiki-go
```

</details>

### Option 2 – Native HTTPS (port 443)

```bash
# Place certificate + key in ./ssl/
mkdir -p ssl
docker-compose -f docker-compose-ssl.yml up -d
```

`docker-compose-ssl.yml` maps host port 443 → container port 443 and mounts your certificate/key. Enable TLS in the application config.

---

### Binary

Download the latest release for your platform from the [GitHub Releases](https://github.com/leomoon-studios/wiki-go/releases) page.
```bash
# Run the application
./wiki-go  # or wiki-go.exe on Windows
```

### Build from Source

Requirements:
- Go 1.21 or later
- Git

```bash
# Clone the repository
git clone https://github.com/leomoon-studios/wiki-go.git
cd wiki-go

# Build the binary
go build -o wiki-go

# Run the application
./wiki-go  # or wiki-go.exe on Windows
```

### Command-Line Options

| Option        | Description                    | Default                                          |
| ------------- | ------------------------------ | ------------------------------------------------ |
| `-configfile` | Path to the configuration file | `data/config.yaml` (relative to binary location) |

**Example:**

```bash
# Use a custom config file location
./wiki-go -configfile /etc/wiki-go/config.yaml
# Or with a relative path
./wiki-go -configfile ./custom-config.yaml
```

This is useful when:
- Running multiple Wiki-Go instances with different configurations
- Deploying in environments where config should be stored separately from the binary
- Using containerized deployments with mounted config files
