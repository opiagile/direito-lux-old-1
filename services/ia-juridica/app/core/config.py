"""
Configuration settings for Direito Lux IA Jurídica
"""

import os
from functools import lru_cache
from typing import List, Optional

from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings"""
    
    # Application
    app_name: str = Field(default="Direito Lux - IA Jurídica")
    environment: str = Field(default="development")
    port: int = Field(default=9003)
    workers: int = Field(default=4)
    log_level: str = Field(default="INFO")
    
    # Security
    secret_key: str = Field(default="your-secret-key-change-in-production")
    allowed_origins: List[str] = Field(default=["*"])
    allowed_hosts: List[str] = Field(default=["*"])
    
    # Vector Database (Chroma)
    chroma_host: str = Field(default="localhost")
    chroma_port: int = Field(default=8000)
    chroma_collection_name: str = Field(default="direito_lux_legal_docs")
    
    # LLM Configuration
    llm_provider: str = Field(default="openai")  # openai, vertex_ai
    openai_api_key: Optional[str] = Field(default=None)
    openai_model: str = Field(default="gpt-4-turbo-preview")
    
    # Vertex AI (Google Cloud)
    google_cloud_project: Optional[str] = Field(default=None)
    google_cloud_location: str = Field(default="us-central1")
    vertex_ai_model: str = Field(default="text-bison")
    
    # Embeddings
    embedding_model: str = Field(default="sentence-transformers/all-MiniLM-L6-v2")
    embedding_dimension: int = Field(default=384)
    
    # RAG Configuration
    retrieval_top_k: int = Field(default=5)
    chunk_size: int = Field(default=1000)
    chunk_overlap: int = Field(default=200)
    similarity_threshold: float = Field(default=0.7)
    
    # Evaluation (Ragas)
    evaluation_enabled: bool = Field(default=True)
    evaluation_batch_size: int = Field(default=10)
    
    # Observability
    metrics_enabled: bool = Field(default=True)
    tracing_enabled: bool = Field(default=True)
    
    # Redis (for caching)
    redis_host: str = Field(default="localhost")
    redis_port: int = Field(default=6379)
    redis_db: int = Field(default=2)
    redis_password: Optional[str] = Field(default=None)
    
    # Data Loss Prevention (DLP)
    dlp_enabled: bool = Field(default=True)
    google_cloud_dlp_project: Optional[str] = Field(default=None)
    
    # Legal Knowledge Base
    knowledge_base_path: str = Field(default="./data/knowledge_base")
    update_knowledge_base: bool = Field(default=False)
    
    # Rate Limiting
    rate_limit_enabled: bool = Field(default=True)
    rate_limit_requests: int = Field(default=100)
    rate_limit_window: int = Field(default=3600)  # 1 hour
    
    # Background Tasks
    celery_broker_url: str = Field(default="redis://localhost:6379/3")
    celery_result_backend: str = Field(default="redis://localhost:6379/3")
    
    class Config:
        env_prefix = "DIREITO_LUX_IA_"
        env_file = ".env"
        env_file_encoding = "utf-8"


@lru_cache()
def get_settings() -> Settings:
    """Get cached settings instance"""
    return Settings()