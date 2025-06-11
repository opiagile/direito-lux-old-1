"""
RAG API endpoints for legal consultations
"""

from typing import List, Dict, Any, Optional
from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks
from pydantic import BaseModel, Field

from app.services.rag_service import RAGService
from app.services.evaluation_service import EvaluationService
from main import get_rag_service, get_evaluation_service

router = APIRouter()


class LegalQueryRequest(BaseModel):
    """Request model for legal queries"""
    question: str = Field(..., min_length=10, max_length=2000, description="Legal question")
    query_type: str = Field(default="geral", description="Type of query: processo, legislacao, jurisprudencia, geral")
    context_filters: Optional[Dict[str, Any]] = Field(default=None, description="Filters for document retrieval")
    evaluate_response: bool = Field(default=True, description="Whether to evaluate the response quality")
    ground_truth: Optional[str] = Field(default=None, description="Ground truth answer for evaluation")


class LegalQueryResponse(BaseModel):
    """Response model for legal queries"""
    query_id: str
    timestamp: str
    question: str
    answer: str
    query_type: str
    sources: List[Dict[str, Any]]
    processing_time: float
    retrieved_docs_count: int
    total_docs_found: int
    evaluation: Optional[Dict[str, Any]] = None


class BatchQueryRequest(BaseModel):
    """Request model for batch legal queries"""
    queries: List[LegalQueryRequest] = Field(..., min_items=1, max_items=50)
    
    
class BatchQueryResponse(BaseModel):
    """Response model for batch legal queries"""
    batch_id: str
    timestamp: str
    total_queries: int
    successful_queries: int
    failed_queries: int
    results: List[Dict[str, Any]]
    total_processing_time: float


@router.post("/query", response_model=LegalQueryResponse)
async def legal_query(
    request: LegalQueryRequest,
    background_tasks: BackgroundTasks,
    rag_service: RAGService = Depends(get_rag_service),
    evaluation_service: EvaluationService = Depends(get_evaluation_service)
):
    """
    Process a legal query using RAG
    
    This endpoint accepts legal questions and returns AI-generated answers
    based on the legal knowledge base using Retrieval-Augmented Generation.
    """
    try:
        query_id = f"query_{int(datetime.utcnow().timestamp())}"
        
        # Validate query type
        valid_types = ["processo", "legislacao", "jurisprudencia", "geral"]
        if request.query_type not in valid_types:
            raise HTTPException(
                status_code=400,
                detail=f"Invalid query_type. Must be one of: {valid_types}"
            )
        
        # Process the legal query
        result = await rag_service.process_legal_query(
            question=request.question,
            query_type=request.query_type,
            filters=request.context_filters
        )
        
        # Prepare response
        response = LegalQueryResponse(
            query_id=query_id,
            timestamp=datetime.utcnow().isoformat(),
            question=request.question,
            answer=result["answer"],
            query_type=result["query_type"],
            sources=result["sources"],
            processing_time=result["processing_time"],
            retrieved_docs_count=result["retrieved_docs_count"],
            total_docs_found=result["total_docs_found"]
        )
        
        # Evaluate response in background if requested
        if request.evaluate_response:
            background_tasks.add_task(
                _evaluate_response_background,
                evaluation_service,
                request.question,
                result["answer"],
                [source.get("title", "") for source in result["sources"]],
                request.ground_truth
            )
        
        return response
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error processing legal query: {str(e)}")


@router.post("/batch-query", response_model=BatchQueryResponse)
async def batch_legal_query(
    request: BatchQueryRequest,
    rag_service: RAGService = Depends(get_rag_service)
):
    """
    Process multiple legal queries in batch
    
    This endpoint allows processing multiple legal questions at once
    for better efficiency.
    """
    try:
        batch_id = f"batch_{int(datetime.utcnow().timestamp())}"
        start_time = datetime.utcnow()
        
        results = []
        successful_queries = 0
        failed_queries = 0
        
        for i, query_request in enumerate(request.queries):
            try:
                # Validate query type
                valid_types = ["processo", "legislacao", "jurisprudencia", "geral"]
                if query_request.query_type not in valid_types:
                    failed_queries += 1
                    results.append({
                        "query_index": i,
                        "status": "failed",
                        "error": f"Invalid query_type. Must be one of: {valid_types}"
                    })
                    continue
                
                # Process query
                result = await rag_service.process_legal_query(
                    question=query_request.question,
                    query_type=query_request.query_type,
                    filters=query_request.context_filters
                )
                
                result["query_index"] = i
                result["status"] = "success"
                results.append(result)
                successful_queries += 1
                
            except Exception as e:
                failed_queries += 1
                results.append({
                    "query_index": i,
                    "status": "failed",
                    "error": str(e)
                })
        
        total_processing_time = (datetime.utcnow() - start_time).total_seconds()
        
        return BatchQueryResponse(
            batch_id=batch_id,
            timestamp=datetime.utcnow().isoformat(),
            total_queries=len(request.queries),
            successful_queries=successful_queries,
            failed_queries=failed_queries,
            results=results,
            total_processing_time=total_processing_time
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error processing batch queries: {str(e)}")


@router.get("/stats")
async def get_rag_stats(
    rag_service: RAGService = Depends(get_rag_service)
):
    """
    Get RAG system statistics
    
    Returns information about the RAG system including knowledge base stats,
    model configuration, and system health.
    """
    try:
        stats = await rag_service.get_knowledge_base_stats()
        return stats
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error retrieving RAG stats: {str(e)}")


async def _evaluate_response_background(
    evaluation_service: EvaluationService,
    question: str,
    answer: str,
    contexts: List[str],
    ground_truth: Optional[str] = None
):
    """Background task to evaluate RAG response"""
    try:
        await evaluation_service.evaluate_rag_response(
            question=question,
            answer=answer,
            contexts=contexts,
            ground_truth=ground_truth
        )
    except Exception as e:
        # Log error but don't fail the main request
        print(f"Background evaluation failed: {e}")