# CASCI API Documentation

## Authentication

CASCI supports three authentication methods:

1. **JWT Token** (Recommended for web applications)
   - Header: `Authorization: Bearer <token>`
   - Expires after 24 hours
   - Can be refreshed

2. **API Token** (Recommended for CI/CD integrations)
   - Header: `X-API-Token: <token>`
   - Never expires (until regenerated)
   - User-specific

3. **Query Parameter** (For webhooks and simple integrations)
   - URL: `?token=<api-token>`
   - Same as API Token

## API Endpoints

### Authentication & Registration

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "john",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com",
    "api_token": "casci_abc123...",
    "is_admin": true,
    "created_at": "2025-09-29T00:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Notes:**
- First user to register becomes administrator
- Returns both JWT token and API token
- Password must be at least 8 characters
- Username must be at least 3 characters

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "john",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com",
    "api_token": "casci_abc123...",
    "is_admin": true,
    "created_at": "2025-09-29T00:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Authorization: Bearer <existing-token>
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### User Management

#### Get Current User
```http
GET /api/v1/users/me
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 1,
  "username": "john",
  "email": "john@example.com",
  "api_token": "casci_abc123...",
  "is_admin": true,
  "created_at": "2025-09-29T00:00:00Z"
}
```

#### Regenerate API Token
```http
POST /api/v1/users/me/token
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 1,
  "username": "john",
  "email": "john@example.com",
  "api_token": "casci_xyz789...",
  "is_admin": true,
  "created_at": "2025-09-29T00:00:00Z"
}
```

#### List All Users (Admin Only)
```http
GET /api/v1/users
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": 1,
    "username": "john",
    "email": "john@example.com",
    "is_admin": true,
    "created_at": "2025-09-29T00:00:00Z"
  },
  {
    "id": 2,
    "username": "jane",
    "email": "jane@example.com",
    "is_admin": false,
    "created_at": "2025-09-29T01:00:00Z"
  }
]
```

#### Get User by ID
```http
GET /api/v1/users/1
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 1,
  "username": "john",
  "email": "john@example.com",
  "is_admin": true,
  "created_at": "2025-09-29T00:00:00Z"
}
```

#### Update User
```http
PUT /api/v1/users/1
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "newemail@example.com",
  "password": "newpassword123"
}
```

**Response:**
```json
{
  "id": 1,
  "username": "john",
  "email": "newemail@example.com",
  "is_admin": true,
  "created_at": "2025-09-29T00:00:00Z"
}
```

**Notes:**
- Users can update their own profile
- Admins can update any user
- Both fields are optional

#### Delete User (Admin Only)
```http
DELETE /api/v1/users/2
Authorization: Bearer <token>
```

**Response:**
- Status: 204 No Content

### System

#### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy"
}
```

#### API Info
```http
GET /api/v1/
```

**Response:**
```json
{
  "message": "CASCI API v1",
  "version": "1.0.0"
}
```

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message"
}
```

### Common HTTP Status Codes

- `200 OK` - Request succeeded
- `201 Created` - Resource created successfully
- `204 No Content` - Request succeeded with no response body
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required or failed
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `500 Internal Server Error` - Server error

## Usage Examples

### cURL Examples

#### Register and Login
```bash
# Register first user (becomes admin)
curl -X POST http://localhost:64500/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@example.com","password":"admin123456"}'

# Login
curl -X POST http://localhost:64500/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}'

# Save the token from response
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Get current user
curl -X GET http://localhost:64500/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"

# Regenerate API token
curl -X POST http://localhost:64500/api/v1/users/me/token \
  -H "Authorization: Bearer $TOKEN"
```

#### Using API Token
```bash
# Save the API token from registration/profile
API_TOKEN="casci_abc123..."

# Get current user with API token
curl -X GET http://localhost:64500/api/v1/users/me \
  -H "X-API-Token: $API_TOKEN"

# Or use query parameter
curl -X GET "http://localhost:64500/api/v1/users/me?token=$API_TOKEN"
```

### JavaScript/Fetch Example

```javascript
// Register
const response = await fetch('http://localhost:64500/api/v1/auth/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'john',
    email: 'john@example.com',
    password: 'securepass123'
  })
});

const { user, token } = await response.json();
localStorage.setItem('token', token);

// Get current user
const meResponse = await fetch('http://localhost:64500/api/v1/users/me', {
  headers: { 'Authorization': `Bearer ${token}` }
});

const currentUser = await meResponse.json();
```

### Python Example

```python
import requests

# Register
response = requests.post('http://localhost:64500/api/v1/auth/register', json={
    'username': 'john',
    'email': 'john@example.com',
    'password': 'securepass123'
})

data = response.json()
token = data['token']
api_token = data['user']['api_token']

# Get current user with JWT
response = requests.get('http://localhost:64500/api/v1/users/me',
    headers={'Authorization': f'Bearer {token}'})

# Or with API token
response = requests.get('http://localhost:64500/api/v1/users/me',
    headers={'X-API-Token': api_token})
```

## Security Considerations

1. **Passwords**: Hashed with bcrypt (cost 10)
2. **JWT Tokens**: Signed with HS256, expire after 24 hours
3. **API Tokens**: Cryptographically random, 64 hex characters
4. **HTTPS**: Enable TLS in production
5. **Rate Limiting**: Coming soon
6. **CORS**: Configure for your domains

## Next Steps

- **Projects API**: Create and manage CI/CD projects
- **Builds API**: Trigger and monitor builds
- **Pipelines**: Define build workflows
- **Webhooks**: Automatic build triggers

See [DEVELOPMENT.md](./DEVELOPMENT.md) for development roadmap.