"""
Evaluation API endpoints for RAG quality assessment
"""

from typing import List, Dict, Any, Optional
from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel, Field

from app.services.evaluation_service import EvaluationService
from main import get_evaluation_service

router = APIRouter()


class EvaluationRequest(BaseModel):
    """Request model for single evaluation"""
    question: str = Field(..., min_length=10, max_length=2000)
    answer: str = Field(..., min_length=10)
    contexts: List[str] = Field(..., min_items=1)
    ground_truth: Optional[str] = Field(default=None)


class BatchEvaluationRequest(BaseModel):
    """Request model for batch evaluation"""
    evaluations: List[EvaluationRequest] = Field(..., min_items=1, max_items=100)


class EvaluationResponse(BaseModel):
    """Response model for evaluation results"""
    evaluation_id: str
    timestamp: str
    scores: Dict[str, float]
    overall_score: float
    processing_time: float
    has_ground_truth: bool


class BatchEvaluationResponse(BaseModel):
    """Response model for batch evaluation results"""
    batch_id: str
    timestamp: str
    batch_size: int
    metrics: Dict[str, Any]
    overall_stats: Dict[str, Any]
    processing_time: float
    has_ground_truth: bool


class EvaluationSummaryResponse(BaseModel):
    """Response model for evaluation summary"""
    period_days: int
    total_evaluations: int
    overall_mean_score: float
    metric_statistics: Dict[str, Dict[str, float]]
    evaluation_frequency: float
    last_evaluation: Optional[str]


@router.post("/evaluate", response_model=EvaluationResponse)
async def evaluate_response(
    request: EvaluationRequest,
    evaluation_service: EvaluationService = Depends(get_evaluation_service)
):
    """
    Evaluate a single RAG response using Ragas metrics
    
    This endpoint evaluates the quality of a RAG response using various metrics
    like faithfulness, answer relevancy, context precision, etc.
    """
    try:
        result = await evaluation_service.evaluate_rag_response(
            question=request.question,
            answer=request.answer,
            contexts=request.contexts,
            ground_truth=request.ground_truth
        )
        
        if "error" in result:
            raise HTTPException(status_code=500, detail=result["error"])
        
        if "status" in result and result["status"] == "evaluation_disabled":
            raise HTTPException(status_code=503, detail="Evaluation service is disabled")
        
        return EvaluationResponse(
            evaluation_id=result["evaluation_id"],
            timestamp=result["timestamp"],
            scores=result["scores"],
            overall_score=result["overall_score"],
            processing_time=result["processing_time"],
            has_ground_truth=result["has_ground_truth"]
        )
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error during evaluation: {str(e)}")


@router.post("/batch-evaluate", response_model=BatchEvaluationResponse)
async def batch_evaluate_responses(
    request: BatchEvaluationRequest,
    evaluation_service: EvaluationService = Depends(get_evaluation_service)
):
    """
    Evaluate multiple RAG responses in batch using Ragas metrics
    
    This endpoint allows batch evaluation of multiple RAG responses
    for better efficiency and statistical analysis.
    """
    try:
        # Prepare evaluation data
        evaluation_data = []
        for eval_request in request.evaluations:
            eval_item = {
                "question": eval_request.question,
                "answer": eval_request.answer,
                "contexts": eval_request.contexts
            }
            if eval_request.ground_truth:
                eval_item["ground_truth"] = eval_request.ground_truth
            evaluation_data.append(eval_item)
        
        result = await evaluation_service.batch_evaluate(evaluation_data)
        
        if "error" in result:
            raise HTTPException(status_code=500, detail=result["error"])
        
        if "status" in result and result["status"] == "evaluation_disabled":
            raise HTTPException(status_code=503, detail="Evaluation service is disabled")
        
        return BatchEvaluationResponse(
            batch_id=result["batch_id"],
            timestamp=result["timestamp"],
            batch_size=result["batch_size"],
            metrics=result["metrics"],
            overall_stats=result["overall_stats"],
            processing_time=result["processing_time"],
            has_ground_truth=result["has_ground_truth"]
        )
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error during batch evaluation: {str(e)}")


@router.get("/summary", response_model=EvaluationSummaryResponse)
async def get_evaluation_summary(
    days: int = Field(default=7, ge=1, le=365, description="Number of days to include in summary"),
    evaluation_service: EvaluationService = Depends(get_evaluation_service)
):
    """
    Get evaluation summary for the specified period
    
    Returns statistics and trends from evaluations performed
    in the last N days.
    """
    try:
        summary = await evaluation_service.get_evaluation_summary(days=days)
        
        if "error" in summary:
            raise HTTPException(status_code=500, detail=summary["error"])
        
        # Handle case where no evaluations are available
        if "message" in summary:
            return EvaluationSummaryResponse(
                period_days=days,
                total_evaluations=summary.get("total_evaluations", 0),
                overall_mean_score=0.0,
                metric_statistics={},
                evaluation_frequency=0.0,
                last_evaluation=None
            )
        
        return EvaluationSummaryResponse(
            period_days=summary["period_days"],
            total_evaluations=summary["total_evaluations"],
            overall_mean_score=summary["overall_mean_score"],
            metric_statistics=summary["metric_statistics"],
            evaluation_frequency=summary["evaluation_frequency"],
            last_evaluation=summary["last_evaluation"]
        )
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error retrieving evaluation summary: {str(e)}")


@router.get("/metrics")
async def get_available_metrics():
    """
    Get information about available evaluation metrics
    
    Returns details about the Ragas metrics used for evaluation.
    """
    return {
        "metrics": [
            {
                "name": "faithfulness",
                "description": "Measures how faithful the answer is to the given context",
                "range": [0, 1],
                "higher_is_better": True
            },
            {
                "name": "answer_relevancy", 
                "description": "Measures how relevant the answer is to the given question",
                "range": [0, 1],
                "higher_is_better": True
            },
            {
                "name": "context_precision",
                "description": "Measures how precise the retrieved context is",
                "range": [0, 1], 
                "higher_is_better": True
            },
            {
                "name": "context_recall",
                "description": "Measures how well the context covers the information needed to answer the question",
                "range": [0, 1],
                "higher_is_better": True
            },
            {
                "name": "answer_similarity",
                "description": "Measures semantic similarity between generated and ground truth answers",
                "range": [0, 1],
                "higher_is_better": True,
                "requires_ground_truth": True
            },
            {
                "name": "answer_correctness",
                "description": "Measures factual correctness of the answer against ground truth", 
                "range": [0, 1],
                "higher_is_better": True,
                "requires_ground_truth": True
            }
        ],
        "note": "Some metrics require ground truth answers to be provided"
    }


@router.get("/health")
async def evaluation_health_check(
    evaluation_service: EvaluationService = Depends(get_evaluation_service)
):
    """
    Health check for evaluation service
    
    Returns the current status of the evaluation service.
    """
    return {
        "service": "Evaluation Service",
        "status": "healthy" if evaluation_service.is_healthy() else "unhealthy",
        "initialized": evaluation_service._is_initialized,
        "evaluation_enabled": evaluation_service.settings.evaluation_enabled,
        "timestamp": datetime.utcnow().isoformat()
    }