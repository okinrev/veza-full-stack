# SharedResourcesHandler Documentation

This file refer to : backend/internal/handlers/shared_resources.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, pq, io, os, path/filepath, mime, fmt, strconv, strings, time
```

## Struct
```go
type SharedResourcesHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewSharedResourcesHandler(db *database.DB) *SharedResourcesHandler
```

## Constants
```go
const (
    MaxResourceSize = 50 << 20  // 50MB
)
```

## Types/Models
```go
type SharedResourceResponse struct {
    ID, UploaderID, DownloadCount int
    Title, Filename, URL, Type, UploaderUsername string
    Description *string
    Tags []string
    IsPublic bool
    UploadedAt, UpdatedAt, DownloadURL string
}

type CreateSharedResourceRequest struct {
    Title, Type string `binding:"required"`
    Description *string
    Tags []string
    IsPublic bool
}

type UpdateSharedResourceRequest struct {
    Title, Description, Type *string
    Tags *[]string
    IsPublic *bool
}

type ResourceType struct {
    Name, Description string
    Extensions []string
    MaxSize int64
}

type DetailedResourceStats struct {
    TotalResources, TotalDownloads int
    TotalSize int64
    ResourcesByType map[string]int
    PopularResources []PopularResource
    RecentUploads []RecentUpload
    TopUploaders []TopUploader
}
```

## Resource Types
```go
resourceTypes := map[string]ResourceType{
    "sample":   {Extensions: {".wav", ".mp3", ".aiff", ".flac"}, MaxSize: 50MB},
    "preset":   {Extensions: {".fxp", ".vstpreset", ".h2p", ".adg"}, MaxSize: 5MB},
    "plugin":   {Extensions: {".vst", ".vst3", ".dll", ".component"}, MaxSize: 50MB},
    "template": {Extensions: {".als", ".logic", ".ptx", ".flp"}, MaxSize: 50MB},
    "midi":     {Extensions: {".mid", ".midi"}, MaxSize: 1MB},
    "document": {Extensions: {".pdf", ".doc", ".docx", ".txt"}, MaxSize: 20MB},
}
```

## Methods

### Resource Management
```go
func (h *SharedResourcesHandler) UploadSharedResource(c *gin.Context)
func (h *SharedResourcesHandler) ListSharedResources(c *gin.Context)
func (h *SharedResourcesHandler) SearchSharedResources(c *gin.Context)
func (h *SharedResourcesHandler) UpdateSharedResource(c *gin.Context)
func (h *SharedResourcesHandler) DeleteSharedResource(c *gin.Context)
func (h *SharedResourcesHandler) ServeSharedFile(c *gin.Context)
```

### UploadSharedResource
- **Auth**: common.GetUserIDFromContext required
- **Process**: Multipart upload with type validation
- **Validation**: File type/size based on resource type
- **Storage**: shared_resources/ directory with unique filenames

### ListSharedResources
- **Auth**: Optional (show_private requires auth)
- **Params**: page, limit, show_private
- **Process**: Public resources or user's own resources

### SearchSharedResources
- **Auth**: None (public only)
- **Params**: q, type, tag, uploader, limit
- **Process**: Multi-field search on public resources

### ServeSharedFile
- **Auth**: None for public, required for private
- **Process**: File serving with download tracking
- **Features**: MIME type detection, download counting

### Statistics
```go
func (h *SharedResourcesHandler) GetPredefinedResourceTypes(c *gin.Context)
func (h *SharedResourcesHandler) GetDetailedStats(c *gin.Context)
func (h *SharedResourcesHandler) GetDownloadStats(c *gin.Context)
```

## Validation Methods
```go
func (h *SharedResourcesHandler) validateResourceType(filename, resourceType string, fileSize int64) error
```

## Route Mapping Expectations
- `POST /shared-resources` → UploadSharedResource
- `GET /shared-resources` → ListSharedResources
- `GET /shared-resources/search` → SearchSharedResources
- `PUT /shared-resources/:id` → UpdateSharedResource
- `DELETE /shared-resources/:id` → DeleteSharedResource
- `GET /shared-resources/:filename` → ServeSharedFile
- `GET /shared-resources/types` → GetPredefinedResourceTypes
- `GET /shared-resources/stats` → GetDetailedStats
- `GET /shared-resources/:id/stats` → GetDownloadStats

## Middleware Dependencies
- Authentication middleware for upload/update/delete
- Multipart form parsing (handled in methods)

## Database Tables
- shared_resources (metadata, tags as array)
- users (uploader details)

## File System Structure
```
shared_resources/     # All shared files
```

## Features
- Type-specific validation (extensions, size limits)
- Public/private access control
- Download tracking and statistics
- Tag-based organization (PostgreSQL arrays)
- Search across multiple fields
- MIME type detection for proper serving

## Security Features
- File type validation per resource type
- Size limits per type
- Access control for private resources
- Path traversal prevention

## Error Patterns
- 401: User not authenticated (private operations)
- 403: Access denied (private resources)
- 404: Resource not found
- 413: File too large
- 415: Unsupported file type
- 500: File I/O, database errors