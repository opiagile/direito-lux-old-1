#!/usr/bin/env python3
"""
Script para inicializar a base de conhecimento jurídico
com documentos básicos do direito brasileiro
"""

import asyncio
import os
import sys
import json
from datetime import datetime
from pathlib import Path

# Add services directory to path
sys.path.insert(0, str(Path(__file__).parent.parent / "services" / "ia-juridica"))

from app.core.config import get_settings
from app.core.logging import setup_logging, get_logger
from app.services.vector_service import VectorService
from app.services.rag_service import RAGService


# Sample legal documents for knowledge base initialization
SAMPLE_LEGAL_DOCUMENTS = [
    {
        "title": "Constituição Federal - Art. 5º (Direitos Fundamentais)",
        "content": """Art. 5º Todos são iguais perante a lei, sem distinção de qualquer natureza, garantindo-se aos brasileiros e aos estrangeiros residentes no País a inviolabilidade do direito à vida, à liberdade, à igualdade, à segurança e à propriedade, nos termos seguintes:

I - homens e mulheres são iguais em direitos e obrigações, nos termos desta Constituição;
II - ninguém será obrigado a fazer ou deixar de fazer alguma coisa senão em virtude de lei;
III - ninguém será submetido a tortura nem a tratamento desumano ou degradante;
IV - é livre a manifestação do pensamento, sendo vedado o anonimato;
V - é assegurado o direito de resposta, proporcional ao agravo, além da indenização por dano material, moral ou à imagem;
VI - é inviolável a liberdade de consciência e de crença, sendo assegurado o livre exercício dos cultos religiosos e garantida, na forma da lei, a proteção aos locais de culto e a suas liturgias;
VII - é assegurada, nos termos da lei, a prestação de assistência religiosa nas entidades civis e militares de internação coletiva;
VIII - ninguém será privado de direitos por motivo de crença religiosa ou de convicção filosófica ou política, salvo se as invocar para eximir-se de obrigação legal a todos imposta e recusar-se a cumprir prestação alternativa, fixada em lei;""",
        "metadata": {
            "source_type": "constituicao",
            "law_number": "Constituição Federal de 1988",
            "article_number": "Art. 5º",
            "chapter": "Dos Direitos e Deveres Individuais e Coletivos",
            "subject": "direitos_fundamentais",
            "keywords": ["direitos fundamentais", "igualdade", "liberdade", "dignidade humana"]
        }
    },
    {
        "title": "Código Civil - Art. 186 (Ato Ilícito)",
        "content": """Art. 186. Aquele que, por ação ou omissão voluntária, negligência ou imprudência, violar direito e causar dano a outrem, ainda que exclusivamente moral, comete ato ilícito.

Este artigo estabelece os elementos essenciais da responsabilidade civil extracontratual:
1. Conduta (ação ou omissão)
2. Culpa (negligência ou imprudência) ou dolo (ação voluntária)
3. Nexo causal
4. Dano (material ou moral)

A responsabilidade civil tem como objetivo reparar o dano causado, restabelecendo o equilíbrio patrimonial e moral da vítima.""",
        "metadata": {
            "source_type": "codigo_civil",
            "law_number": "Lei nº 10.406/2002",
            "article_number": "Art. 186",
            "book": "Parte Geral",
            "title": "Dos Atos Jurídicos Lícitos e Ilícitos",
            "subject": "responsabilidade_civil",
            "keywords": ["ato ilícito", "responsabilidade civil", "dano", "culpa", "negligência"]
        }
    },
    {
        "title": "CLT - Art. 482 (Justa Causa)",
        "content": """Art. 482 - Constituem justa causa para rescisão do contrato de trabalho pelo empregador:

a) ato de improbidade;
b) incontinência de conduta ou mau procedimento;
c) negociação habitual por conta própria ou alheia sem permissão do empregador, e quando constituir ato de concorrência à empresa para a qual trabalha o empregado, ou for prejudicial ao serviço;
d) condenação criminal do empregado, passada em julgado, caso não tenha havido suspensão da execução da pena;
e) desídia no desempenho das respectivas funções;
f) embriaguez habitual ou em serviço;
g) violação de segredo da empresa;
h) ato de indisciplina ou de insubordinação;
i) abandono de emprego;
j) ato lesivo da honra ou da boa fama praticado no serviço contra qualquer pessoa, ou ofensas físicas, nas mesmas condições, salvo em caso de legítima defesa, própria ou de outrem;
k) ato lesivo da honra ou da boa fama ou ofensas físicas praticadas contra o empregador e superiores hierárquicos, salvo em caso de legítima defesa, própria ou de outrem;
l) prática constante de jogos de azar.""",
        "metadata": {
            "source_type": "clt",
            "law_number": "Decreto-Lei nº 5.452/1943",
            "article_number": "Art. 482",
            "chapter": "Da Rescisão",
            "subject": "justa_causa",
            "keywords": ["justa causa", "rescisão", "trabalho", "empregado", "demissão"]
        }
    },
    {
        "title": "Código Penal - Art. 121 (Homicídio)",
        "content": """Art. 121. Matar alguém:
Pena - reclusão, de seis a vinte anos.

§ 1º Se o agente comete o crime impelido por motivo de relevante valor social ou moral, ou sob o domínio de violenta emoção, logo em seguida a injusta provocação da vítima, o juiz pode reduzir a pena de um sexto a um terço.

§ 2º Se o homicídio é cometido:
I - mediante paga ou promessa de recompensa, ou por outro motivo torpe;
II - por motivo fútil;
III - com emprego de veneno, fogo, explosivo, asfixia, tortura ou outro meio insidioso ou cruel, ou de que possa resultar perigo comum;
IV - à traição, de emboscada, ou mediante dissimulação ou outro recurso que dificulte ou torne impossível a defesa do ofendido;
V - para assegurar a execução, a ocultação, a impunidade ou vantagem de outro crime:
Pena - reclusão, de doze a trinta anos.

§ 3º Se o homicídio é culposo:
Pena - detenção, de um a três anos.""",
        "metadata": {
            "source_type": "codigo_penal",
            "law_number": "Decreto-Lei nº 2.848/1940",
            "article_number": "Art. 121",
            "title": "Dos Crimes Contra a Vida",
            "subject": "homicidio",
            "keywords": ["homicídio", "crime", "vida", "qualificado", "culposo", "privilegiado"]
        }
    },
    {
        "title": "Lei Maria da Penha - Art. 7º (Formas de Violência)",
        "content": """Art. 7º São formas de violência doméstica e familiar contra a mulher, entre outras:

I - a violência física, entendida como qualquer conduta que ofenda sua integridade ou saúde corporal;

II - a violência psicológica, entendida como qualquer conduta que lhe cause dano emocional e diminuição da autoestima ou que lhe prejudique e perturbe o pleno desenvolvimento ou que vise degradar ou controlar suas ações, comportamentos, crenças e decisões, mediante ameaça, constrangimento, humilhação, manipulação, isolamento, vigilância constante, perseguição contumaz, insulto, chantagem, violação de sua intimidade, ridicularização, exploração e limitação do direito de ir e vir ou qualquer outro meio que lhe cause prejuízo à saúde psicológica e à autodeterminação;

III - a violência sexual, entendida como qualquer conduta que a constranja a presenciar, a manter ou a participar de relação sexual não desejada, mediante intimidação, ameaça, coação ou uso da força;

IV - a violência patrimonial, entendida como qualquer conduta que configure retenção, subtração, destruição parcial ou total de seus objetos, instrumentos de trabalho, documentos pessoais, bens, valores e direitos ou recursos econômicos, incluindo os destinados a satisfazer suas necessidades;

V - a violência moral, entendida como qualquer conduta que configure calúnia, difamação ou injúria.""",
        "metadata": {
            "source_type": "lei_especial",
            "law_number": "Lei nº 11.340/2006",
            "article_number": "Art. 7º",
            "subject": "violencia_domestica",
            "keywords": ["Maria da Penha", "violência doméstica", "mulher", "violência física", "violência psicológica"]
        }
    }
]


