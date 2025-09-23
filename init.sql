-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create Document table
CREATE TABLE IF NOT EXISTS "Document" (
    id VARCHAR(255) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    content TEXT,
    metadata JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create DocumentChunk table
CREATE TABLE IF NOT EXISTS "DocumentChunk" (
    id VARCHAR(255) PRIMARY KEY,
    document_id VARCHAR(255) NOT NULL REFERENCES "Document"(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    token_count INTEGER,
    embedding vector(1536),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_document_user_id ON "Document"(user_id);
CREATE INDEX IF NOT EXISTS idx_document_status ON "Document"(status);
CREATE INDEX IF NOT EXISTS idx_document_chunk_document_id ON "DocumentChunk"(document_id);
CREATE INDEX IF NOT EXISTS idx_document_chunk_embedding ON "DocumentChunk" USING hnsw (embedding vector_cosine_ops);

-- Alternative index (choose one based on your use case)
-- CREATE INDEX IF NOT EXISTS idx_document_chunk_embedding_ivfflat ON "DocumentChunk" USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
