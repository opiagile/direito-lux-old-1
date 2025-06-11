"""
Testes para o serviço RAG
"""
import pytest
from unittest.mock import Mock, patch, MagicMock
import os

# Mock environment variables before importing
os.environ.setdefault('OPENAI_API_KEY', 'test-key')
os.environ.setdefault('CHROMA_DB_HOST', 'localhost')
os.environ.setdefault('CHROMA_DB_PORT', '8000')


@pytest.fixture
def mock_dependencies():
    """Mock das dependências externas"""
    with patch('app.services.rag_service.OpenAI') as mock_openai, \
         patch('app.services.rag_service.chromadb.HttpClient') as mock_chroma:
        
        # Mock OpenAI
        mock_openai_instance = Mock()
        mock_openai.return_value = mock_openai_instance
        
        # Mock Chroma
        mock_chroma_instance = Mock()
        mock_chroma.return_value = mock_chroma_instance
        
        yield {
            'openai': mock_openai_instance,
            'chroma': mock_chroma_instance
        }


def test_rag_service_initialization(mock_dependencies):
    """Testa inicialização do serviço RAG"""
    from app.services.rag_service import RagService
    
    rag_service = RagService()
    assert rag_service is not None


@patch('app.services.rag_service.RagService._search_knowledge_base')
@patch('app.services.rag_service.RagService._generate_response')  
def test_query_success(mock_generate, mock_search, mock_dependencies):
    """Testa consulta RAG com sucesso"""
    from app.services.rag_service import RagService
    
    # Setup mocks
    mock_search.return_value = [
        {"content": "Artigo 121 do CPC", "source": "cpc.pdf", "score": 0.9}
    ]
    mock_generate.return_value = "O prazo é de 15 dias úteis"
    
    rag_service = RagService()
    result = rag_service.query("Qual o prazo para recurso?")
    
    assert result["answer"] == "O prazo é de 15 dias úteis"
    assert len(result["sources"]) > 0
    assert result["confidence"] > 0


def test_query_empty_results(mock_dependencies):
    """Testa consulta RAG sem resultados"""
    from app.services.rag_service import RagService
    
    with patch.object(RagService, '_search_knowledge_base', return_value=[]):
        rag_service = RagService()
        result = rag_service.query("Pergunta sem resposta")
        
        assert "Não encontrei" in result["answer"] or "sem informações" in result["answer"]
        assert result["confidence"] < 0.5


@patch('app.services.rag_service.RagService._get_collection_stats')
def test_get_stats(mock_stats, mock_dependencies):
    """Testa obtenção de estatísticas"""
    from app.services.rag_service import RagService
    
    mock_stats.return_value = {
        "document_count": 100,
        "chunk_count": 2000
    }
    
    rag_service = RagService()
    stats = rag_service.get_stats()
    
    assert stats["total_documents"] == 100
    assert stats["total_chunks"] == 2000
    assert "last_updated" in stats


def test_sanitize_query():
    """Testa sanitização de consultas"""
    from app.services.rag_service import RagService
    
    rag_service = RagService()
    
    # Test normal query
    clean = rag_service._sanitize_query("Qual o prazo para recurso?")
    assert clean == "Qual o prazo para recurso?"
    
    # Test with special characters
    clean = rag_service._sanitize_query("Query com <script>alert('xss')</script>")
    assert "<script>" not in clean
    
    # Test empty query
    clean = rag_service._sanitize_query("")
    assert clean == ""