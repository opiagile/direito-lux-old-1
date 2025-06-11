#!/usr/bin/env python3
"""
Versão simplificada do serviço de IA para teste rápido
"""

import os
import asyncio
from fastapi import FastAPI
from pydantic import BaseModel
from typing import List, Dict, Any, Optional
import uvicorn

# Configuração simples
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "sk-xxx-configure-sua-chave")

app = FastAPI(
    title="Direito Lux - IA Jurídica (Teste)",
    description="Versão simplificada para teste",
    version="1.0.0-test"
)

# Modelos de requisição
class LegalQueryRequest(BaseModel):
    question: str
    query_type: str = "geral"
    evaluate_response: bool = False

class LegalQueryResponse(BaseModel):
    query_id: str
    question: str
    answer: str
    query_type: str
    sources: List[Dict[str, Any]]
    processing_time: float

# Base de conhecimento simulada
MOCK_KNOWLEDGE_BASE = {
    "direitos_fundamentais": {
        "content": "Art. 5º Todos são iguais perante a lei, sem distinção de qualquer natureza, garantindo-se aos brasileiros e aos estrangeiros residentes no País a inviolabilidade do direito à vida, à liberdade, à igualdade, à segurança e à propriedade.",
        "source": "Constituição Federal - Art. 5º",
        "law_number": "Constituição Federal de 1988"
    },
    "responsabilidade_civil": {
        "content": "Art. 186. Aquele que, por ação ou omissão voluntária, negligência ou imprudência, violar direito e causar dano a outrem, ainda que exclusivamente moral, comete ato ilícito.",
        "source": "Código Civil - Art. 186",
        "law_number": "Lei nº 10.406/2002"
    },
    "justa_causa": {
        "content": "Art. 482 - Constituem justa causa para rescisão do contrato de trabalho pelo empregador: a) ato de improbidade; b) incontinência de conduta ou mau procedimento; c) negociação habitual por conta própria...",
        "source": "CLT - Art. 482",
        "law_number": "Decreto-Lei nº 5.452/1943"
    }
}

def mock_search_knowledge(question: str) -> List[Dict[str, Any]]:
    """Busca simulada na base de conhecimento"""
    results = []
    question_lower = question.lower()
    
    for key, data in MOCK_KNOWLEDGE_BASE.items():
        if any(term in question_lower for term in key.split('_')):
            results.append({
                "id": key,
                "document": data["content"],
                "similarity_score": 0.85,
                "metadata": {
                    "title": data["source"],
                    "law_number": data["law_number"],
                    "source_type": "mock_data"
                }
            })
    
    return results[:3]  # Retorna no máximo 3 resultados

