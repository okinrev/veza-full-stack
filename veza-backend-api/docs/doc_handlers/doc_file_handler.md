# FileHandler Documentation

This file refer to : backend/internal/handlers/file.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, io, os, path/filepath, strconv, strings, fmt, time
```

## Struct
```go
type FileHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewFileHandler(db *database.DB) *FileHandler
```

## Constants
```go
const (
    MaxFileSize = 10 << 20        // 10MB
    MaxInternalDocSize = 50 << 20 // 50MB
)
```

## Types/Models
```go
type FileResponse struct {
    ID, ProductID int
    Filename, URL, Type, MimeType string
    Size int64
    UploadedAt, UpdatedAt string
}

type InternalDocResponse struct {
    FileResponse + Title string
}
```

## Methods

### File Management
```go
func (h *FileHandler) UploadFile(c *gin.Context)
func (h *FileHandler) ListProductFiles(c *gin.Context)
func (h *FileHandler) DownloadFile(c *gin.Context)
func (h *FileHandler) DeleteFile(c *gin.Context)
```
- **Auth**: common.GetUserIDFromContext required
- **Validation**: Product ownership, file size/type validation
- **Process**: Multipart upload, secure file storage in uploads/
- **Database**: files table linked to user_products

### Internal Documents
```go
func (h *FileHandler) UploadInternalDoc(c *gin.Context)
func (h *FileHandler) ListInternalDocs(c *gin.Context)
func (h *FileHandler) ServeInternalDoc(c *gin.Context)
```
- **Auth**: Required, product ownership validation
- **Process**: Document upload with titles, storage in internal_docs/
- **Database**: internal_documents table

### Validation Helpers
```go
func (h *FileHandler) validateFileSize(size int64, fileType string) error
func (h *FileHandler) validateFileType(filename, expectedType string) error
```
- **Logic**: Type-specific size limits, extension validation
- **Types**: manual, warranty, invoice, image, document

### Statistics
```go
func (h *FileHandler) GetFileStats(c *gin.Context)
```
- **Auth**: Required
- **Response**: User's file statistics (count, size, types)

## File Type Mappings
```go
allowedExtensions := map[string][]string{
    "manual":   {".pdf", ".doc", ".docx", ".txt"},
    "warranty": {".pdf", ".jpg", ".jpeg", ".png"},
    "invoice":  {".pdf", ".jpg", ".jpeg", ".png"},
    "image":    {".jpg", ".jpeg", ".png", ".gif", ".webp"},
    "document": {".pdf", ".doc", ".docx", ".txt", ".rtf"},
}
```

## Route Mapping Expectations
- `POST /products/:id/files` → UploadFile
- `GET /products/:id/files` → ListProductFiles
- `GET /files/:id/download` → DownloadFile
- `DELETE /files/:id` → DeleteFile
- `POST /products/:id/docs` → UploadInternalDoc
- `GET /products/:id/docs` → ListInternalDocs
- `GET /docs/:id` → ServeInternalDoc
- `GET /files/stats` → GetFileStats

## Middleware Dependencies
- Authentication middleware
- Multipart form parsing (handled in methods)

## Database Tables
- user_products (ownership validation)
- files (file metadata)
- internal_documents (document metadata)

## File System Structure
```
uploads/          # User uploaded files
internal_docs/    # Internal documentation
```

## Security Features
- Path traversal prevention (filepath.Base)
- Ownership validation
- File type/size restrictions
- MIME type detection

## Error Patterns
- 401: User not authenticated
- 403: Not authorized (ownership)
- 404: Product/file not found
- 413: File too large
- 415: Unsupported file type
- 500: File I/O, database errors

## Helper Methods
```go
func (h *FileHandler) getFileByID(fileID, userID int) (*FileResponse, error)
func (h *FileHandler) getInternalDocByID(docID, userID int) (*InternalDocResponse, error)
```