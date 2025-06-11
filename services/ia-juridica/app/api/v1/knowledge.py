"""
Knowledge Base API endpoints for managing legal documents
"""

from typing import List, Dict, Any, Optional
from datetime import datetime
import uuid

from fastapi import APIRouter, Depends, HTTPException, UploadFile, File, Form
from pydantic import BaseModel, Field

from app.services.rag_service import RAGService
from app.services.vector_service import VectorService
from main import get_rag_service, get_vector_service

router = APIRouter()


class DocumentRequest(BaseModel):
    """Request model for adding documents"""
    content: str = Field(..., min_length=50, description="Document content")
    metadata: Dict[str, Any] = Field(..., description="Document metadata")
    

class DocumentResponse(BaseModel):
    """Response model for document operations"""
    document_ids: List[str]
    chunks_created: int
    timestamp: str
    message: str


class SearchRequest(BaseModel):
    """Request model for document search"""
    query: str = Field(..., min_length=3, max_length=500)
    top_k: int = Field(default=5, ge=1, le=20)
    filters: Optional[Dict[str, Any]] = Field(default=None)
    min_similarity: float = Field(default=0.0, ge=0.0, le=1.0)


class SearchResponse(BaseModel):
    """Response model for document search"""
    query: str
    results: List[Dict[str, Any]]
    total_found: int
    processing_time: float
    timestamp: str


class KnowledgeBaseStats(BaseModel):
    """Response model for knowledge base statistics"""
    total_documents: int
    collection_name: str
    metadata_fields: List[str]
    status: str
    last_updated: str


