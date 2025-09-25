// internal/services/processing.go
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"

	"document-embeddings/internal/config"
	"document-embeddings/internal/models"
	"document-embeddings/internal/repository"
	"document-embeddings/pkg/logger"
	minioClient "document-embeddings/pkg/minio"
	"document-embeddings/pkg/openai"
)

type ProcessingService struct {
	repo   *repository.Repository
	minio  *minioClient.Client
	openai *openai.Client
	cfg    *config.Config
	logger *logger.Logger
}

func NewProcessingService(repo *repository.Repository, minio *minioClient.Client, openai *openai.Client, cfg *config.Config, logger *logger.Logger) *ProcessingService {
	return &ProcessingService{
		repo:   repo,
		minio:  minio,
		openai: openai,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *ProcessingService) ProcessDocument(ctx context.Context, documentID string) error {
	// Get document from database
	doc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	if doc.Status == "processing" {
		return fmt.Errorf("document is already being processed")
	}

	// Update status to processing
	if err := s.repo.UpdateDocumentStatus(ctx, documentID, "processing"); err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	// Process document in background
	go func() {
		if err := s.processDocumentAsync(context.Background(), doc); err != nil {
			s.logger.Error("Failed to process document", "documentId", documentID, "error", err)
			s.repo.UpdateDocumentStatus(context.Background(), documentID, "failed")
		}
	}()

	return nil
}

func (s *ProcessingService) ProcessDocumentWithFile(ctx context.Context, documentID string, file *multipart.FileHeader) error {
	// Check if document already exists and is processing
	existingDoc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err == nil && existingDoc.Status == "processing" {
		return fmt.Errorf("document is already being processed")
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file data
	fileData, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("failed to read uploaded file: %w", err)
	}

	// Determine file type from content type
	fileType := s.getFileTypeFromContentType(file.Header.Get("Content-Type"))
	if fileType == "" {
		return fmt.Errorf("unsupported file type")
	}

	// Upload file to MinIO
	filePath := fmt.Sprintf("documents/%s", file.Filename)
	if err := s.uploadFileToMinIO(ctx, filePath, fileData, file.Header.Get("Content-Type")); err != nil {
		return fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	// Create document record
	doc := &models.Document{
		ID:       documentID,
		Filename: file.Filename,
		FileType: fileType,
		FilePath: filePath,
		Status:   "pending",
	}

	// Insert document into database
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return fmt.Errorf("failed to create document record: %w", err)
	}

	// Update status to processing
	if err := s.repo.UpdateDocumentStatus(ctx, documentID, "processing"); err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	// Process document in background
	go func() {
		if err := s.processDocumentAsync(context.Background(), doc); err != nil {
			s.logger.Error("Failed to process document", "documentId", documentID, "error", err)
			s.repo.UpdateDocumentStatus(context.Background(), documentID, "failed")
		}
	}()

	return nil
}

func (s *ProcessingService) processDocumentAsync(ctx context.Context, doc *models.Document) error {
	// Download file from MinIO
	fileData, err := s.downloadFile(ctx, doc.FilePath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Extract text and analyze based on file type
	var extractedText string
	var summary string
	var metadata []byte

	if s.isImageFile(doc.FileType) {
		// For images, use the new analyzeImage method
		analysis, err := s.analyzeImage(ctx, fileData, doc.FileType)
		if err != nil {
			return fmt.Errorf("failed to analyze image: %w", err)
		}

		// Extract text content from analysis
		if textContent, exists := analysis.Metadata["raw_text_content"]; exists {
			extractedText = textContent
		} else {
			extractedText = analysis.Summary
		}

		summary = analysis.Summary

		// Store only the metadata part, not the full analysis
		metadata, err = json.Marshal(analysis.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal analysis metadata: %w", err)
		}
	} else {
		// For other files, use the existing extractText method
		extractedText, err = s.extractText(ctx, fileData, doc.FileType, doc.Filename)
		if err != nil {
			return fmt.Errorf("failed to extract text: %w", err)
		}
		summary = extractedText // Use extracted text as summary for non-image files
	}

	// Update document content
	if err := s.repo.UpdateDocumentContent(ctx, doc.ID, extractedText); err != nil {
		return fmt.Errorf("failed to update document content: %w", err)
	}

	// Update document summary
	if err := s.repo.UpdateDocumentSummary(ctx, doc.ID, summary); err != nil {
		return fmt.Errorf("failed to update document summary: %w", err)
	}

	// Update document metadata if available
	if len(metadata) > 0 {
		if err := s.repo.UpdateDocumentMetadata(ctx, doc.ID, metadata); err != nil {
			return fmt.Errorf("failed to update document metadata: %w", err)
		}
	}

	// Chunk the text
	// chunks := s.chunkText(extractedText, 1000, 200) // 1000 chars with 200 overlap

	// Generate embeddings and store chunks
	// for i, chunkContent := range chunks {
	// 	chunkID := uuid.New().String()

	// 	// Generate embedding
	// 	embeddings, err := s.openai.GenerateEmbeddings(ctx, []string{chunkContent})
	// 	if err != nil {
	// 		return fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err)
	// 	}

	// 	// Create chunk metadata
	// 	metadata := map[string]interface{}{
	// 		"chunk_index": i,
	// 		"token_count": len(strings.Fields(chunkContent)),
	// 		"created_at":  time.Now(),
	// 	}
	// 	metadataJSON, _ := json.Marshal(metadata)

	// 	// Create document chunk
	// 	chunk := &models.DocumentChunk{
	// 		ID:         chunkID,
	// 		DocumentID: doc.ID,
	// 		ChunkIndex: i,
	// 		Content:    chunkContent,
	// 		TokenCount: &[]int{len(strings.Fields(chunkContent))}[0],
	// 		Embedding:  embeddings[0],
	// 		Metadata:   metadataJSON,
	// 	}

	// 	if err := s.repo.CreateDocumentChunk(ctx, chunk); err != nil {
	// 		return fmt.Errorf("failed to create document chunk: %w", err)
	// 	}
	// }

	// Update document status to processed
	if err := s.repo.UpdateDocumentStatus(ctx, doc.ID, "processed"); err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	s.logger.Info("Document processed successfully", "documentId", doc.ID)
	return nil
}

func (s *ProcessingService) downloadFile(ctx context.Context, filePath string) ([]byte, error) {
	reader, err := s.minio.GetObject(ctx, filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (s *ProcessingService) extractText(ctx context.Context, fileData []byte, fileType, filename string) (string, error) {
	switch strings.ToLower(fileType) {
	case "pdf":
		return s.extractTextFromPDF(ctx, fileData, filename)
	case "jpeg", "jpg", "png", "gif", "bmp", "webp", "tiff":
		return s.extractTextFromImage(ctx, fileData, fileType)
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func (s *ProcessingService) extractTextFromPDF(ctx context.Context, pdfData []byte, filename string) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "pdf_*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(pdfData); err != nil {
		return "", err
	}
	tmpFile.Close()

	// Convert PDF to images using ImageMagick
	outputDir := filepath.Dir(tmpFile.Name())
	outputPattern := filepath.Join(outputDir, "page_%03d.png")

	cmd := exec.CommandContext(ctx, "convert", tmpFile.Name(), outputPattern)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to convert PDF to images: %w", err)
	}

	// Find generated images
	images, err := filepath.Glob(filepath.Join(outputDir, "page_*.png"))
	if err != nil {
		return "", err
	}

	// Clean up images after processing
	defer func() {
		for _, img := range images {
			os.Remove(img)
		}
	}()

	var extractedTexts []string

	// Extract text from each image
	for _, imagePath := range images {
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			continue
		}

		text, err := s.openai.ExtractTextFromImage(ctx, imageData, "image/png")
		if err != nil {
			s.logger.Warn("Failed to extract text from PDF page", "error", err)
			continue
		}

		if text != "" {
			extractedTexts = append(extractedTexts, text)
		}
	}

	return strings.Join(extractedTexts, "\n\n"), nil
}

func (s *ProcessingService) extractTextFromImage(ctx context.Context, imageData []byte, fileType string) (string, error) {
	mimeType := fmt.Sprintf("image/%s", strings.ToLower(fileType))
	if fileType == "jpg" {
		mimeType = "image/jpeg"
	}

	return s.openai.ExtractTextFromImage(ctx, imageData, mimeType)
}

func (s *ProcessingService) analyzeImage(ctx context.Context, imageData []byte, fileType string) (*openai.ImageAnalysis, error) {
	mimeType := fmt.Sprintf("image/%s", strings.ToLower(fileType))
	if fileType == "jpg" {
		mimeType = "image/jpeg"
	}

	return s.openai.AnalyzeImage(ctx, imageData, mimeType)
}

func (s *ProcessingService) isImageFile(fileType string) bool {
	imageTypes := []string{"jpeg", "jpg", "png", "gif", "bmp", "webp", "tiff"}
	for _, imgType := range imageTypes {
		if strings.ToLower(fileType) == imgType {
			return true
		}
	}
	return false
}

// func (s *ProcessingService) chunkText(text string, chunkSize, overlap int) []string {
// 	if len(text) <= chunkSize {
// 		return []string{text}
// 	}

// 	var chunks []string
// 	start := 0

// 	for start < len(text) {
// 		end := start + chunkSize
// 		if end > len(text) {
// 			end = len(text)
// 		}

// 		// Find a good breaking point (space or newline)
// 		if end < len(text) {
// 			for i := end; i > start+chunkSize/2 && i < len(text); i-- {
// 				if text[i] == ' ' || text[i] == '\n' || text[i] == '.' {
// 					end = i + 1
// 					break
// 				}
// 			}
// 		}

// 		chunks = append(chunks, strings.TrimSpace(text[start:end]))
// 		start = end - overlap
// 		if start < 0 {
// 			start = 0
// 		}
// 	}

// 	return chunks
// }

func (s *ProcessingService) GetProcessingStatus(ctx context.Context, documentID string) (*models.StatusResponse, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// chunkCount, err := s.repo.GetDocumentChunkCount(ctx, documentID)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get chunk count: %w", err)
	// }

	return &models.StatusResponse{
		Status: doc.Status,
		// ChunkCount: chunkCount,
	}, nil
}

func (s *ProcessingService) getFileTypeFromContentType(contentType string) string {
	contentTypeMap := map[string]string{
		"application/pdf": "pdf",
		"image/jpeg":      "jpg",
		"image/jpg":       "jpg",
		"image/png":       "png",
		"image/gif":       "gif",
		"image/bmp":       "bmp",
		"image/webp":      "webp",
		"image/tiff":      "tiff",
	}

	if fileType, exists := contentTypeMap[contentType]; exists {
		return fileType
	}
	return ""
}

func (s *ProcessingService) uploadFileToMinIO(ctx context.Context, filePath string, fileData []byte, contentType string) error {
	// Create a reader from the file data
	reader := bytes.NewReader(fileData)

	// Upload file to MinIO
	_, err := s.minio.PutObject(ctx, filePath, reader, int64(len(fileData)), minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	s.logger.Info("File uploaded to MinIO", "filePath", filePath, "size", len(fileData))
	return nil
}
