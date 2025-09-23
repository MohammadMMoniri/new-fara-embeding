# Document Embeddings API

## Routes

### 1. Health Check
**GET** `/api/v1/health`

**Input:** None

**Output:**
```json
{
  "status": "ok",
  "service": "document-embeddings"
}
```

---

### 2. Process Document
**POST** `/api/v1/process`

**Input:** `multipart/form-data`
- `documentId` (string) - Document ID from external service
- `userId` (string) - User identifier
- `file` (file) - Document file (PDF, images)

**Output:**
```json
{
  "message": "Document processing started",
  "documentId": "doc-123",
  "filename": "document.pdf"
}
```

---

### 3. Get Processing Status
**GET** `/api/v1/process/{id}/status`

**Input:** Path parameter `id` (document ID)

**Output:**
```json
{
  "status": "processed",
  "chunkCount": 15
}
```

---

### 4. Search Documents
**POST** `/api/v1/search`

**Input:**
```json
{
  "query": "search query",
  "userId": "user-456",
  "limit": 10,
  "filters": {}
}
```

**Output:**
```json
{
  "results": [
    {
      "id": "chunk-123",
      "documentId": "doc-123",
      "content": "chunk content...",
      "similarity": 0.95,
      "document": {
        "filename": "document.pdf",
        "fileType": "pdf"
      }
    }
  ],
  "total": 1
}
```

---

### 5. Get Document Chunks
**GET** `/api/v1/documents/{id}/chunks`

**Input:** Path parameter `id` (document ID)

**Output:**
```json
{
  "chunks": [
    {
      "id": "chunk-123",
      "content": "chunk content...",
      "chunkIndex": 0,
      "tokenCount": 150
    }
  ],
  "total": 15
}
```

---

### 6. Delete Document
**DELETE** `/api/v1/documents/{id}`

**Input:** Path parameter `id` (document ID)

**Output:**
```json
{
  "message": "Document deleted successfully",
  "documentId": "doc-123"
}
```

## Example Usage

```bash
# Upload document
curl -X POST http://localhost:8080/api/v1/process \
  -F 'documentId=doc-123' \
  -F 'userId=user-456' \
  -F 'file=@document.pdf'

# Check status
curl -X GET http://localhost:8080/api/v1/process/doc-123/status

# Search
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"query": "machine learning", "userId": "user-456"}'
```
