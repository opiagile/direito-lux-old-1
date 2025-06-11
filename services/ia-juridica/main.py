"""
Direito Lux - Módulo 4: IA Jurídica (RAG + Avaliação)
FastAPI service for legal AI with RAG and quality evaluation
"""

import logging
import os
from contextlib import asynccontextmanager
from typing import Optional

import uvicorn
from fastapi import FastAPI, HTTPException, Depends, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.responses import JSONResponse
from prometheus_fastapi_instrumentator import Instrumentator

from app.core.config import get_settings
from app.core.logging import setup_logging
from app.api.v1 import router as api_v1_router
from app.services.vector_service import VectorService
from app.services.rag_service import RAGService
from app.services.evaluation_service import EvaluationService

# Global services
vector_service: Optional[VectorService] = None
rag_service: Optional[RAGService] = None
evaluation_service: Optional[EvaluationService] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifecycle management for FastAPI application"""
    global vector_service, rag_service, evaluation_service
    
    settings = get_settings()
    
    # Setup logging
    setup_logging(settings.log_level)
    logger = logging.getLogger(__name__)
    
    # Initialize services
    logger.info("🚀 Iniciando Direito Lux - IA Jurídica")
    
    try:
        # Vector database service
        logger.info("📊 Inicializando Vector Database (Chroma)...")
        vector_service = VectorService(settings.chroma_host, settings.chroma_port)
        await vector_service.initialize()
        
        # RAG service with LangChain
        logger.info("🧠 Inicializando RAG Service (LangChain)...")
        rag_service = RAGService(vector_service, settings)
        await rag_service.initialize()
        
        # Evaluation service with Ragas
        logger.info("📈 Inicializando Evaluation Service (Ragas)...")
        evaluation_service = EvaluationService(settings)
        await evaluation_service.initialize()
        
        logger.info("✅ Todos os serviços de IA inicializados com sucesso")
        
    except Exception as e:
        logger.error(f"❌ Erro ao inicializar serviços: {e}")
        raise
    
    yield
    
    # Cleanup
    logger.info("🛑 Finalizando serviços de IA...")
    if vector_service:
        await vector_service.close()
    if rag_service:
        await rag_service.close()
    if evaluation_service:
        await evaluation_service.close()


def create_app() -> FastAPI:
    """Create and configure FastAPI application"""
    settings = get_settings()
    
    app = FastAPI(
        title="Direito Lux - IA Jurídica",
        description="API de Inteligência Artificial Jurídica com RAG e Avaliação de Qualidade",
        version="1.0.0",
        lifespan=lifespan,
        docs_url="/docs" if settings.environment != "production" else None,
        redoc_url="/redoc" if settings.environment != "production" else None,
    )
    
    # Middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.allowed_origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    app.add_middleware(
        TrustedHostMiddleware,
        allowed_hosts=settings.allowed_hosts,
    )
    
    # Prometheus metrics
    if settings.metrics_enabled:
        Instrumentator().instrument(app).expose(app)
    
    # Routes
    app.include_router(api_v1_router, prefix="/api/v1")
    
    # Health check
    @app.get("/health")
    async def health_check():
        """Health check endpoint"""
        return {
            "service": "Direito Lux - IA Jurídica",
            "status": "healthy",
            "version": "1.0.0",
            "environment": settings.environment,
            "services": {
                "vector_db": vector_service.is_healthy() if vector_service else False,
                "rag_service": rag_service.is_healthy() if rag_service else False,
                "evaluation": evaluation_service.is_healthy() if evaluation_service else False,
            }
        }
    
    # Global exception handler
    @app.exception_handler(Exception)
    async def global_exception_handler(request, exc):
        logger = logging.getLogger(__name__)
        logger.error(f"Unhandled exception: {exc}", exc_info=True)
        
        return JSONResponse(
            status_code=500,
            content={
                "error": "internal_server_error",
                "message": "Erro interno do servidor",
                "request_id": getattr(request.state, "request_id", None)
            }
        )
    
    return app


# Dependency to get services
def get_vector_service() -> VectorService:
    if vector_service is None:
        raise HTTPException(status_code=503, detail="Vector service not available")
    return vector_service


def get_rag_service() -> RAGService:
    if rag_service is None:
        raise HTTPException(status_code=503, detail="RAG service not available")
    return rag_service


def get_evaluation_service() -> EvaluationService:
    if evaluation_service is None:
        raise HTTPException(status_code=503, detail="Evaluation service not available")
    return evaluation_service


# Create app instance
app = create_app()


if __name__ == "__main__":
    settings = get_settings()
    
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.port,
        reload=settings.environment == "development",
        log_level=settings.log_level.lower(),
        workers=1 if settings.environment == "development" else settings.workers,
    )