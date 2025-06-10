# Módulo 3 - Consulta Jurídica + Circuit Breaker

Este módulo implementa um serviço de consultas jurídicas com padrão Circuit Breaker para resiliência e stack completa ELK para observabilidade.

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                    MÓDULO 3 - CONSULTA JURÍDICA             │
├─────────────────────────────────────────────────────────────┤
│  Kong Gateway (8002) → Circuit Breaker → Consulta Service   │
│                    ↓                                        │
│  ELK Stack ← OpenTelemetry ← Go Services → External APIs    │
│                    ↓                                        │
│  [Elasticsearch] [Logstash] [Kibana] [Prometheus]          │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Componentes

### 1. Consulta Service (Go)
- **Porta:** 9002
- **Funcionalidades:**
  - Consulta de processos judiciais
  - Consulta de legislação
  - Consulta de jurisprudência
  - Circuit breaker integrado
  - Logging estruturado

### 2. Circuit Breaker
- **Padrão:** Hystrix-like
- **Estados:** Closed, Open, Half-Open
- **Configuração:** Falha em 60% das requisições
- **Timeout:** 60 segundos no estado aberto
- **Métricas:** Expostas via API REST

### 3. ELK Stack

#### Elasticsearch
- **Porta:** 9200
- **Função:** Armazenamento de logs
- **Índices:** `direito-lux-logs-YYYY.MM.dd`

#### Logstash
- **Porta:** 5000 (TCP), 5044 (Beats), 9600 (API)
- **Função:** Processamento de logs
- **Pipeline:** Parsing, enriquecimento, indexação

#### Kibana
- **Porta:** 5601
- **Função:** Visualização e dashboards
- **Dashboards:** Métricas de consulta e circuit breaker

## 📊 APIs Disponíveis

### Health Check
```bash
GET /health
```

### Circuit Breaker Status
```bash
GET /api/v1/circuit-breaker/status
```

### Consulta de Processo
```bash
POST /api/v1/consultas/processos
Content-Type: application/json

{
  "numero_processo": "1234567-89.2023.1.23.4567",
  "tribunal": "TJSP"
}
```

### Consulta de Legislação
```bash
POST /api/v1/consultas/legislacao
Content-Type: application/json

{
  "tema": "direito civil",
  "jurisdicao": "federal"
}
```

### Consulta de Jurisprudência
```bash
POST /api/v1/consultas/jurisprudencia
Content-Type: application/json

{
  "tema": "responsabilidade civil",
  "tribunal": "STJ"
}
```

## 🔧 Como Usar

### Iniciar Ambiente
```bash
./scripts/consulta-start.sh
```

### Verificar Status
```bash
./scripts/consulta-status.sh
```

### Executar Testes
```bash
./scripts/consulta-test.sh
```

### Parar Ambiente
```bash
./scripts/consulta-stop.sh
```

## 📊 Monitoramento

### URLs de Monitoramento
- **Elasticsearch:** http://localhost:9200
- **Kibana:** http://localhost:5601
- **Logstash:** http://localhost:9600
- **Consulta API:** http://localhost:9002

### Métricas Disponíveis
- Requisições por minuto
- Taxa de sucesso/falha
- Estado do circuit breaker
- Latência das consultas
- Geolocalização dos requests

### Dashboards Kibana
1. **Consultas Overview:** Visão geral das consultas
2. **Circuit Breaker:** Estado e métricas do circuit breaker
3. **Error Analysis:** Análise de erros e falhas
4. **Performance:** Métricas de performance

## 🧪 Testes

### Teste Manual
```bash
# Health check
curl http://localhost:9002/health

# Circuit breaker status
curl http://localhost:9002/api/v1/circuit-breaker/status

# Consulta normal
curl -X POST http://localhost:9002/api/v1/consultas/processos \
  -H "Content-Type: application/json" \
  -d '{"numero_processo": "1234567-89.2023.1.23.4567", "tribunal": "TJSP"}'

# Testar falha (circuit breaker)
curl -X POST http://localhost:9002/api/v1/consultas/processos \
  -H "Content-Type: application/json" \
  -d '{"numero_processo": "0000000-00.0000.0.00.0000", "tribunal": "FAIL"}'
```

### Teste Automatizado
```bash
./scripts/consulta-test.sh
```

## 🔍 Logs

### Visualizar Logs
```bash
# Logs do serviço de consulta
docker logs -f consulta-service

# Logs de todos os serviços
docker compose -f docker-compose.consulta.yml logs -f

# Logs no Kibana
# Acesse: http://localhost:5601
# Índice: direito-lux-logs-*
```

### Estrutura dos Logs
```json
{
  "@timestamp": "2023-06-10T15:30:00.000Z",
  "level": "info",
  "message": "Consulta de processo concluída",
  "service": "direito-lux-consulta",
  "module": "module-3",
  "request_id": "req-123456",
  "consulta_id": "consulta-789",
  "tribunal": "TJSP",
  "circuit_breaker_state": "closed"
}
```

## 🔐 Segurança

### Headers de Segurança
- X-Request-ID para correlação
- Rate limiting via Kong Gateway
- Validação de entrada nos endpoints

### Logging Seguro
- Dados sensíveis não são logados
- IPs são anonimizados para LGPD
- Request IDs para auditoria

## 🚨 Circuit Breaker

### Configuração
- **Máximo de requisições:** 3 no estado half-open
- **Intervalo:** 10 segundos para resetar contadores
- **Timeout:** 60 segundos no estado aberto
- **Threshold:** 60% de falhas para abrir

### Estados
- **Closed:** Funcionamento normal
- **Open:** Bloqueando requisições (falha rápida)
- **Half-Open:** Testando se serviço recuperou

### Monitoramento
```bash
# Via API
curl http://localhost:9002/api/v1/circuit-breaker/status

# Via logs
docker logs consulta-service | grep "circuit breaker"

# Via Kibana
# Filtro: circuit_breaker_state:open
```

## 🔧 Configuração

### Variáveis de Ambiente
```bash
# Consulta Service
GIN_MODE=release
LOG_LEVEL=info
LOGSTASH_HOST=logstash:5000

# ELK Stack
ES_JAVA_OPTS=-Xms512m -Xmx512m
LS_JAVA_OPTS=-Xmx256m -Xms256m
```

### Portas Utilizadas
- **9002:** Consulta Service
- **9200:** Elasticsearch
- **5601:** Kibana
- **5044:** Logstash (Beats)
- **5000:** Logstash (TCP)
- **9600:** Logstash (API)

## 🔄 Integração com Módulos

### Módulo 1 (Auth/Admin)
- Usa autenticação via Keycloak
- Compartilha rede Docker

### Módulo 2 (Gateway)
- Roteamento via Kong Gateway
- Métricas no Prometheus
- Traces no Jaeger

### Próximos Módulos
- **Módulo 4:** IA Jurídica consumirá as consultas
- **Módulo 5:** Eventos serão gerados para mensageria
- **Módulo 6:** Painel web mostrará métricas