@router.post("/documents", response_model=DocumentResponse)
async def add_legal_document(
    request: DocumentRequest,
    rag_service: RAGService = Depends(get_rag_service)
):
    """
    Add a legal document to the knowledge base
    
    This endpoint accepts legal documents and adds them to the vector database
    after splitting them into appropriate chunks for RAG.
    """
    try:
        # Validate required metadata fields
        required_fields = ["title", "source_type"]
        for field in required_fields:
            if field not in request.metadata:
                raise HTTPException(
                    status_code=400,
                    detail=f"Missing required metadata field: {field}"
                )
        
        # Add document to knowledge base
        document_ids = await rag_service.add_legal_document(
            content=request.content,
            metadata=request.metadata
        )
        
        return DocumentResponse(
            document_ids=document_ids,
            chunks_created=len(document_ids),
            timestamp=datetime.utcnow().isoformat(),
            message=f"Successfully added document with {len(document_ids)} chunks"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error adding document: {str(e)}")


@router.post("/documents/upload", response_model=DocumentResponse)
async def upload_legal_document(
    file: UploadFile = File(...),
    title: str = Form(...),
    source_type: str = Form(...),
    law_number: Optional[str] = Form(default=None),
    article_number: Optional[str] = Form(default=None),
    court: Optional[str] = Form(default=None),
    date: Optional[str] = Form(default=None),
    rag_service: RAGService = Depends(get_rag_service)
):
    """
    Upload and process a legal document file
    
    Accepts various file formats (PDF, DOCX, TXT) and extracts text
    for addition to the knowledge base.
    """
    try:
        # Validate file type
        allowed_types = ["text/plain", "application/pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"]
        if file.content_type not in allowed_types:
            raise HTTPException(
                status_code=400,
                detail=f"Unsupported file type: {file.content_type}. Allowed types: {allowed_types}"
            )
        
        # Read file content
        content_bytes = await file.read()
        
        # Extract text based on file type
        if file.content_type == "text/plain":
            content = content_bytes.decode("utf-8")
        elif file.content_type == "application/pdf":
            # TODO: Implement PDF extraction
            raise HTTPException(status_code=501, detail="PDF extraction not yet implemented")
        elif file.content_type == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
            # TODO: Implement DOCX extraction  
            raise HTTPException(status_code=501, detail="DOCX extraction not yet implemented")
        else:
            raise HTTPException(status_code=400, detail="Unsupported file type")
        
        # Validate content length
        if len(content) < 50:
            raise HTTPException(status_code=400, detail="Document content too short (minimum 50 characters)")
        
        # Prepare metadata
        metadata = {
            "title": title,
            "source_type": source_type,
            "filename": file.filename,
            "file_type": file.content_type,
            "uploaded_at": datetime.utcnow().isoformat()
        }
        
        # Add optional metadata
        if law_number:
            metadata["law_number"] = law_number
        if article_number:
            metadata["article_number"] = article_number
        if court:
            metadata["court"] = court
        if date:
            metadata["date"] = date
        
        # Add document to knowledge base
        document_ids = await rag_service.add_legal_document(
            content=content,
            metadata=metadata
        )
        
        return DocumentResponse(
            document_ids=document_ids,
            chunks_created=len(document_ids),
            timestamp=datetime.utcnow().isoformat(),
            message=f"Successfully uploaded and processed {file.filename} with {len(document_ids)} chunks"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error uploading document: {str(e)}")


@router.post("/search", response_model=SearchResponse)
async def search_documents(
    request: SearchRequest,
    vector_service: VectorService = Depends(get_vector_service)
):
    """
    Search for documents in the knowledge base
    
    Performs semantic search to find relevant legal documents
    based on the query.
    """
    try:
        start_time = datetime.utcnow()
        
        # Search for similar documents
        results = await vector_service.search_similar(
            query=request.query,
            top_k=request.top_k,
            filters=request.filters
        )
        
        # Filter by minimum similarity if specified
        if request.min_similarity > 0:
            results = [
                result for result in results 
                if result["similarity_score"] >= request.min_similarity
            ]
        
        processing_time = (datetime.utcnow() - start_time).total_seconds()
        
        return SearchResponse(
            query=request.query,
            results=results,
            total_found=len(results),
            processing_time=processing_time,
            timestamp=datetime.utcnow().isoformat()
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error searching documents: {str(e)}")


@router.get("/document/{document_id}")
async def get_document(
    document_id: str,
    vector_service: VectorService = Depends(get_vector_service)
):
    """
    Get a specific document by ID
    
    Retrieves a document from the knowledge base using its unique identifier.
    """
    try:
        document = await vector_service.get_document_by_id(document_id)
        
        if not document:
            raise HTTPException(status_code=404, detail="Document not found")
        
        return document
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error retrieving document: {str(e)}")


@router.delete("/document/{document_id}")
async def delete_document(
    document_id: str,
    vector_service: VectorService = Depends(get_vector_service)
):
    """
    Delete a document from the knowledge base
    
    Removes a document and all its chunks from the vector database.
    """
    try:
        # Check if document exists
        document = await vector_service.get_document_by_id(document_id)
        if not document:
            raise HTTPException(status_code=404, detail="Document not found")
        
        # Delete document
        await vector_service.delete_documents([document_id])
        
        return {
            "message": f"Document {document_id} deleted successfully",
            "timestamp": datetime.utcnow().isoformat()
        }
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error deleting document: {str(e)}")


@router.get("/stats", response_model=KnowledgeBaseStats)
async def get_knowledge_base_stats(
    vector_service: VectorService = Depends(get_vector_service)
):
    """
    Get knowledge base statistics
    
    Returns information about the current state of the knowledge base
    including document counts and metadata fields.
    """
    try:
        stats = await vector_service.get_collection_stats()
        
        return KnowledgeBaseStats(
            total_documents=stats["total_documents"],
            collection_name=stats["collection_name"],
            metadata_fields=stats["metadata_fields"],
            status=stats["status"],
            last_updated=datetime.utcnow().isoformat()
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error retrieving stats: {str(e)}")


@router.get("/health")
async def knowledge_base_health(
    vector_service: VectorService = Depends(get_vector_service),
    rag_service: RAGService = Depends(get_rag_service)
):
    """
    Health check for knowledge base services
    
    Returns the health status of vector database and RAG services.
    """
    return {
        "service": "Knowledge Base",
        "vector_service": {
            "status": "healthy" if vector_service.is_healthy() else "unhealthy",
            "initialized": vector_service._is_initialized
        },
        "rag_service": {
            "status": "healthy" if rag_service.is_healthy() else "unhealthy", 
            "initialized": rag_service._is_initialized
        },
        "timestamp": datetime.utcnow().isoformat()
    }