# TrackHandler Documentation

This file refer to : backend/internal/handlers/track.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, pq, io, os, path/filepath, crypto/hmac, crypto/sha256, encoding/hex, os/exec, regexp, time, fmt, strconv, strings
```

## Struct
```go
type TrackHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewTrackHandler(db *database.DB) *TrackHandler
```

## Constants
```go
const (
    MaxAudioSize = 100 << 20  // 100MB
    MaxAudioDuration = 600    // 10 minutes in seconds
)
```

## Types/Models
```go
type TrackResponse struct {
    ID, UploaderID int
    Title, Artist, Filename, UploaderName, CreatedAt, UpdatedAt, StreamURL string
    DurationSeconds *int
    Tags []string
    IsPublic bool
}

type CreateTrackRequest struct {
    Title, Artist string `binding:"required"`
    Tags []string
    IsPublic bool
}

type UpdateTrackRequest struct {
    Title, Artist *string
    Tags *[]string
    IsPublic *bool
}

type AudioMetadata struct {
    Duration, Bitrate, SampleRate int
    Format string
    Size int64
}
```

## Methods

### Track Management
```go
func (h *TrackHandler) AddTrackWithUpload(c *gin.Context)
func (h *TrackHandler) ListTracks(c *gin.Context)
func (h *TrackHandler) GetTrack(c *gin.Context)
func (h *TrackHandler) UpdateTrack(c *gin.Context)
func (h *TrackHandler) DeleteTrack(c *gin.Context)
```

### Audio Streaming
```go
func (h *TrackHandler) StreamAudio(c *gin.Context)
func (h *TrackHandler) StreamAudioSigned(c *gin.Context)
func (h *TrackHandler) GenerateStreamURL(c *gin.Context)
```

### AddTrackWithUpload
- **Auth**: common.GetUserIDFromContext required
- **Process**: Multipart upload with audio validation
- **Validation**: File size, format, duration via ffprobe
- **Storage**: audio/ directory with unique filenames
- **Metadata**: Extracts duration, validates format

### Audio Validation
```go
func (h *TrackHandler) validateAudioFile(filePath string) (*AudioMetadata, error)
```
- **Tool**: Uses ffprobe for metadata extraction
- **Formats**: mp3, wav, flac, ogg, m4a, aac
- **Limits**: Duration max 10 minutes, size max 100MB

### Signed URL System
```go
func (h *TrackHandler) generateSignedURL(filename string, userID int, secretKey string) (string, error)
func (h *TrackHandler) validateSignature(filename string, userID int, expiration int64, signature, secretKey string) bool
```
- **Security**: HMAC-SHA256 signatures
- **Expiration**: 1-hour signed URLs
- **Access Control**: Private track protection

### Statistics
```go
func (h *TrackHandler) GetTrackStats(c *gin.Context)
```
- **Data**: Track metadata, file info
- **Public**: Only public tracks accessible

## Audio File Formats
```go
allowedExts := []string{".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac"}
```

## Route Mapping Expectations
- `POST /tracks` → AddTrackWithUpload
- `GET /tracks` → ListTracks
- `GET /tracks/:id` → GetTrack
- `PUT /tracks/:id` → UpdateTrack
- `DELETE /tracks/:id` → DeleteTrack
- `GET /stream/:filename` → StreamAudio
- `GET /stream/signed/:filename` → StreamAudioSigned
- `GET /tracks/stream-url` → GenerateStreamURL
- `GET /tracks/:id/stats` → GetTrackStats

## Middleware Dependencies
- Authentication middleware for upload/update/delete
- Multipart form parsing (handled in methods)

## Database Tables
- tracks (metadata, tags as array, duration)
- users (uploader details)

## File System Structure
```
audio/               # Audio files storage
```

## Validation Pipeline
1. **Size Check**: MaxAudioSize (100MB)
2. **Extension Check**: Allowed audio formats
3. **Format Validation**: ffprobe metadata extraction
4. **Duration Check**: MaxAudioDuration (10 minutes)
5. **Storage**: Unique filename generation

## Streaming Features
- **Public Streaming**: Direct file serving
- **Signed URLs**: Protected access for private tracks
- **Range Requests**: Accept-Ranges header for seeking
- **Access Control**: Public/private track permissions

## Security Features
- **Signed URLs**: HMAC signatures with expiration
- **Path Security**: filepath.Base() prevents traversal
- **Access Validation**: Track ownership/publicity checks
- **Unique Filenames**: Prevent conflicts and guessing

## Error Patterns
- 401: User not authenticated
- 403: Access denied (private tracks)
- 404: Track/file not found
- 413: File too large
- 415: Unsupported audio format
- 422: Invalid duration, ffprobe validation failure
- 500: File I/O, database errors

## ffprobe Integration
- **Purpose**: Audio metadata extraction and validation
- **Output**: JSON format parsing with regex
- **Fallback**: Graceful handling if ffprobe unavailable
- **Validation**: Format compatibility, duration limits