async def initialize_knowledge_base():
    """Initialize the knowledge base with sample legal documents"""
    logger = get_logger(__name__)
    
    try:
        # Setup logging
        settings = get_settings()
        setup_logging(settings.log_level)
        
        logger.info("🚀 Iniciando configuração da base de conhecimento jurídico")
        
        # Initialize services
        logger.info("📊 Inicializando Vector Service")
        vector_service = VectorService(settings.chroma_host, settings.chroma_port)
        await vector_service.initialize()
        
        logger.info("🧠 Inicializando RAG Service") 
        rag_service = RAGService(vector_service, settings)
        await rag_service.initialize()
        
        # Check current knowledge base size
        stats = await vector_service.get_collection_stats()
        logger.info(f"📈 Base de conhecimento atual: {stats['total_documents']} documentos")
        
        if stats['total_documents'] > 0:
            response = input("A base de conhecimento já contém documentos. Deseja adicionar os exemplos mesmo assim? (s/N): ")
            if response.lower() not in ['s', 'sim', 'y', 'yes']:
                logger.info("❌ Operação cancelada pelo usuário")
                return
        
        # Add sample documents
        logger.info(f"📝 Adicionando {len(SAMPLE_LEGAL_DOCUMENTS)} documentos de exemplo")
        
        total_chunks = 0
        for i, doc in enumerate(SAMPLE_LEGAL_DOCUMENTS, 1):
            logger.info(f"➕ Processando documento {i}/{len(SAMPLE_LEGAL_DOCUMENTS)}: {doc['title']}")
            
            doc_ids = await rag_service.add_legal_document(
                content=doc['content'],
                metadata=doc['metadata']
            )
            
            total_chunks += len(doc_ids)
            logger.info(f"   ✅ Criado {len(doc_ids)} chunks para este documento")
        
        # Final statistics
        final_stats = await vector_service.get_collection_stats()
        logger.info(f"🎉 Base de conhecimento inicializada com sucesso!")
        logger.info(f"📊 Total de documentos: {final_stats['total_documents']}")
        logger.info(f"📦 Total de chunks adicionados: {total_chunks}")
        
        # Test a sample query
        logger.info("🔍 Testando consulta de exemplo...")
        test_result = await rag_service.process_legal_query(
            question="O que são direitos fundamentais?",
            query_type="geral"
        )
        
        logger.info(f"✅ Teste concluído - encontrados {test_result['retrieved_docs_count']} documentos relevantes")
        
        # Clean up
        await rag_service.close()
        await vector_service.close()
        
        logger.info("🏁 Configuração da base de conhecimento concluída!")
        
    except Exception as e:
        logger.error(f"❌ Erro ao inicializar base de conhecimento: {e}")
        raise


