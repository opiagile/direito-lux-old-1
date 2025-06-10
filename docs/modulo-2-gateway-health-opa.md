# Módulo 2: API Gateway, Health, OPA

## Visão Geral

Este módulo implementa o API Gateway com Kong, health checks avançados, autorização com Open Policy Agent (OPA), e observabilidade completa para o Direito Lux.

## Arquitetura

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Client App    │────▶│   Kong Gateway  │────▶│  Direito Lux    │
│  (Frontend)     │     │  (Rate Limit)   │     │     API Go      │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │                          │
                               ▼                          ▼
                        ┌─────────────────┐     ┌─────────────────┐
                        │      OPA        │     │   Health Check  │
                        │ (Authorization) │     │   Endpoints     │
                        └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐     ┌─────────────────┐
                        │   Prometheus    │────▶│     Grafana     │
                        │   (Metrics)     │     │  (Dashboards)   │
                        └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │     Jaeger      │
                        │   (Tracing)     │
                        └─────────────────┘
```

## Componentes Implementados

### 1. **Kong API Gateway**
- **Rate Limiting**: Por tenant, baseado no plano
- **JWT Validation**: Integrado com Keycloak
- **CORS**: Configurado para todas as rotas
- **Request/Response Transformation**: Headers de segurança
- **Health Checks**: Monitoramento de upstreams
- **Load Balancing**: Round-robin entre instâncias

### 2. **Open Policy Agent (OPA)**
- **Autorização Granular**: RBAC + ABAC
- **Multi-tenancy**: Isolamento rigoroso
- **Políticas Complexas**: 
  - Acesso por role (admin, lawyer, secretary, client)
  - Restrições por IP
  - Janelas de tempo
  - Limites por plano
  - Compliance (GDPR)

### 3. **Health Checks Avançados**
- **/health**: Status geral do sistema
- **/health/live**: Liveness probe (Kubernetes)
- **/health/ready**: Readiness probe
- **Componentes monitorados**:
  - Database (PostgreSQL)
  - Cache (Redis)
  - Auth (Keycloak)
  - Gateway (Kong)
  - Authorization (OPA)

### 4. **Circuit Breaker**
- **Estados**: Closed, Open, Half-Open
- **Configurável por serviço**
- **Proteção contra cascading failures**
- **Métricas de circuit breaker**

### 5. **Observabilidade**
- **Prometheus**: Métricas de sistema e aplicação
- **Grafana**: Dashboards customizados
- **Jaeger**: Distributed tracing
- **Logs estruturados**: Com correlation ID

## Configuração

### Iniciar o Módulo 2

```bash
# Criar rede se não existir
docker network create direito-lux-network

# Iniciar serviços do gateway
docker compose -f docker-compose.gateway.yml up -d

# Verificar status
docker compose -f docker-compose.gateway.yml ps
```

### URLs de Acesso

- **Kong Gateway**: http://localhost:8000
- **Kong Admin**: http://localhost:8001
- **OPA API**: http://localhost:8181
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger UI**: http://localhost:16686

## Políticas OPA

### Estrutura da Política

```rego
package direitolux.authz

# Regras por Role
- admin: Acesso total dentro do tenant
- lawyer: CRUD em cases, clients, documents
- secretary: Read/Create/Update (sem delete)
- client: Apenas leitura dos próprios dados

# Multi-tenancy
- Isolamento rigoroso por tenant_id
- Violações são logadas e bloqueadas

# Rate Limiting
- Baseado no plano do tenant
- Starter: 100 req/min
- Professional: 1000 req/min
- Enterprise: Ilimitado

# Compliance
- GDPR: Direito de acesso/exclusão
- Audit trail obrigatório
- Anonimização de dados sensíveis
```

### Exemplo de Autorização

```json
// Request para OPA
{
  "input": {
    "user": {
      "id": "user-123",
      "tenant_id": "tenant-abc",
      "role": "lawyer",
      "tenant_plan": "professional"
    },
    "resource": {
      "type": "case",
      "id": "case-456",
      "tenant_id": "tenant-abc"
    },
    "action": "update",
    "method": "PUT"
  }
}

