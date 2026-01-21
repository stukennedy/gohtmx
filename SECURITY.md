# Security Model

This document describes the security architecture for Irgo desktop applications.

## Threat Model

Desktop Irgo applications run a local HTTP server on `127.0.0.1` (localhost) to serve the application UI. While this server only binds to the loopback interface and is not accessible from other machines, it faces several potential threats:

### DNS Rebinding Attacks

A malicious website could use DNS rebinding to make requests to your localhost server:
1. User visits `malicious.com` which initially resolves to attacker's IP
2. Attacker's DNS then rebinds to `127.0.0.1`
3. JavaScript on the page can now make requests to your local server

### CSRF via Localhost

Any website the user visits can attempt to make requests to known localhost ports:
- Forms can POST to `http://127.0.0.1:PORT/endpoint`
- Images/scripts can trigger GET requests
- JavaScript can make fetch/XHR requests (limited by CORS)

### Malicious Browser Extensions

Browser extensions with appropriate permissions can:
- Read and modify requests to localhost
- Bypass CORS restrictions
- Access cookies and local storage

### Port Scanning

Malicious websites can probe localhost ports to discover running services and potentially fingerprint applications.

## Mitigations

### 1. Loopback-Only Binding

The server **only** binds to `127.0.0.1`, never to `0.0.0.0` or any external interface. This ensures the server is not accessible from the network.

```go
server.Addr = fmt.Sprintf("127.0.0.1:%d", port)
```

### 2. Per-Launch Secret

Every application launch generates a cryptographically secure random secret using `crypto/rand`. This secret must be included in all requests:

- **HTTP Requests**: `X-Irgo-Secret` header
- **WebSocket Connections**: `?secret=` query parameter (WebSocket API limitation)

The secret is:
- 32 bytes of random data, base64 encoded (43 characters)
- Generated fresh for each application launch
- Injected into the WebView via JavaScript before page load
- Never logged or exposed in URLs (except WebSocket query param, which is unavoidable)

```javascript
// Injected into WebView
window.__IRGO_SECRET__ = "randomSecretHere";
```

This defeats DNS rebinding and CSRF attacks because external sites cannot know the secret.

### 3. Strict Origin Validation

Non-safe HTTP methods (POST, PUT, DELETE, PATCH) require a valid `Origin` header:

- Origin must **exactly match** the application's origin (e.g., `http://127.0.0.1:PORT`)
- No suffix/prefix matching that could be bypassed
- Missing Origin is allowed for same-origin requests from the WebView

```go
// Only exact matches allowed
if origin != allowedOrigin {
    return 403 Forbidden
}
```

### 4. CORS Configuration

CORS headers are set to only allow the application's own origin:

```
Access-Control-Allow-Origin: http://127.0.0.1:PORT
Access-Control-Allow-Credentials: true
```

### 5. Minimal Attack Surface

The server only exposes routes defined by the application:
- No general RPC endpoint
- No file system access by default
- No shell execution by default

## Security Middleware Stack

Middleware is applied in this order (outermost first):

```go
// 1. CORS headers for preflight requests
handler = CORSMiddleware(allowedOrigins)(handler)

// 2. Origin validation for state-changing requests
handler = StrictOriginMiddleware(allowedOrigins)(handler)

// 3. Secret validation (excludes /static/)
handler = SecretValidationMiddleware(secret, []string{"/static/"})(handler)

// 4. WebSocket secret validation (in query param)
handler = WebSocketSecretMiddleware(secret)(handler)
```

## Static Assets

Static assets (`/static/*`) bypass secret validation for performance, but still require valid Origin for non-GET requests. This is safe because:
- GET requests to static assets cannot mutate state
- Static assets are typically cached by the browser

## WebSocket Security

WebSocket connections face additional challenges:
- Browser WebSocket API does not support custom headers
- Secret must be passed as query parameter

The `WebSocketSecretMiddleware` validates the secret during the upgrade handshake before establishing the connection.

## Configuration

### Environment Variable

```bash
# Use in-process transport (no network server)
IRGO_TRANSPORT=inprocess ./myapp
```

### Programmatic Configuration

```go
config := desktop.Config{
    Transport: "loopback",  // Default - localhost HTTP server
    // or
    Transport: "inprocess", // No network - for testing mobile code path
}
```

## Testing Security

### Verify Secret Rejection

```bash
# Should fail (no secret)
curl http://127.0.0.1:PORT/api/data

# Should fail (wrong secret)
curl -H "X-Irgo-Secret: wrong" http://127.0.0.1:PORT/api/data
```

### Verify Origin Rejection

```bash
# Should fail (wrong origin)
curl -X POST -H "Origin: http://evil.com" http://127.0.0.1:PORT/api/data
```

## Known Limitations

1. **WebSocket Secret in URL**: The secret appears in the WebSocket URL query parameter. This is logged by some proxies/tools but is unavoidable due to browser API limitations.

2. **Port Visibility**: The application's port is visible to any local process. The secret prevents unauthorized access.

3. **Memory-Based Attacks**: A malicious process with memory access to the WebView could extract the secret. This is outside the threat model (assumes compromised local machine).

## Reporting Security Issues

If you discover a security vulnerability, please report it privately by emailing [security contact] rather than opening a public issue.
