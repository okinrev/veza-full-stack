package shared_resources

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/okinrev/veza-web-app/internal/utils/response"
)

// Dans search/handler.go, tag/handler.go
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UploadSharedResource(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	title := c.PostForm("title")
	resourceType := c.PostForm("type")
	description := c.PostForm("description")

	if title == "" {
		response.ErrorJSON(c.Writer, "Title is required", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		response.ErrorJSON(c.Writer, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	resource := map[string]interface{}{
		"id":          1,
		"title":       title,
		"type":        resourceType,
		"description": description,
		"filename":    fileHeader.Filename,
		"uploader_id": userID,
		"is_public":   true,
		"uploaded_at": "2025-01-01T00:00:00Z",
	}

	response.SuccessJSON(c.Writer, resource, "Resource uploaded successfully")
}

func (h *Handler) ListSharedResources(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resources := []map[string]interface{}{
		{
			"id":       1,
			"title":    "Sample Resource",
			"type":     "sample",
			"filename": "sample.zip",
		},
	}

	meta := &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      len(resources),
		TotalPages: 1,
	}

	response.PaginatedJSON(c.Writer, resources, meta, "Resources retrieved successfully")
}

func (h *Handler) SearchSharedResources(c *gin.Context) {
	_ = c.Query("q")
	// TODO: Implement search
	results := []map[string]interface{}{}
	response.SuccessJSON(c.Writer, results, "Search completed")
}

func (h *Handler) UpdateSharedResource(c *gin.Context) {
	// TODO: Implement update
	response.SuccessJSON(c.Writer, nil, "Resource updated successfully")
}

func (h *Handler) DeleteSharedResource(c *gin.Context) {
	// TODO: Implement delete
	response.SuccessJSON(c.Writer, nil, "Resource deleted successfully")
}

func (h *Handler) ServeSharedFile(c *gin.Context) {
	_ = c.Param("filename")
	response.ErrorJSON(c.Writer, "File serving not implemented yet", http.StatusNotImplemented)
}
