// internal/services/services.go
package services

import (
	"document-embeddings/internal/config"
	"document-embeddings/internal/repository"
	"document-embeddings/pkg/logger"
	minioClient "document-embeddings/pkg/minio"
	"document-embeddings/pkg/openai"
)

type Services struct {
	Processing *ProcessingService
	Search     *SearchService
}

func New(repo *repository.Repository, minio *minioClient.Client, openai *openai.Client, cfg *config.Config, logger *logger.Logger) *Services {
	return &Services{
		Processing: NewProcessingService(repo, minio, openai, cfg, logger),
		Search:     NewSearchService(repo, openai, logger),
	}
}
