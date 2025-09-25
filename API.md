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
- `documentId` (string) - Document ID
- `file` (file) - Document file (PDF, images)

**Supported File Types:**
- PDF: `application/pdf`
- Images: `image/jpeg`, `image/png`, `image/gif`, `image/bmp`, `image/webp`, `image/tiff`

**Processing Workflow:**
1. File uploaded to MinIO storage
2. Document record created in database (status: "pending")
3. Status updated to "processing"
4. Background processing starts:
   - **PDFs**: Converted to images → OpenAI OCR → Text extraction
   - **Images**: Direct OpenAI OCR → Text extraction
5. Extracted text stored in database
6. Status updated to "processed"

**Output:**
```json
{
  "message": "Document processing started",
  "documentId": "doc-123",
  "filename": "document.pdf"
}
```

**Processing Status:**
- `pending` - Document queued for processing
- `processing` - Currently being processed
- `processed` - Successfully completed
- `failed` - Processing failed

---

### 3. Get Processing Status
**GET** `/api/v1/process/{id}/status`

**Input:** Path parameter `id` (document ID)

**Output:**
```json
{
  "status": "processed"
}
```

---

### 4. List Documents
**GET** `/api/v1/documents`

**Input:** None

**Output:**
```json
{
  "documents": [
    {
      "id": "doc-123",
      "filename": "document.pdf",
      "summary": "Document summary...",
      "metadata": {...},
      "status": "processed",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

---

### 5. Get Documents by IDs
**POST** `/api/v1/documents/batch`

**Input:**
```json
{
  "ids": ["doc-1", "doc-2", "doc-3"]
}
```

**Output:**
```json
{
  "documents": [
    {
      "id": "doc-1",
      "filename": "document1.pdf",
      "fileType": "pdf",
      "filePath": "documents/document1.pdf",
      "content": "Full document content...",
      "summary": "Document summary...",
      "metadata": {...},
      "status": "processed",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
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
# Health check
curl -X GET http://localhost:8080/api/v1/health

# Upload document
curl -X POST http://localhost:8080/api/v1/process \
  -F 'documentId=doc-123' \
  -F 'file=@document.pdf'

# Check status
curl -X GET http://localhost:8080/api/v1/process/doc-123/status

# List all documents
curl -X GET http://localhost:8080/api/v1/documents

# Get documents by IDs
curl -X POST http://localhost:8080/api/v1/documents/batch \
  -H "Content-Type: application/json" \
  -d '{"ids": ["doc-1", "doc-2", "doc-3"]}'

# Delete document
curl -X DELETE http://localhost:8080/api/v1/documents/doc-123
```

## Document Model

```json
{
  "id": "string",
  "filename": "string",
  "fileType": "string",
  "filePath": "string",
  "content": "string",
  "summary": "string",
  "metadata": "object",
  "status": "string",
  "createdAt": "datetime",
  "updatedAt": "datetime"
}
```

## Status Values

- `pending` - Document is queued for processing
- `processing` - Document is currently being processed
- `processed` - Document has been successfully processed
- `failed` - Document processing failed