def mock_generate_answer(question: str, context: str) -> str:
    """Geração de resposta simulada (sem OpenAI para teste)"""
    if "direitos fundamentais" in question.lower():
        return """Os direitos fundamentais são direitos básicos e essenciais reconhecidos e protegidos pela Constituição Federal brasileira. Estão previstos principalmente no Art. 5º da Constituição Federal de 1988.

Segundo o Art. 5º: "Todos são iguais perante a lei, sem distinção de qualquer natureza, garantindo-se aos brasileiros e aos estrangeiros residentes no País a inviolabilidade do direito à vida, à liberdade, à igualdade, à segurança e à propriedade."

Os direitos fundamentais incluem:
- Direito à vida
- Direito à liberdade
- Direito à igualdade
- Direito à segurança
- Direito à propriedade

Estes direitos são considerados cláusulas pétreas e não podem ser abolidos nem mesmo por emenda constitucional."""

    elif "responsabilidade civil" in question.lower():
        return """A responsabilidade civil está prevista no Art. 186 do Código Civil brasileiro. Segundo este artigo: "Aquele que, por ação ou omissão voluntária, negligência ou imprudência, violar direito e causar dano a outrem, ainda que exclusivamente moral, comete ato ilícito."

Os elementos da responsabilidade civil são:
1. **Conduta**: ação ou omissão do agente
2. **Culpa ou Dolo**: negligência, imprudência ou ação intencional
3. **Nexo Causal**: relação de causa e efeito entre a conduta e o dano
4. **Dano**: prejuízo material ou moral causado à vítima

O objetivo da responsabilidade civil é reparar o dano causado, restabelecendo o equilíbrio patrimonial e moral da vítima."""

    elif "justa causa" in question.lower():
        return """A justa causa trabalhista está prevista no Art. 482 da CLT (Consolidação das Leis do Trabalho). As principais hipóteses de justa causa são:

a) Ato de improbidade
b) Incontinência de conduta ou mau procedimento
c) Negociação habitual por conta própria sem permissão do empregador
d) Condenação criminal transitada em julgado
e) Desídia no desempenho das funções
f) Embriaguez habitual ou em serviço
g) Violação de segredo da empresa
h) Ato de indisciplina ou insubordinação
i) Abandono de emprego
j) Ato lesivo da honra ou ofensas físicas no serviço
k) Ato lesivo contra empregador e superiores
l) Prática constante de jogos de azar

A justa causa permite a demissão sem aviso prévio e sem direito a algumas verbas rescisórias."""
    
    else:
        return f"""Com base no contexto jurídico disponível, posso fornecer informações sobre sua consulta: "{question}".

{context}

Para uma análise mais específica, recomendo consultar um advogado especializado na área ou verificar a legislação aplicável ao seu caso concreto."""

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "service": "Direito Lux - IA Jurídica (Teste)",
        "status": "healthy",
        "version": "1.0.0-test",
        "environment": "test",
        "services": {
            "mock_knowledge_base": True,
            "mock_llm": True
        }
    }

@app.post("/api/v1/rag/query", response_model=LegalQueryResponse)
async def legal_query(request: LegalQueryRequest):
    """Consulta jurídica com RAG simulado"""
    import time
    start_time = time.time()
    
    # Simular ID da consulta
    query_id = f"test_query_{int(time.time())}"
    
    # Buscar na base de conhecimento simulada
    relevant_docs = mock_search_knowledge(request.question)
    
    # Preparar contexto
    context = ""
    sources = []
    
    for doc in relevant_docs:
        context += f"{doc['document']}\n\n"
        sources.append({
            "id": doc["id"],
            "title": doc["metadata"]["title"],
            "law_number": doc["metadata"]["law_number"],
            "similarity_score": doc["similarity_score"],
            "source_type": doc["metadata"]["source_type"]
        })
    
    # Gerar resposta simulada
    answer = mock_generate_answer(request.question, context)
    
    processing_time = time.time() - start_time
    
    return LegalQueryResponse(
        query_id=query_id,
        question=request.question,
        answer=answer,
        query_type=request.query_type,
        sources=sources,
        processing_time=processing_time
    )

@app.get("/api/v1/rag/stats")
async def get_rag_stats():
    """Estatísticas do sistema RAG"""
    return {
        "total_documents": len(MOCK_KNOWLEDGE_BASE),
        "collection_name": "mock_legal_docs",
        "llm_provider": "mock",
        "llm_model": "mock-legal-model",
        "status": "healthy"
    }

@app.get("/api/v1/knowledge/stats")
async def get_knowledge_stats():
    """Estatísticas da base de conhecimento"""
    return {
        "total_documents": len(MOCK_KNOWLEDGE_BASE),
        "collection_name": "mock_legal_docs",
        "metadata_fields": ["title", "law_number", "source_type"],
        "status": "healthy"
    }

@app.get("/docs")
async def get_docs():
    """Redirect to automatic docs"""
    from fastapi.responses import RedirectResponse
    return RedirectResponse(url="/docs")

if __name__ == "__main__":
    print("🧠 Iniciando Direito Lux - IA Jurídica (Modo Teste)")
    print("📊 Usando base de conhecimento simulada")
    print("🔗 API disponível em: http://localhost:9003")
    print("📖 Documentação em: http://localhost:9003/docs")
    
    uvicorn.run(
        "test_service:app",
        host="0.0.0.0",
        port=9003,
        reload=True,
        log_level="info"
    )