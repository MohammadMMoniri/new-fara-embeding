// internal/api/api.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"document-embeddings/internal/services"
	"document-embeddings/pkg/logger"
)

type Handler struct {
	services *services.Services
	logger   *logger.Logger
}

func New(services *services.Services, logger *logger.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func RegisterRoutes(r *gin.Engine, h *Handler) {
	api := r.Group("/api/v1")
	{
		api.GET("/health", h.HealthCheck)
		api.POST("/process", h.ProcessDocument)
		api.GET("/process/:id/status", h.GetProcessingStatus)
		// api.POST("/search", h.SearchDocuments)
		// api.GET("/documents/:id/chunks", h.GetDocumentChunks)
		api.GET("/documents", h.ListDocuments)
		api.DELETE("/documents/:id", h.DeleteDocument)
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "document-embeddings",
	})
}

func (h *Handler) ProcessDocument(c *gin.Context) {
	// Get form data
	documentID := c.PostForm("documentId")
	userID := c.PostForm("userId")

	if documentID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "documentId and userId are required"})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/gif":       true,
		"image/bmp":       true,
		"image/webp":      true,
		"image/tiff":      true,
	}

	if !allowedTypes[file.Header.Get("Content-Type")] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type"})
		return
	}

	// Process the document with file upload
	if err := h.services.Processing.ProcessDocumentWithFile(c.Request.Context(), documentID, userID, file); err != nil {
		h.logger.Error("Failed to process document", "documentId", documentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start document processing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Document processing started",
		"documentId": documentID,
		"filename":   file.Filename,
	})
}

func (h *Handler) GetProcessingStatus(c *gin.Context) {
	documentID := c.Param("id")

	status, err := h.services.Processing.GetProcessingStatus(c.Request.Context(), documentID)
	if err != nil {
		h.logger.Error("Failed to get processing status", "documentId", documentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get processing status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// func (h *Handler) SearchDocuments(c *gin.Context) {
// 	var req struct {
// 		Query   string                 `json:"query" binding:"required"`
// 		Limit   int                    `json:"limit"`
// 		UserID  string                 `json:"userId" binding:"required"`
// 		Filters map[string]interface{} `json:"filters"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	searchReq := &models.SearchRequest{
// 		Query:   req.Query,
// 		Limit:   req.Limit,
// 		UserID:  req.UserID,
// 		Filters: req.Filters,
// 	}

// 	results, err := h.services.Search.SearchSimilarDocuments(c.Request.Context(), searchReq)
// 	if err != nil {
// 		h.logger.Error("Failed to search documents", "error", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search documents"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, results)
// }

// func (h *Handler) GetDocumentChunks(c *gin.Context) {
// 	documentID := c.Param("id")

// 	chunks, err := h.services.Search.GetDocumentChunks(c.Request.Context(), documentID)
// 	if err != nil {
// 		h.logger.Error("Failed to get document chunks", "documentId", documentID, "error", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get document chunks"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"chunks": chunks,
// 		"total":  len(chunks),
// 	})
// }

func (h *Handler) ListDocuments(c *gin.Context) {

	documents, err := h.services.Search.ListDocuments(c.Request.Context(), "0", "500")
	if err != nil {
		h.logger.Error("Failed to list documents", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list documents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"total":     len(documents),
	})
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	documentID := c.Param("id")

	if err := h.services.Search.DeleteDocument(c.Request.Context(), documentID); err != nil {
		h.logger.Error("Failed to delete document", "documentId", documentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Document deleted successfully",
		"documentId": documentID,
	})
}
