// internal/models/models.go
package models

import (
	"time"
)

type Document struct {
	ID        string                 `json:"id" db:"id"`
	Filename  string                 `json:"filename" db:"filename"`
	FileType  string                 `json:"fileType" db:"file_type"`
	FilePath  string                 `json:"filePath" db:"file_path"`
	Content   *string                `json:"content" db:"content"`
	Summary   *string                `json:"summary" db:"summary"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
	Status    string                 `json:"status" db:"status"`
	CreatedAt time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time              `json:"updatedAt" db:"updated_at"`
}

// type DocumentChunk struct {
// 	ID         string    `json:"id" db:"id"`
// 	DocumentID string    `json:"documentId" db:"document_id"`
// 	ChunkIndex int       `json:"chunkIndex" db:"chunk_index"`
// 	Content    string    `json:"content" db:"content"`
// 	TokenCount *int      `json:"tokenCount" db:"token_count"`
// 	Embedding  []float32 `json:"embedding" db:"embedding"`
// 	Metadata   []byte    `json:"metadata" db:"metadata"`
// 	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
// 	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
// 	Document   *Document `json:"document,omitempty"`
// 	Similarity *float64  `json:"similarity,omitempty"`
// }

type ProcessRequest struct {
	ID string `json:"id" binding:"required"`
}

type SearchRequest struct {
	Query   string                 `json:"query" binding:"required"`
	Limit   int                    `json:"limit"`
	UserID  string                 `json:"userId" binding:"required"`
	Filters map[string]interface{} `json:"filters"`
}

// type SearchResponse struct {
// 	Results []DocumentChunk `json:"results"`
// 	Total   int             `json:"total"`
// }

type StatusResponse struct {
	Status string `json:"status"`
	// ChunkCount int    `json:"chunkCount"`
}
