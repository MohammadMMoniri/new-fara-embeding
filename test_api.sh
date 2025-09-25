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
echo "  -F 'file=@/path/to/your/document.pdf'"
echo -e "\n"

# List documents
echo "3. List Documents:"
curl -X GET "http://localhost:8080/api/v1/documents"
echo -e "\n\n"

# Get documents by IDs
echo "4. Get Documents by IDs:"
curl -X POST http://localhost:8080/api/v1/documents/batch \
  -H "Content-Type: application/json" \
  -d '{
    "ids": ["doc-1", "doc-2", "doc-3"]
  }'
echo -e "\n\n"

# Search documents (commented out since search is disabled)
echo "5. Search Documents (disabled):"
echo "Note: Search functionality is currently commented out"
echo -e "\n"

echo "API Test Complete!"
