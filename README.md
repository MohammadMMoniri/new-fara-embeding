# Document Embeddings Service

A comprehensive Go service for document processing, text extraction, embedding generation, and vector similarity search using pgvector.

## Features

- Document processing pipeline (PDF, images)
- Text extraction using OpenAI GPT-4o-mini OCR
- Text chunking with configurable overlap
- Embedding generation using OpenAI text-embedding models
- Vector similarity search with pgvector
- RESTful API with comprehensive endpoints
- Graceful shutdown and error handling
- Production-ready with logging and monitoring

## Quick Start

1. **Setup Environment**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

2. **Start Dependencies**
   ```bash
   docker-compose up postgres minio
   ```

3. **Run Migrations**
   ```bash
   # Apply your Prisma migrations or run init.sql
   ```

4. **Build and Run**
   ```bash
   go mod tidy
   go run main.go
   ```

## API Endpoints

- `POST /api/v1/process` - Process document and generate embeddings
- `GET /api/v1/process/{id}/status` - Get processing status
- `POST /api/v1/search` - Semantic search across documents
- `GET /api/v1/documents/{id}/chunks` - Get all chunks for a document
- `DELETE /api/v1/documents/{id}` - Remove document and chunks
- `GET /api/v1/health` - Health check

## Configuration

Environment variables:
- `DATABASE_URL` - PostgreSQL connection string
- `MINIO_*` - MinIO object storage configuration
- `OPENAI_API_KEY` - OpenAI API key for embeddings and OCR
- `LOG_LEVEL` - Logging level (debug, info, warn, error)

## Dependencies

- PostgreSQL with pgvector extension
- MinIO for object storage
- ImageMagick for PDF processing
- OpenAI API for embeddings and OCR

## Production Deployment

Use the included Dockerfile and docker-compose.yml for containerized deployment.
