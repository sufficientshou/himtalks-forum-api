# HimTalks Backend API Specification

## 🎯 Base Information
- **Base URL**: `https://api.teknohive.me`
- **Authentication**: JWT Cookie-based (HTTP-only cookies)
- **Content-Type**: `application/json`
- **CORS**: Enabled for specific domains

---

## 🔐 Authentication Endpoints

### Google OAuth Login
```http
GET /auth/google/login
```
**Description**: Redirect user to Google OAuth consent screen
**Response**: Redirects to Google OAuth

### Google OAuth Callback
```http
GET /auth/google/callback
```
**Description**: Handle Google OAuth callback and set JWT cookie
**Response**: Redirects to frontend with authentication

### Logout
```http
POST /auth/logout
```
**Description**: Clear authentication cookie
**Response**: 200 OK

### Check Authentication Status
```http
GET /api/protected
```
**Description**: Check if user is authenticated
**Response**: User email string
**Auth Required**: Yes

---

## 📨 Public Message Endpoints

### Send Message (Kritik/Saran)
```http
POST /message
```
**Body**:
```json
{
  "content": "Your message content",
  "sender_name": "John Doe",
  "recipient_name": "Admin",
  "category": "kritik" // or "saran"
}
```
**Description**: Send anonymous message (kritik or saran)
**Response**: Created message object
**Validation**: 
- Content length limit (configurable)
- Blacklist word checking
- Category must be "kritik" or "saran"

---

## 🎵 Public Songfess Endpoints

### Send Songfess
```http
POST /songfess
```
**Body**:
```json
{
  "content": "Your message content",
  "song_id": "4iV5W9uYEdYUVa79Axb7Rh",
  "song_title": "Shape of You",
  "artist": "Ed Sheeran",
  "album_art": "https://i.scdn.co/image/...",
  "preview_url": "https://p.scdn.co/mp3-preview/...",
  "start_time": 30,
  "end_time": 60,
  "sender_name": "Alice",
  "recipient_name": "Bob"
}
```
**Description**: Send songfess with Spotify track
**Response**: Created songfess object

### Get Recent Songfess (Public)
```http
GET /songfess
```
**Description**: Get songfess within time limit (default 7 days)
**Response**: Array of songfess objects
**Query Parameters**: None (uses server configuration)

### Get Songfess by ID
```http
GET /songfess/{id}
```
**Description**: Get specific songfess by ID
**Response**: Single songfess object

---

## 🎵 Spotify Integration

### Search Tracks
```http
GET /api/spotify/search?q={query}
```
**Description**: Search Spotify tracks
**Query Parameters**:
- `q`: Search query string
**Response**: Spotify search results

### Get Track Details
```http
GET /api/spotify/track?id={track_id}
```
**Description**: Get detailed track information
**Query Parameters**:
- `id`: Spotify track ID
**Response**: Track details

---

## 🔒 Admin Endpoints

> **Auth Required**: All admin endpoints require Google OAuth login + Admin privileges

### Message Management

#### Get All Messages
```http
GET /api/admin/messages
```
**Description**: Get all messages (kritik & saran)
**Response**: Array of message objects
```json
[
  {
    "id": 1,
    "content": "Great app!",
    "sender_name": "John",
    "recipient_name": "Admin", 
    "category": "saran",
    "created_at": "2025-07-25T10:30:00Z"
  }
]
```

#### Delete Message
```http
POST /api/admin/message/delete
```
**Body**:
```json
{
  "ID": 1
}
```
**Response**:
```json
{
  "message": "Message deleted successfully",
  "id": 1
}
```

### Songfess Management

#### Get All Songfess (No Time Limit)
```http
GET /api/admin/songfessAll
```
**Description**: Get all songfess regardless of age
**Response**: Array of all songfess objects

#### Get Recent Songfess (Admin View)
```http
GET /api/admin/songfess
```
**Description**: Get songfess within configured time limit
**Response**: Array of recent songfess objects

#### Delete Songfess
```http
POST /api/admin/songfess/delete
```
**Body**:
```json
{
  "ID": 1
}
```
**Response**:
```json
{
  "message": "Songfess deleted successfully",
  "id": 1
}
```

### Admin User Management

#### Get Admin List
```http
GET /api/admin/list
```
**Response**: Array of admin email addresses

#### Add New Admin
```http
POST /api/admin/addAdmin
```
**Body**:
```json
{
  "Email": "newadmin@example.com"
}
```

#### Remove Admin
```http
POST /api/admin/removeAdmin
```
**Body**:
```json
{
  "Email": "admin@example.com"
}
```

### System Configuration

#### Get All Configs
```http
GET /api/admin/configs
```
**Response**: Object with configuration key-value pairs
```json
{
  "songfess_days": "7",
  "message_char_limit": "500"
}
```

#### Update Songfess Days Limit
```http
POST /api/admin/configSongfessDays
```
**Body**:
```json
{
  "Days": "7"
}
```

### Blacklist Management

#### Get Blacklisted Words
```http
GET /api/admin/blacklist
```
**Response**: Array of blacklisted words

#### Add Blacklisted Word
```http
POST /api/admin/blacklist
```
**Body**:
```json
{
  "Word": "badword"
}
```

#### Remove Blacklisted Word
```http
POST /api/admin/blacklist/remove
```
**Body**:
```json
{
  "Word": "badword"
}
```

---

## 🌐 WebSocket Real-time Updates

### Connection
```
wss://api.teknohive.me/ws
```

### Message Types
- `message`: New message created
- `songfess`: New songfess created  
- `delete_message`: Message deleted by admin
- `delete_songfess`: Songfess deleted by admin

### Example WebSocket Message
```json
{
  "type": "songfess",
  "data": {
    "id": 1,
    "content": "Love this song!",
    "song_title": "Shape of You",
    "artist": "Ed Sheeran",
    // ... other songfess fields
  }
}
```

---

## 📋 Data Models

### Message Model
```json
{
  "id": 1,
  "content": "Message content",
  "sender_name": "John Doe",
  "recipient_name": "Admin",
  "category": "kritik", // or "saran"
  "created_at": "2025-07-25T10:30:00Z"
}
```

### Songfess Model
```json
{
  "id": 1,
  "content": "Message content",
  "song_id": "4iV5W9uYEdYUVa79Axb7Rh",
  "song_title": "Shape of You", 
  "artist": "Ed Sheeran",
  "album_art": "https://i.scdn.co/image/...",
  "preview_url": "https://p.scdn.co/mp3-preview/...",
  "start_time": 30,
  "end_time": 60,
  "sender_name": "Alice",
  "recipient_name": "Bob",
  "created_at": "2025-07-25T10:30:00Z"
}
```

---

## ⚠️ Error Responses

### Standard Error Format
```json
{
  "error": "Error message description"
}
```

### HTTP Status Codes
- `200`: Success
- `201`: Created
- `400`: Bad Request (validation error)
- `401`: Unauthorized (not logged in)
- `403`: Forbidden (not admin)
- `404`: Not Found
- `500`: Internal Server Error

---

## 🔧 CORS Configuration
**Allowed Origins**:
- `http://himtalks.japaneast.cloudapp.azure.com`
- `https://himtalks.vercel.app`
- `https://admin-himtalks.vercel.app`

**Allowed Methods**: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
**Allowed Headers**: `Content-Type`, `Authorization`
**Credentials**: Enabled (for cookies)
