"""
API v1 package for Direito Lux IA Jurídica
"""

from fastapi import APIRouter

from .rag import router as rag_router
from .evaluation import router as evaluation_router
from .knowledge import router as knowledge_router

# Main v1 router
router = APIRouter()

# Include sub-routers
router.include_router(rag_router, prefix="/rag", tags=["RAG"])
router.include_router(evaluation_router, prefix="/evaluation", tags=["Evaluation"])
router.include_router(knowledge_router, prefix="/knowledge", tags=["Knowledge Base"])