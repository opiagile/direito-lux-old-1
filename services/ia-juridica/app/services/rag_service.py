"""
RAG (Retrieval-Augmented Generation) Service using LangChain
"""

import asyncio
from typing import List, Dict, Any, Optional, Tuple
from datetime import datetime

from langchain.schema import Document
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.prompts import PromptTemplate
from langchain.chains import RetrievalQA
from langchain.llms.base import LLM
from langchain_openai import ChatOpenAI
from langchain_community.llms import VertexAI

from app.core.config import Settings
from app.core.logging import LoggingMixin
from app.services.vector_service import VectorService


class RAGService(LoggingMixin):
    """Service for Retrieval-Augmented Generation using LangChain"""
    
    def __init__(self, vector_service: VectorService, settings: Settings):
        self.vector_service = vector_service
        self.settings = settings
        self.llm = None
        self.text_splitter = None
        self.qa_chain = None
        self._is_initialized = False
        self._is_healthy = False
        
        # Legal domain prompts
        self.legal_prompts = {
            "processo": PromptTemplate(
                input_variables=["context", "question"],
                template="""Você é um assistente jurídico especializado em análise processual.

Contexto jurídico relevante:
{context}

Pergunta sobre processo: {question}

Instruções:
1. Analise o contexto fornecido cuidadosamente
2. Responda de forma precisa e fundamentada juridicamente
3. Cite as fontes legais relevantes quando possível
4. Se não houver informação suficiente, indique claramente
5. Use linguagem jurídica apropriada mas acessível

Resposta:"""
            ),
            
            "legislacao": PromptTemplate(
                input_variables=["context", "question"],
                template="""Você é um assistente jurídico especializado em legislação brasileira.

Legislação e normas relevantes:
{context}

Pergunta sobre legislação: {question}

Instruções:
1. Baseie sua resposta nas normas jurídicas fornecidas
2. Explique os artigos e dispositivos legais aplicáveis
3. Indique hierarquia normativa quando relevante
4. Mencione eventuais alterações ou revogações
5. Forneça interpretação jurídica fundamentada

Resposta:"""
            ),
            
            "jurisprudencia": PromptTemplate(
                input_variables=["context", "question"],
                template="""Você é um assistente jurídico especializado em jurisprudência.

Precedentes e decisões judiciais relevantes:
{context}

Pergunta sobre jurisprudência: {question}

Instruções:
1. Analise os precedentes apresentados
2. Explique as teses jurídicas predominantes
3. Indique possíveis divergências entre tribunais
4. Contextualize com a legislação aplicável
5. Avalie a aplicabilidade ao caso concreto

Resposta:"""
            ),
            
            "geral": PromptTemplate(
                input_variables=["context", "question"],
                template="""Você é um assistente jurídico especializado em Direito brasileiro.

Informações jurídicas relevantes:
{context}

Pergunta: {question}

Instruções:
1. Forneça resposta fundamentada juridicamente
2. Use fontes confiáveis (leis, jurisprudência, doutrina)
3. Seja preciso e objetivo
4. Indique limitações ou necessidade de análise adicional
5. Mantenha linguagem técnica mas compreensível

Resposta:"""
            )
        }
    
    async def initialize(self) -> None:
        """Initialize RAG service components"""
        try:
            self.log_info("Inicializando RAG Service")
            
            # Initialize LLM based on configuration
            await self._initialize_llm()
            
            # Initialize text splitter for document chunking
            self.text_splitter = RecursiveCharacterTextSplitter(
                chunk_size=self.settings.chunk_size,
                chunk_overlap=self.settings.chunk_overlap,
                separators=["\n\n", "\n", ". ", " ", ""]
            )
            
            self._is_initialized = True
            self._is_healthy = True
            self.log_info("RAG Service inicializado com sucesso")
            
        except Exception as e:
            self._is_healthy = False
            self.log_error("Erro ao inicializar RAG Service", error=str(e))
            raise
    
    async def _initialize_llm(self) -> None:
        """Initialize the Language Model"""
        try:
            if self.settings.llm_provider == "openai":
                if not self.settings.openai_api_key:
                    raise ValueError("OpenAI API key not configured")
                
                self.llm = ChatOpenAI(
                    model=self.settings.openai_model,
                    api_key=self.settings.openai_api_key,
                    temperature=0.1,  # Low temperature for consistent legal responses
                    max_tokens=2048
                )
                self.log_info("OpenAI LLM inicializado", model=self.settings.openai_model)
                
            elif self.settings.llm_provider == "vertex_ai":
                if not self.settings.google_cloud_project:
                    raise ValueError("Google Cloud project not configured")
                
                self.llm = VertexAI(
                    model_name=self.settings.vertex_ai_model,
                    project=self.settings.google_cloud_project,
                    location=self.settings.google_cloud_location,
                    temperature=0.1,
                    max_output_tokens=2048
                )
                self.log_info("Vertex AI LLM inicializado", model=self.settings.vertex_ai_model)
                
            else:
                raise ValueError(f"Unsupported LLM provider: {self.settings.llm_provider}")
                
        except Exception as e:
            self.log_error("Erro ao inicializar LLM", error=str(e))
            raise
    
    async def process_legal_query(
        self, 
        question: str, 
        query_type: str = "geral",
        filters: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Process a legal query using RAG"""
        try:
            if not self._is_initialized:
                raise RuntimeError("RAG service not initialized")
            
            start_time = datetime.utcnow()
            
            # Step 1: Retrieve relevant documents
            self.log_info("Iniciando busca por documentos relevantes", 
                         question_length=len(question), 
                         query_type=query_type)
            
            relevant_docs = await self.vector_service.search_similar(
                query=question,
                top_k=self.settings.retrieval_top_k,
                filters=filters
            )
            
            # Filter by similarity threshold
            filtered_docs = [
                doc for doc in relevant_docs 
                if doc["similarity_score"] >= self.settings.similarity_threshold
            ]
            
            if not filtered_docs:
                return {
                    "answer": "Não foram encontrados documentos jurídicos relevantes para sua consulta. "
                             "Por favor, reformule sua pergunta ou forneça mais contexto.",
                    "sources": [],
                    "query_type": query_type,
                    "processing_time": (datetime.utcnow() - start_time).total_seconds(),
                    "retrieved_docs_count": 0
                }
            
            # Step 2: Prepare context from retrieved documents
            context = self._prepare_context(filtered_docs)
            
            # Step 3: Generate response using appropriate prompt
            prompt_template = self.legal_prompts.get(query_type, self.legal_prompts["geral"])
            prompt = prompt_template.format(context=context, question=question)
            
            self.log_info("Gerando resposta com LLM", 
                         context_length=len(context),
                         docs_used=len(filtered_docs))
            
            # Generate response
            response = await asyncio.to_thread(self.llm.invoke, prompt)
            
            # Extract answer from response
            if hasattr(response, 'content'):
                answer = response.content
            else:
                answer = str(response)
            
            # Prepare sources information
            sources = self._prepare_sources(filtered_docs)
            
            processing_time = (datetime.utcnow() - start_time).total_seconds()
            
            result = {
                "answer": answer,
                "sources": sources,
                "query_type": query_type,
                "processing_time": processing_time,
                "retrieved_docs_count": len(filtered_docs),
                "total_docs_found": len(relevant_docs)
            }
            
            self.log_info("Consulta jurídica processada com sucesso", 
                         processing_time=processing_time,
                         answer_length=len(answer))
            
            return result
            
        except Exception as e:
            self.log_error("Erro ao processar consulta jurídica", error=str(e))
            raise
    
    def _prepare_context(self, documents: List[Dict[str, Any]]) -> str:
        """Prepare context string from retrieved documents"""
        context_parts = []
        
        for i, doc in enumerate(documents, 1):
            metadata = doc.get("metadata", {})
            source_info = ""
            
            # Add source information
            if metadata.get("title"):
                source_info += f"Título: {metadata['title']}\n"
            if metadata.get("source_type"):
                source_info += f"Tipo: {metadata['source_type']}\n"
            if metadata.get("article_number"):
                source_info += f"Artigo: {metadata['article_number']}\n"
            if metadata.get("law_number"):
                source_info += f"Lei: {metadata['law_number']}\n"
            
            context_part = f"--- Documento {i} ---\n"
            if source_info:
                context_part += source_info + "\n"
            context_part += f"Conteúdo: {doc['document']}\n"
            context_part += f"Relevância: {doc['similarity_score']:.3f}\n\n"
            
            context_parts.append(context_part)
        
        return "".join(context_parts)
    
    def _prepare_sources(self, documents: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Prepare sources information for response"""
        sources = []
        
        for doc in documents:
            metadata = doc.get("metadata", {})
            source = {
                "id": doc["id"],
                "title": metadata.get("title", "Documento Jurídico"),
                "source_type": metadata.get("source_type", "unknown"),
                "similarity_score": doc["similarity_score"],
            }
            
            # Add specific legal information
            if metadata.get("law_number"):
                source["law_number"] = metadata["law_number"]
            if metadata.get("article_number"):
                source["article_number"] = metadata["article_number"]
            if metadata.get("court"):
                source["court"] = metadata["court"]
            if metadata.get("date"):
                source["date"] = metadata["date"]
            
            sources.append(source)
        
        return sources
    
    async def add_legal_document(
        self, 
        content: str, 
        metadata: Dict[str, Any],
        doc_id: Optional[str] = None
    ) -> List[str]:
        """Add a legal document to the knowledge base"""
        try:
            if not self._is_initialized:
                raise RuntimeError("RAG service not initialized")
            
            # Split document into chunks
            documents = self.text_splitter.split_text(content)
            
            # Prepare metadata for each chunk
            chunk_metadatas = []
            for i, chunk in enumerate(documents):
                chunk_metadata = metadata.copy()
                chunk_metadata.update({
                    "chunk_index": i,
                    "total_chunks": len(documents),
                    "chunk_size": len(chunk),
                    "document_type": "legal_document"
                })
                chunk_metadatas.append(chunk_metadata)
            
            # Add to vector database
            doc_ids = await self.vector_service.add_documents(
                documents=documents,
                metadatas=chunk_metadatas,
                ids=None  # Let vector service generate IDs
            )
            
            self.log_info("Documento jurídico adicionado à base de conhecimento",
                         chunks_created=len(documents),
                         doc_type=metadata.get("source_type", "unknown"))
            
            return doc_ids
            
        except Exception as e:
            self.log_error("Erro ao adicionar documento jurídico", error=str(e))
            raise
    
    async def get_knowledge_base_stats(self) -> Dict[str, Any]:
        """Get knowledge base statistics"""
        try:
            if not self._is_initialized:
                raise RuntimeError("RAG service not initialized")
            
            vector_stats = await self.vector_service.get_collection_stats()
            
            return {
                "total_documents": vector_stats["total_documents"],
                "collection_name": vector_stats["collection_name"],
                "llm_provider": self.settings.llm_provider,
                "llm_model": (self.settings.openai_model 
                            if self.settings.llm_provider == "openai" 
                            else self.settings.vertex_ai_model),
                "chunk_size": self.settings.chunk_size,
                "retrieval_top_k": self.settings.retrieval_top_k,
                "similarity_threshold": self.settings.similarity_threshold,
                "status": "healthy" if self._is_healthy else "unhealthy"
            }
            
        except Exception as e:
            self.log_error("Erro ao obter estatísticas da base de conhecimento", error=str(e))
            raise
    
    def is_healthy(self) -> bool:
        """Check if the RAG service is healthy"""
        return (self._is_healthy and 
                self._is_initialized and 
                self.vector_service.is_healthy())
    
    async def close(self) -> None:
        """Close RAG service"""
        try:
            self.log_info("Fechando RAG Service")
            self._is_initialized = False
            self._is_healthy = False
            
        except Exception as e:
            self.log_error("Erro ao fechar RAG Service", error=str(e))