# WebSocket Chat Demo

This application demonstrates how to implement secure WebSocket authentication using JWT tokens with the [Gorilla WebSocket](https://github.com/gorilla/websocket) package. The demo includes an interactive UI that visualizes the authentication flow step-by-step.

Based on the [Gorilla WebSocket Chat Example](https://github.com/gorilla/websocket/tree/main/examples/chat) with added JWT authentication layer.

## Authentication Implementation

### Current Implementation: JWT via Query Parameter

This demo uses **JWT token authentication via query string parameter** during the WebSocket upgrade request.

**How it works:**
1. Client requests a JWT token from `GET /api/auth/token`
2. Server generates a JWT containing guest identity and expiration (24 hours)
3. Client connects to WebSocket with token: `ws://localhost:8080/ws?token=<jwt_token>`
4. Server validates token during connection upgrade before establishing WebSocket
5. If valid, server accepts connection and assigns guest identity

**Advantages:**
- ✅ Simple to implement
- ✅ Works in all browsers (no custom headers needed)
- ✅ Token validated before WebSocket upgrade
- ✅ Standard HTTP authentication flow

**Disadvantages:**
- ⚠️ Token visible in URL (though only during upgrade, not in browser address bar)
- ⚠️ May appear in server logs if not configured carefully

### Alternative Authentication Methods

Based on [WebSocket Authentication Best Practices](https://websockets.readthedocs.io/en/latest/topics/authentication.html):

#### 1. **Cookie-Based Authentication**
```go
// Server side
http.SetCookie(w, &http.Cookie{
    Name:     "auth_token",
    Value:    token,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteStrictMode,
})

// In WebSocket upgrade
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("auth_token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // Validate cookie.Value...
}
```
**Pros:** Most secure, automatic browser handling, HttpOnly prevents XSS  
**Cons:** Requires cookie setup, CSRF considerations

#### 2. **First Message Authentication**
```go
// Client sends token as first message after connection
conn.send("AUTH:" + token);

// Server waits for auth message before processing
func (c *Client) readPump() {
    authenticated := false
    for {
        _, message, err := c.conn.ReadMessage()
        if !authenticated {
            if bytes.HasPrefix(message, []byte("AUTH:")) {
                token := string(bytes.TrimPrefix(message, []byte("AUTH:")))
                // Validate token...
                authenticated = true
                continue
            }
            return // Disconnect if no auth
        }
        // Process normal messages...
    }
}
```
**Pros:** Token not in URL or headers  
**Cons:** More complex, connection established before auth, needs timeout handling

#### 3. **Subprotocol-Based Authentication**
```javascript
// Client
const ws = new WebSocket('ws://localhost:8080/ws', ['auth.token.' + token]);

// Server
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
    subprotocols := r.Header.Get("Sec-WebSocket-Protocol")
    // Parse and validate token from subprotocol
}
```
**Pros:** Clean separation, part of WebSocket spec  
**Cons:** Limited browser support, complex parsing

## Running the Example

The example requires a working Go development environment. The [Getting
Started](http://golang.org/doc/install) page describes how to install the
development environment.

Once you have Go up and running, you can download dependencies and run:

    $ go mod tidy
    $ go run *.go

To use the chat, open http://localhost:8080/ in your browser.

## API Endpoints

### GET/POST `/api/auth/token`

Generates a JWT token for guest authentication.

**Response:**
```json
{
  "token": "eyJhbGci...",
  "guest_name": "guest-d3e6",
  "expires_at": 1764671496
}
```

### WebSocket `/ws`

WebSocket endpoint for real-time chat. **Requires authentication via query parameter.**

**Connection URL:**
```
ws://localhost:8080/ws?token=<jwt_token>
```

**Authentication Flow:**
1. Token is extracted from query parameter
2. Token is validated (signature, expiration)
3. Guest identity is extracted from token claims
4. WebSocket upgrade is completed
5. Client receives `IDENTITY:<guest_name>` message

**Supported Authentication Methods in Code:**
- ✅ Query parameter: `?token=<jwt_token>` (active)
- ⚠️ Authorization header: `Authorization: Bearer <jwt_token>` (implemented but not used by browser WebSocket API)

> **Note:** Browser WebSocket API doesn't support custom headers. The Authorization header method is implemented server-side but cannot be used from browsers. Use query parameter instead.

## References

- [Gorilla WebSocket Package](https://github.com/gorilla/websocket)
- [Gorilla WebSocket Chat Example](https://github.com/gorilla/websocket/tree/main/examples/chat)
- [WebSocket Authentication Best Practices](https://websockets.readthedocs.io/en/latest/topics/authentication.html)
- [JWT (JSON Web Tokens)](https://jwt.io/)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
