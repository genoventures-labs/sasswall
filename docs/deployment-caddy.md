# Sasswall deployment (Caddy)

This document provides Caddy patterns for routing unwanted traffic to Sasswall.

> Assumption: Sasswall listens on `127.0.0.1:8182` and is only reachable from the reverse proxy.

---

## Pattern A: HTTP :80 catch-all sink

Use Sasswall as the default for port 80 (useful for by-IP traffic and random bots).

```caddyfile
:80 {
  # Optionally keep known hosts redirecting to HTTPS
  @known host chat.example.com benchmark.example.com
  handle @known {
    redir https://{host}{uri} 308
  }

  # Everything else on :80 goes to Sasswall
  handle {
    reverse_proxy 127.0.0.1:8182
  }
}
```

---

## Pattern B: Unknown host handling on HTTPS

If your TLS strategy permits it, you can return a strict response for unknown hosts.

```caddyfile
:443 {
  @unknown not host chat.example.com benchmark.example.com
  handle @unknown {
    respond "Forbidden\n" 403
  }
}
```

> Many environments prefer hard 403 on unknown TLS SNI rather than proxying to an application.

---

## Pattern C: Route explicit honey paths to Sasswall

You can direct a subset of paths to Sasswall even on valid hosts.

```caddyfile
chat.example.com {
  @honey path /.env /.git/* /wp-login.php /phpmyadmin*
  handle @honey {
    reverse_proxy 127.0.0.1:8182
  }

  # normal app routes
  handle {
    reverse_proxy 127.0.0.1:8080
  }
}
```

---

## Operational notes

- Ensure Caddy passes real client IP if behind a CDN/WAF (e.g., `CF-Connecting-IP`).
- Keep Sasswall responses cache-disabled (Sasswall sets `Cache-Control: no-store`).
- Consider adding dedicated access logs for the Sasswall route for easier ingestion.
