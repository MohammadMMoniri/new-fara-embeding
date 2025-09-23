#!/bin/bash

# Test script for the document embeddings API

echo "Testing Document Embeddings API"
echo "================================"

# Health check
echo "1. Health Check:"
curl -X GET http://localhost:8080/api/v1/health
echo -e "\n\n"

# Process document (example with a sample file)
echo "2. Process Document:"
echo "Note: This requires a file upload. Example curl command:"
echo "curl -X POST http://localhost:8080/api/v1/process \\"
echo "  -F 'documentId=test-doc-123' \\"
echo "  -F 'userId=user-456' \\"
echo "  -F 'file=@/path/to/your/document.pdf'"
echo -e "\n"

# Search documents
echo "3. Search Documents:"
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "example search query",
    "userId": "user-456",
    "limit": 10
  }'
echo -e "\n\n"

echo "API Test Complete!"
