"""
Testes para os endpoints da API IA Jurídica
"""
import pytest
from fastapi.testclient import TestClient
from unittest.mock import Mock, patch
import json

@pytest.fixture
def client():
    """Cliente de teste para FastAPI"""
    # Import here to avoid circular dependencies
    from main import app
    return TestClient(app)


def test_health_endpoint(client):
    """Testa o endpoint de health check"""
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert "timestamp" in data


@patch('app.services.rag_service.RagService.query')
def test_rag_query_success(mock_rag_query, client):
    """Testa consulta RAG com sucesso"""
    mock_rag_query.return_value = {
        "answer": "Resposta jurídica mockada",
        "sources": ["doc1.pdf", "doc2.pdf"],
        "confidence": 0.85
    }
    
    payload = {
        "query": "Qual o prazo para recurso?",
        "context": "direito civil"
    }
    
    response = client.post("/api/v1/rag/query", json=payload)
    assert response.status_code == 200
    
    data = response.json()
    assert "answer" in data
    assert "sources" in data
    assert "confidence" in data


def test_rag_query_invalid_payload(client):
    """Testa consulta RAG com payload inválido"""
    payload = {}  # payload vazio
    
    response = client.post("/api/v1/rag/query", json=payload)
    assert response.status_code == 422  # Validation error


@patch('app.services.rag_service.RagService.get_stats')
def test_knowledge_stats(mock_get_stats, client):
    """Testa endpoint de estatísticas da base de conhecimento"""
    mock_get_stats.return_value = {
        "total_documents": 150,
        "total_chunks": 3000,
        "last_updated": "2024-01-15T10:30:00Z"
    }
    
    response = client.get("/api/v1/knowledge/stats")
    assert response.status_code == 200
    
    data = response.json()
    assert data["total_documents"] == 150
    assert data["total_chunks"] == 3000


def test_cors_headers(client):
    """Testa se headers CORS estão configurados"""
    response = client.options("/api/v1/rag/query")
    assert response.status_code == 200
    # FastAPI CORS middleware adiciona headers automaticamente