async def export_knowledge_base():
    """Export current knowledge base to JSON file"""
    logger = get_logger(__name__)
    
    try:
        settings = get_settings()
        setup_logging(settings.log_level)
        
        logger.info("📤 Exportando base de conhecimento")
        
        # Initialize vector service
        vector_service = VectorService(settings.chroma_host, settings.chroma_port)
        await vector_service.initialize()
        
        # Get all documents (this is a simplified export)
        stats = await vector_service.get_collection_stats()
        
        export_data = {
            "export_timestamp": datetime.utcnow().isoformat(),
            "total_documents": stats["total_documents"],
            "collection_name": stats["collection_name"],
            "metadata_fields": stats["metadata_fields"]
        }
        
        # Save export
        export_file = f"knowledge_base_export_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(export_file, 'w', encoding='utf-8') as f:
            json.dump(export_data, f, indent=2, ensure_ascii=False)
        
        logger.info(f"✅ Base de conhecimento exportada para: {export_file}")
        
        await vector_service.close()
        
    except Exception as e:
        logger.error(f"❌ Erro ao exportar base de conhecimento: {e}")
        raise


def main():
    """Main function"""
    if len(sys.argv) < 2:
        print("Uso: python setup-knowledge-base.py [init|export]")
        print("  init   - Inicializa a base de conhecimento com documentos de exemplo")
        print("  export - Exporta a base de conhecimento atual")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "init":
        asyncio.run(initialize_knowledge_base())
    elif command == "export":
        asyncio.run(export_knowledge_base())
    else:
        print(f"Comando desconhecido: {command}")
        sys.exit(1)


if __name__ == "__main__":
    main()