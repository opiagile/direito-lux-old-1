"""
Vector Database Service using ChromaDB for RAG implementation
"""

import asyncio
import uuid
from typing import List, Dict, Any, Optional, Tuple
from datetime import datetime

import chromadb
from chromadb.config import Settings as ChromaSettings
from chromadb.utils import embedding_functions
from sentence_transformers import SentenceTransformer

from app.core.config import get_settings
from app.core.logging import LoggingMixin


class VectorService(LoggingMixin):
    """Service for managing vector database operations with ChromaDB"""
    
    def __init__(self, host: str = "localhost", port: int = 8000):
        self.host = host
        self.port = port
        self.client = None
        self.collection = None
        self.embedding_function = None
        self.embedding_model = None
        self._is_initialized = False
        self._is_healthy = False
        
    async def initialize(self) -> None:
        """Initialize ChromaDB client and collection"""
        try:
            settings = get_settings()
            
            # Initialize ChromaDB client
            self.log_info("Conectando ao ChromaDB", host=self.host, port=self.port)
            
            # Use HTTP client for remote ChromaDB
            self.client = chromadb.HttpClient(
                host=self.host,
                port=self.port,
                settings=ChromaSettings(
                    chroma_client_auth_provider="chromadb.auth.basic.BasicAuthClientProvider",
                    chroma_client_auth_credentials="admin:admin"
                )
            )
            
            # Initialize embedding model
            self.log_info("Carregando modelo de embeddings", model=settings.embedding_model)
            self.embedding_model = SentenceTransformer(settings.embedding_model)
            
            # Create embedding function
            self.embedding_function = embedding_functions.SentenceTransformerEmbeddingFunction(
                model_name=settings.embedding_model
            )
            
            # Get or create collection
            try:
                self.collection = self.client.get_collection(
                    name=settings.chroma_collection_name,
                    embedding_function=self.embedding_function
                )
                self.log_info("Collection existente encontrada", 
                            collection=settings.chroma_collection_name)
            except Exception:
                self.log_info("Criando nova collection", 
                            collection=settings.chroma_collection_name)
                self.collection = self.client.create_collection(
                    name=settings.chroma_collection_name,
                    embedding_function=self.embedding_function,
                    metadata={"description": "Legal documents for Direito Lux RAG system"}
                )
            
            # Test connection
            await self._test_connection()
            
            self._is_initialized = True
            self._is_healthy = True
            self.log_info("Vector Database inicializado com sucesso")
            
        except Exception as e:
            self._is_healthy = False
            self.log_error("Erro ao inicializar Vector Database", error=str(e))
            raise
    
    async def _test_connection(self) -> None:
        """Test ChromaDB connection"""
        try:
            # Simple test query
            test_result = await asyncio.to_thread(self.collection.count)
            self.log_debug("Connection test successful", document_count=test_result)
        except Exception as e:
            self.log_error("Connection test failed", error=str(e))
            raise
    
    async def add_documents(
        self, 
        documents: List[str], 
        metadatas: List[Dict[str, Any]], 
        ids: Optional[List[str]] = None
    ) -> List[str]:
        """Add documents to the vector database"""
        try:
            if not self._is_initialized:
                raise RuntimeError("Vector service not initialized")
            
            # Generate IDs if not provided
            if ids is None:
                ids = [str(uuid.uuid4()) for _ in documents]
            
            # Add timestamp to metadata
            for metadata in metadatas:
                metadata["added_at"] = datetime.utcnow().isoformat()
                metadata["source"] = "direito_lux_system"
            
            # Add documents to collection
            await asyncio.to_thread(
                self.collection.add,
                documents=documents,
                metadatas=metadatas,
                ids=ids
            )
            
            self.log_info("Documentos adicionados à base vetorial", 
                         count=len(documents), ids=ids[:3])
            
            return ids
            
        except Exception as e:
            self.log_error("Erro ao adicionar documentos", error=str(e))
            raise
    
    async def search_similar(
        self, 
        query: str, 
        top_k: int = 5,
        filters: Optional[Dict[str, Any]] = None
    ) -> List[Dict[str, Any]]:
        """Search for similar documents using semantic similarity"""
        try:
            if not self._is_initialized:
                raise RuntimeError("Vector service not initialized")
            
            # Prepare where clause for filtering
            where_clause = filters if filters else None
            
            # Search similar documents
            results = await asyncio.to_thread(
                self.collection.query,
                query_texts=[query],
                n_results=top_k,
                where=where_clause,
                include=["documents", "metadatas", "distances"]
            )
            
            # Format results
            formatted_results = []
            if results and results["documents"] and results["documents"][0]:
                for i in range(len(results["documents"][0])):
                    formatted_results.append({
                        "id": results["ids"][0][i],
                        "document": results["documents"][0][i],
                        "metadata": results["metadatas"][0][i],
                        "similarity_score": 1 - results["distances"][0][i],  # Convert distance to similarity
                        "distance": results["distances"][0][i]
                    })
            
            self.log_info("Busca semântica realizada", 
                         query_length=len(query), 
                         results_count=len(formatted_results))
            
            return formatted_results
            
        except Exception as e:
            self.log_error("Erro na busca semântica", error=str(e))
            raise
    
    async def get_document_by_id(self, doc_id: str) -> Optional[Dict[str, Any]]:
        """Get a specific document by ID"""
        try:
            if not self._is_initialized:
                raise RuntimeError("Vector service not initialized")
            
            results = await asyncio.to_thread(
                self.collection.get,
                ids=[doc_id],
                include=["documents", "metadatas"]
            )
            
            if results and results["documents"]:
                return {
                    "id": doc_id,
                    "document": results["documents"][0],
                    "metadata": results["metadatas"][0]
                }
            
            return None
            
        except Exception as e:
            self.log_error("Erro ao buscar documento por ID", doc_id=doc_id, error=str(e))
            raise
    
    async def delete_documents(self, ids: List[str]) -> None:
        """Delete documents from the vector database"""
        try:
            if not self._is_initialized:
                raise RuntimeError("Vector service not initialized")
            
            await asyncio.to_thread(self.collection.delete, ids=ids)
            
            self.log_info("Documentos removidos da base vetorial", ids=ids)
            
        except Exception as e:
            self.log_error("Erro ao remover documentos", error=str(e))
            raise
    
    async def update_document(
        self, 
        doc_id: str, 
        document: str, 
        metadata: Dict[str, Any]
    ) -> None:
        """Update a document in the vector database"""
        try:
            if not self._is_initialized:
                raise RuntimeError("Vector service not initialized")
            
            # Add update timestamp
            metadata["updated_at"] = datetime.utcnow().isoformat()
            
            await asyncio.to_thread(
                self.collection.update,
                ids=[doc_id],
                documents=[document],
                metadatas=[metadata]
            )
            
            self.log_info("Documento atualizado", doc_id=doc_id)
            
        except Exception as e:
            self.log_error("Erro ao atualizar documento", doc_id=doc_id, error=str(e))
            raise
    
    async def get_collection_stats(self) -> Dict[str, Any]:
        """Get collection statistics"""
        try:
            if not self._is_initialized:
                raise RuntimeError("Vector service not initialized")
            
            count = await asyncio.to_thread(self.collection.count)
            
            # Get sample of documents for metadata analysis
            sample_results = await asyncio.to_thread(
                self.collection.get,
                limit=10,
                include=["metadatas"]
            )
            
            # Analyze metadata
            metadata_keys = set()
            for metadata in sample_results.get("metadatas", []):
                if metadata:
                    metadata_keys.update(metadata.keys())
            
            return {
                "total_documents": count,
                "collection_name": self.collection.name,
                "metadata_fields": list(metadata_keys),
                "status": "healthy" if self._is_healthy else "unhealthy"
            }
            
        except Exception as e:
            self.log_error("Erro ao obter estatísticas da collection", error=str(e))
            raise
    
    async def embed_text(self, text: str) -> List[float]:
        """Generate embeddings for text using the configured model"""
        try:
            if not self.embedding_model:
                raise RuntimeError("Embedding model not initialized")
            
            # Generate embeddings
            embeddings = await asyncio.to_thread(
                self.embedding_model.encode, 
                [text], 
                convert_to_tensor=False
            )
            
            return embeddings[0].tolist()
            
        except Exception as e:
            self.log_error("Erro ao gerar embeddings", error=str(e))
            raise
    
    def is_healthy(self) -> bool:
        """Check if the vector service is healthy"""
        return self._is_healthy and self._is_initialized
    
    async def close(self) -> None:
        """Close vector database connection"""
        try:
            if self.client:
                self.log_info("Fechando conexão com Vector Database")
                # ChromaDB HTTP client doesn't need explicit closing
                self.client = None
                
            self._is_initialized = False
            self._is_healthy = False
            
        except Exception as e:
            self.log_error("Erro ao fechar conexão", error=str(e))