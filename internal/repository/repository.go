// internal/repository/repository.go
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"document-embeddings/internal/models"
	"document-embeddings/pkg/database"
	"document-embeddings/pkg/logger"
)

type Repository struct {
	db     *database.DB
	logger *logger.Logger
}

func New(db *database.DB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) GetDocumentByID(ctx context.Context, id string) (*models.Document, error) {
	query := `SELECT id, filename, file_type, file_path, content, summary, metadata, status, user_id, created_at, updated_at 
			  FROM "Document" WHERE id = $1`

	var doc models.Document
	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.Filename, &doc.FileType, &doc.FilePath,
		&doc.Content, &doc.Summary, &doc.Metadata, &doc.Status, &doc.UserID,
		&doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("document not found")
		}
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) UpdateDocumentStatus(ctx context.Context, id, status string) error {
	query := `UPDATE "Document" SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *Repository) UpdateDocumentContent(ctx context.Context, id, content string) error {
	query := `UPDATE "Document" SET content = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, content, id)
	return err
}

func (r *Repository) UpdateDocumentSummary(ctx context.Context, id, summary string) error {
	query := `UPDATE "Document" SET summary = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, summary, id)
	return err
}

// func (r *Repository) CreateDocumentChunk(ctx context.Context, chunk *models.DocumentChunk) error {
// 	query := `INSERT INTO "DocumentChunk"
// 			  (id, document_id, chunk_index, content, token_count, embedding, metadata, created_at, updated_at)
// 			  VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`

// 	_, err := r.db.Exec(ctx, query,
// 		chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Content,
// 		chunk.TokenCount, chunk.Embedding, chunk.Metadata,
// 	)
// 	return err
// }

// func (r *Repository) GetDocumentChunks(ctx context.Context, documentID string) ([]models.DocumentChunk, error) {
// 	query := `SELECT id, document_id, chunk_index, content, token_count, embedding, metadata, created_at, updated_at
// 			  FROM "DocumentChunk" WHERE document_id = $1 ORDER BY chunk_index`

// 	rows, err := r.db.Query(ctx, query, documentID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var chunks []models.DocumentChunk
// 	for rows.Next() {
// 		var chunk models.DocumentChunk
// 		err := rows.Scan(
// 			&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content,
// 			&chunk.TokenCount, &chunk.Embedding, &chunk.Metadata,
// 			&chunk.CreatedAt, &chunk.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		chunks = append(chunks, chunk)
// 	}

// 	return chunks, nil
// }

// func (r *Repository) SearchSimilarChunks(ctx context.Context, embedding []float32, limit int, userID string) ([]models.DocumentChunk, error) {
// 	query := `SELECT c.id, c.document_id, c.chunk_index, c.content, c.token_count,
// 				 c.embedding, c.metadata, c.created_at, c.updated_at,
// 				 d.filename, d.file_type, d.user_id,
// 				 1 - (c.embedding <=> $1::vector) as similarity
// 			  FROM "DocumentChunk" c
// 			  JOIN "Document" d ON c.document_id = d.id
// 			  WHERE d.user_id = $2 AND d.status = 'processed'
// 			  ORDER BY c.embedding <=> $1::vector
// 			  LIMIT $3`

// 	rows, err := r.db.Query(ctx, query, embedding, userID, limit)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var chunks []models.DocumentChunk
// 	for rows.Next() {
// 		var chunk models.DocumentChunk
// 		var doc models.Document
// 		var similarity float64

// 		err := rows.Scan(
// 			&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content,
// 			&chunk.TokenCount, &chunk.Embedding, &chunk.Metadata,
// 			&chunk.CreatedAt, &chunk.UpdatedAt,
// 			&doc.Filename, &doc.FileType, &doc.UserID,
// 			&similarity,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		chunk.Document = &doc
// 		chunk.Similarity = &similarity
// 		chunks = append(chunks, chunk)
// 	}

// 	return chunks, nil
// }

func (r *Repository) DeleteDocument(ctx context.Context, id string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete chunks first (foreign key constraint)
	// _, err = tx.Exec(ctx, `DELETE FROM "DocumentChunk" WHERE document_id = $1`, id)
	// if err != nil {
	// 	return err
	// }

	// Delete document
	_, err = tx.Exec(ctx, `DELETE FROM "Document" WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// func (r *Repository) GetDocumentChunkCount(ctx context.Context, documentID string) (int, error) {
// 	var count int
// 	query := `SELECT COUNT(*) FROM "DocumentChunk" WHERE document_id = $1`
// 	err := r.db.QueryRow(ctx, query, documentID).Scan(&count)
// 	return count, err
// }

func (r *Repository) ListDocuments(ctx context.Context, limit, offset string) ([]models.Document, error) {
	query := `SELECT id, filename, summary, metadata, status, created_at, updated_at 
			  FROM "Document" 
			  WHERE status = 'processed' 
			  ORDER BY created_at DESC 
			  LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		err := rows.Scan(
			&doc.ID, &doc.Filename, &doc.Summary, &doc.Metadata,
			&doc.Status, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *Repository) CreateDocument(ctx context.Context, doc *models.Document) error {
	query := `INSERT INTO "Document" 
			  (id, filename, file_type, file_path, content, summary, metadata, status, user_id, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())`

	_, err := r.db.Exec(ctx, query,
		doc.ID, doc.Filename, doc.FileType, doc.FilePath,
		doc.Content, doc.Summary, doc.Metadata, doc.Status, doc.UserID,
	)
	return err
}