// Response
{
  "result": {
    "allow": true,
    "audit_required": true,
    "requires_anonymization": false
  }
}
```

## Health Check Response

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-01-15T10:30:00Z",
  "total_duration_ms": 45,
  "checks": [
    {
      "name": "database",
      "status": "healthy",
      "duration_ms": 12,
      "details": {
        "open_connections": 5,
        "in_use": 2
      }
    },
    {
      "name": "redis",
      "status": "healthy",
      "duration_ms": 8,
      "details": {
        "total_conns": 10,
        "idle_conns": 8
      }
    },
    {
      "name": "keycloak",
      "status": "healthy",
      "duration_ms": 25
    }
  ]
}
```

## Circuit Breaker

### Configuração

```go
cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Settings{
    Name:        "external-api",
    MaxRequests: 3,              // Requests permitidas em half-open
    Interval:    10 * time.Second, // Intervalo para closed state
    Timeout:     30 * time.Second, // Timeout para open state
    ReadyToTrip: func(counts Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 3 && failureRatio >= 0.6
    },
})
```

### Uso

```go
result, err := cb.Execute(ctx, func() (interface{}, error) {
    // Chamada para serviço externo
    return externalAPI.Call()
})

if err != nil {
    if _, ok := err.(circuitbreaker.ErrOpenState); ok {
        // Circuit breaker está aberto
        return fallbackResponse()
    }
}
```

## Métricas Disponíveis

### Kong Metrics
- `kong_http_requests_total`: Total de requisições
- `kong_http_latency_ms`: Latência das requisições
- `kong_bandwidth_bytes`: Bandwidth consumido
- `kong_upstream_target_health`: Saúde dos upstreams

### Application Metrics
- `direitolux_api_requests_total`: Requisições por endpoint
- `direitolux_api_duration_seconds`: Duração das requisições
- `direitolux_tenant_usage`: Uso por tenant
- `direitolux_circuit_breaker_state`: Estado dos circuit breakers

### OPA Metrics
- `opa_decisions_total`: Total de decisões
- `opa_decision_duration_seconds`: Tempo de decisão
- `opa_policies_loaded`: Políticas carregadas

## Dashboards Grafana

### Dashboard Principal
- Request rate por serviço
- Latência P95
- Distribuição de status HTTP
- Health status dos serviços

### Dashboard de Tenant
- Uso por tenant
- Rate limit status
- Top tenants por uso
- Violações de policy

### Dashboard de Circuit Breaker
- Estado dos circuit breakers
- Taxa de falha por serviço
- Tempo em open state
- Requests rejeitadas

## Integração com o Módulo 1

O módulo 2 se integra perfeitamente com o módulo 1:

1. **Autenticação**: Kong valida JWT do Keycloak
2. **Autorização**: OPA usa claims do JWT + dados do tenant
3. **Rate Limiting**: Baseado no plano do tenant (do banco)
4. **Audit**: Todas as ações são logadas com contexto completo

## Troubleshooting

### Kong não conecta na API
```bash
# Verificar rede
docker network inspect direito-lux-network

# Verificar DNS interno
docker exec kong-gateway nslookup direito-lux-api
```

### OPA não autoriza requests
```bash
# Testar política diretamente
curl -X POST http://localhost:8181/v1/data/direitolux/authz \
  -H "Content-Type: application/json" \
  -d '{"input": {...}}'

# Ver logs de decisão
docker logs opa-server -f
```

### Métricas não aparecem no Grafana
```bash
# Verificar targets no Prometheus
curl http://localhost:9090/api/v1/targets

# Verificar datasource no Grafana
curl http://admin:admin@localhost:3000/api/datasources
```

## Próximos Passos

- [ ] Implementar cache de decisões OPA
- [ ] Adicionar webhook para eventos do gateway
- [ ] Configurar alertas no Prometheus
- [ ] Implementar distributed rate limiting
- [ ] Adicionar suporte a WebSockets no Kong