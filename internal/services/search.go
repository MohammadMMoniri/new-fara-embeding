// internal/services/search.go
package services

import (
	"context"

	"document-embeddings/internal/repository"
	"document-embeddings/pkg/logger"
	"document-embeddings/pkg/openai"
)

type SearchService struct {
	repo   *repository.Repository
	openai *openai.Client
	logger *logger.Logger
}

func NewSearchService(repo *repository.Repository, openai *openai.Client, logger *logger.Logger) *SearchService {
	return &SearchService{
		repo:   repo,
		openai: openai,
		logger: logger,
	}
}

// func (s *SearchService) SearchSimilarDocuments(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
// 	// Generate embedding for query
// 	embeddings, err := s.openai.GenerateEmbeddings(ctx, []string{req.Query})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
// 	}

// 	// Set default limit
// 	limit := req.Limit
// 	if limit <= 0 || limit > 100 {
// 		limit = 20
// 	}

// 	// Search for similar chunks
// 	chunks, err := s.repo.SearchSimilarChunks(ctx, embeddings[0], limit, req.UserID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to search similar chunks: %w", err)
// 	}

// 	return &models.SearchResponse{
// 		Results: chunks,
// 		Total:   len(chunks),
// 	}, nil
// }

// func (s *SearchService) GetDocumentChunks(ctx context.Context, documentID string) ([]models.DocumentChunk, error) {
// 	return s.repo.GetDocumentChunks(ctx, documentID)
// }

func (s *SearchService) DeleteDocument(ctx context.Context, documentID string) error {
	return s.repo.DeleteDocument(ctx, documentID)
}
