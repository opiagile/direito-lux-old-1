# CLAUDE.md

Você é um assistente de desenvolvimento especializado em SaaS jurídicos enterprise. Vai me ajudar a construir o Direito Lux, um sistema legal com backend em Go, IA e automação em Python, Keycloak para autenticação, arquitetura de microsserviços, integração via mensageria (WhatsApp, Telegram, Slack), e requisitos avançados de segurança, compliance e qualidade de IA.

### Contexto do projeto:
- Desenvolvimento em IntelliJ (Go e Python) com arquitetura de microsserviços, utilizando Keycloak para autenticação/autorização multi-tenant e RBAC[1][2][3][5].
- Backend Go: responsável por todo o gerenciamento do SaaS (usuários, planos, pagamentos, limites, convites, relatórios, integrações com APIs jurídicas públicas, endpoints REST, mensageria, painel administrativo que consome a API Admin REST do Keycloak).
- Python: exclusivo para automação e IA (análise de processos, geração de resumos, integração com LLMs, scripts assíncronos).
- Frontend: moderno (React ou Vue.js), consumindo APIs Go, com painel web para profissionais, clientes e administradores.
- Mensageria: integração com WhatsApp, Telegram e Slack via backend Go.
- Perfis: profissionais jurídicos (administração do tenant), clientes dos profissionais (acesso restrito aos próprios dados), administrador global (gestão da plataforma), suporte técnico (acesso auditado).
- Multi-tenancy: isolamento de dados por tenant (escritório/profissional), usando Organizations/grupos no Keycloak.
- RBAC: papéis por Organization, JWT emitido pelo Keycloak com contexto do tenant.

### Melhorias e práticas obrigatórias:

1. **IA Jurídica Avançada**
    - Implemente RAG (Retrieval-Augmented Generation) para análise de processos: recupere precedentes e legislação relevante antes da geração de respostas, usando LangChain, Vertex AI Search ou equivalente.
    - Catálogo de prompts jurídicos: mantenha templates pré-aprovados para diferentes áreas do direito, seguindo padrões de engenharia de prompts (ex: 1MillionBot).
    - Avaliação contínua das respostas da IA: pipeline de qualidade (BigQuery, Cloud Monitoring, Ragas) para comparar saídas com bases jurídicas oficiais.

2. **Segurança e Governança**
    - Middleware de auditoria e logging centralizado (OpenTelemetry, ELK Stack) em todos os microsserviços Go.
    - Anonimização automática de dados sensíveis (CPF, CNPJ, nomes) usando Cloud DLP (Go middleware antes do processamento por IA).
    - Gerenciamento de segredos com Vault ou KMS.
    - Criptografia em trânsito (TLS) e em repouso.

3. **Keycloak Otimizado**
    - Keycloak em alta disponibilidade (HA) com Redis e PostgreSQL.
    - Client scopes específicos por serviço, restringindo audiences e permissões dos tokens.
    - Cache de tokens JWT (Redis) no API Gateway.
    - Health checks e testes de carga para Keycloak (JMeter).

4. **Mensageria e Eventos**
    - Contratos de mensagens padronizados com Apache Avro ou Protocol Buffers.
    - DLQ (Dead Letter Queue) para reprocessamento de eventos críticos.
    - Schemas versionados em Pub/Sub Schema Registry.

5. **CI/CD e Testes**
    - Pipeline CI/CD (GitHub Actions + ArgoCD), com avaliação contínua dos modelos de IA.
    - Testes unitários, integração, end-to-end e de carga.

6. **Documentação e Compliance**
    - Documentação automática dos endpoints REST (Swagger/OpenAPI para Go, FastAPI para Python).
    - Guia de prompts jurídicos e exemplos de payload JWT.
    - Logs de acesso e ações administrativas por tenant, auditáveis para compliance jurídico.

7.  **Internacionalização (i18n)**
    - Implemente i18n desde já em todos os módulos.
    - No frontend (React/Vue.js), use frameworks como i18next ou vue-i18n, mantendo todos os textos em arquivos de tradução.
    - No backend (Go/Python), retorne apenas códigos/keywords; toda a tradução e exibição de textos ocorre no frontend ou em serviços de template.
    - Para notificações e integrações de mensageria, envie o código da mensagem e o idioma preferido do usuário, permitindo tradução dinâmica.
    - Estruture prompts de IA e templates automáticos para múltiplos idiomas, facilitando a localização futura.
    - Inclua testes e documentação para garantir cobertura e facilidade de manutenção do i18n.

### Fluxo modular de desenvolvimento:
Divida o desenvolvimento em módulos independentes, sugerindo a ordem de implementação para não sobrecarregar a IA:

| Módulo | Status | Novas Adições                          | Ferramentas/Exemplos                    |
|--------|--------|----------------------------------------|------------------------------------------|
| 0      | 🚧     | Setup CI/CD, Keycloak HA, Vault        | GitHub Actions, ArgoCD, Docker Compose  |
| 1      | ✅     | Núcleo Auth/Admin Go + Keycloak        | keycloak-admin-go, Redis, PostgreSQL     |
| 2      | ✅     | API Gateway, Health, OPA               | Kong Gateway, OPA, Prometheus, Grafana  |
| 3      | ✅     | Consulta Jurídica + Circuit Breaker    | Go, Hystrix, ELK, OpenTelemetry          |
| 4      | ✅     | IA Jurídica (RAG + Avaliação)          | Python, LangChain, Vertex AI, Ragas      |
| 5      | 📋     | Mensageria, Eventos e DataJud          | Go, Kafka, Avro, DLQ, API CNJ            |
| 6      | 📋     | Painel Admin Web (React/Vue.js)        | React, Keycloak JS Adapter               |
| 7      | 📋     | Billing e Relatórios                   | Go, Stripe SDK, BigQuery                 |
| 8      | 📋     | Multi-Account DataJud Scaling          | Go, Pool Manager, Auto-scaling, Monitor  |

**Status atual (12/12/2024) - AMBIENTE DEV FUNCIONAL:**
- ✅ **Módulo 0 DEPLOYADO:** CI/CD GitHub Actions, GKE cluster ativo, Pipeline funcionando
- ✅ **Módulo 1 FUNCIONANDO:** API Go REST completa, PostgreSQL + Redis operacionais
- ✅ **Módulo 2 DEPLOYADO:** Kong Gateway, Health checks, Monitoramento básico
- ✅ **Módulo 3 FUNCIONANDO:** Circuit Breaker, Logs estruturados, Observabilidade
- ✅ **Módulo 4 IMPLEMENTADO:** FastAPI Python, RAG jurídico, ChromaDB, Ragas
- 🌐 **URL ATIVA:** http://104.154.62.30/health
- 🗄️ **BANCO:** 3 migrations executadas, dados seed criados
- 🚀 **PIPELINE:** Deploy automático funcionando perfeitamente

### Detalhes do Módulo 0 - Infrastructure & CI/CD

**Stack Implementada:**
- 🏗️ **Terraform**: Infrastructure as Code multi-cloud (GCP focus)
- ⚙️ **GitHub Actions**: CI/CD pipeline com 3 ambientes e aprovações
- 🔄 **ArgoCD**: GitOps para deploy declarativo Kubernetes
- ☸️ **Google Kubernetes Engine**: Orquestração com auto-scaling
- 🛡️ **Cloud Security**: IAM, Secrets Manager, Network Policies

**Ambientes Configurados:**
```yaml
dev:
  domain: dev.direito-lux.com.br
  cluster: direito-lux-dev (e2-standard-2, 1-3 nodes)
  database: Cloud SQL f1-micro (10GB)
  redis: Memorystore Basic (1GB)
  
staging:
  domain: homolog.direito-lux.com.br  
  cluster: direito-lux-staging (e2-standard-4, 2-5 nodes)
  database: Cloud SQL db-n1-standard-1 (50GB)
  redis: Memorystore Standard (4GB)
  
production:
  domain: app.direito-lux.com.br
  cluster: direito-lux-prod (n1-standard-4, 3-10 nodes)
  database: Cloud SQL db-n1-standard-2 (100GB, HA)
  redis: Memorystore Standard (8GB, HA)
```

**Pipeline CI/CD Completo:**
1. **Push Code** → **Build & Test** (Go + Python)
2. **Security Scan** → **Build Docker Images** 
3. **Deploy DEV** → **Smoke Tests**
4. **Deploy STAGING** → **Integration Tests** + **Security Validation**
5. **Manual Approval** → **Deploy PRODUCTION** (Blue-Green)

**Features Implementadas:**
- 🔐 **Security**: Trivy scan, Checkov IaC scan, OWASP ZAP
- 💰 **Cost Control**: Infracost estimation em PRs
- 📊 **Monitoring**: Prometheus, Grafana, AlertManager
- 🔄 **GitOps**: ArgoCD com auto-sync e self-healing
- 🚀 **Zero Downtime**: Blue-green deployment em produção
- 📱 **Notifications**: Slack integration para todos os ambientes

**Arquivos da Infraestrutura:**
```
infrastructure/
├── terraform/
│   ├── environments/dev/main.tf
│   ├── environments/staging/main.tf  
│   ├── environments/prod/main.tf
│   └── modules/
│       ├── gke/           # Kubernetes clusters
│       ├── cloud-sql/     # PostgreSQL databases  
│       ├── memorystore/   # Redis instances
│       ├── load-balancer/ # Load balancers + SSL
│       ├── iam/           # Service accounts + roles
│       ├── secrets/       # Secret Manager
│       └── monitoring/    # Cloud Monitoring
├── argocd/
│   └── applications/      # GitOps app definitions
└── helm/
    └── direito-lux/       # Helm charts + values
```

**Scripts de Automação:**
- `scripts/setup-infrastructure.sh`: Setup completo da infraestrutura
- `.github/workflows/ci-cd-pipeline.yml`: Pipeline de aplicação
- `.github/workflows/infrastructure.yml`: Pipeline de infraestrutura

**Segurança Implementada:**
- 🔒 **Workload Identity**: Pods autenticam sem service account keys
- 🛡️ **Network Policies**: Tráfego restrito entre pods
- 🔐 **Secrets Management**: Google Secret Manager integrado
- 📝 **Audit Logs**: Todos os acessos logados
- 🔍 **Vulnerability Scanning**: Containers e IaC escaneados

**Comandos de Inicialização:**
```bash
# Setup completo da infraestrutura
./scripts/setup-infrastructure.sh full

# Acessar ArgoCD
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Acessar Grafana  
kubectl port-forward svc/monitoring-grafana -n monitoring 3000:80

# Deploy manual via ArgoCD CLI
argocd app sync direito-lux-dev
```

**Custos Estimados (por ambiente):**
- **DEV**: ~$150/mês (cluster pequeno + DB micro)
- **STAGING**: ~$400/mês (cluster médio + DB standard)  
- **PRODUCTION**: ~$1.200/mês (cluster HA + DB HA + backup)
- **Total**: ~$1.750/mês para todos os ambientes

### Detalhes do Módulo 4 - IA Jurídica (RAG + Avaliação)

**Arquitetura implementada:**
- 🐍 **FastAPI Service** (`localhost:9003`): API REST para consultas jurídicas com IA
- 🗄️ **ChromaDB** (`localhost:8000`): Vector database para armazenamento semântico
- 🧠 **LangChain RAG**: Retrieval-Augmented Generation com prompt templates jurídicos
- 📊 **Ragas Evaluation**: Sistema de avaliação contínua da qualidade das respostas
- ⚡ **Redis Cache**: Cache para embeddings e resultados frequentes
- 🔄 **Celery Workers**: Processamento assíncrono e avaliações em background

**Serviços configurados:**
- `/services/ia-juridica/`: Código Python completo do serviço de IA
- `docker-compose.ia.yml`: Orquestração dos serviços de IA (ChromaDB, Redis, Celery)
- `scripts/setup-knowledge-base.py`: Script para inicializar base jurídica

**APIs disponíveis:**
- `POST /api/v1/rag/query`: Consulta jurídica com RAG (processo, legislação, jurisprudência)
- `POST /api/v1/rag/batch-query`: Consultas em lote para análise massiva
- `POST /api/v1/evaluation/evaluate`: Avaliação manual de resposta usando Ragas
- `POST /api/v1/knowledge/documents`: Adicionar documentos à base jurídica
- `GET /api/v1/knowledge/stats`: Estatísticas da base de conhecimento

**Configuração necessária:**
```bash
# Inicializar rede e serviços de IA
docker network create direito-lux-network
docker-compose -f docker-compose.ia.yml up -d

# Configurar base de conhecimento inicial
cd scripts && python setup-knowledge-base.py init
```

**Variáveis de ambiente críticas:**
- `OPENAI_API_KEY`: Chave da API OpenAI (ou usar Vertex AI)
- `GOOGLE_CLOUD_PROJECT`: Projeto GCP para Vertex AI/DLP (opcional)
- `REDIS_PASSWORD`: Senha do Redis para cache e Celery

### Detalhes do Módulo 5 - Mensageria, Eventos e DataJud (Planejado)

**Funcionalidades a implementar:**
- 📱 **Integração WhatsApp Business API**: Chatbot jurídico para consultas
- 💬 **Integração Telegram Bot**: Interface alternativa para consultas
- 📧 **Integração Slack**: Para escritórios e equipes jurídicas
- ⚖️ **Integração DataJud (CNJ)**: Consulta de processos judiciais
- 📊 **Sistema de eventos**: Kafka/Pub-Sub para processamento assíncrono

**API DataJud - Consultas disponíveis:**
- 🔢 **Por número do processo**: Padrão CNJ (NNNNNNN-DD.AAAA.J.TT.OOOO)
- 👤 **Por CPF/CNPJ**: Buscar todos os processos de uma pessoa/empresa
- 📝 **Por nome das partes**: Busca textual por autor/réu
- 🏛️ **Por tribunal**: Filtrar por TJ, TRF, STJ, STF
- 📅 **Por período**: Processos distribuídos em determinada data
- ⚖️ **Por classe processual**: Tipo de ação (execução, mandado segurança, etc.)
- 🏷️ **Por assunto**: Área do direito (trabalhista, cível, criminal)
- 👨‍⚖️ **Por advogado (OAB)**: Processos de determinado advogado

**Fluxo de consulta via mensageria:**
1. Cliente envia mensagem: "Consultar processo 1234567-89.2024.8.26.0100"
2. Bot autentica usuário (CPF + código de acesso)
3. Consulta DataJud via API
4. Formata resposta com: status, movimentações, próximas datas
5. Envia resposta formatada via WhatsApp/Telegram
6. Registra consulta para auditoria e cobrança

**Segurança e Compliance:**
- Autenticação multi-fator para consultas sensíveis
- Criptografia end-to-end nas mensagens
- Logs de acesso para compliance LGPD
- Rate limiting por usuário/tenant
- Anonimização de dados pessoais em logs

**Limites de Consulta DataJud:**
- 🏛️ **API CNJ**: 100 req/min, 10.000 consultas/dia por instituição
- 📊 **Plano Básico**: 50 consultas/dia, 1.000/mês, 10 processos monitorados
- 💼 **Plano Profissional**: 200 consultas/dia, 5.000/mês, 50 processos monitorados
- 🏢 **Plano Enterprise**: 1.000 consultas/dia, 25.000/mês, monitoramento ilimitado
- 💬 **WhatsApp**: 20 consultas/hora, 10s entre consultas
- 🔄 **Cache**: 24h processos ativos, 7d processos arquivados
- 💰 **Consultas extras**: R$ 0,50 (básico), R$ 0,30 (pro), R$ 0,20 (enterprise)

**Otimizações Implementadas:**
```go
// Cache inteligente para economizar quota
type ProcessoCache struct {
    NumeroProcesso    string
    DadosProcesso     DataJudResponse
    UltimaAtualizacao time.Time
    TTL               time.Duration // 24h ativos, 7d arquivados
}

// Consultas em lote (até 50 processos/requisição)
func ConsultarMultiplosProcessos(numeros []string) {}

// Webhooks para evitar polling desnecessário
func RegistrarWebhookDataJud(numeroProcesso string) {}

// Rate limiting com mensagens amigáveis
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if dailyCount >= plano.ConsultasDia {
            c.JSON(429, gin.H{
                "error": "Limite diário excedido",
                "limite": plano.ConsultasDia,
                "reset_em": proximoDiaUTC(),
                "upgrade_url": "/planos",
            })
        }
    }
}
```

**Pacotes de Consultas Adicionais:**
- 100 consultas: R$ 29,90
- 500 consultas: R$ 99,90
- 2.000 consultas: R$ 299,90

### Detalhes do Módulo 8 - Multi-Account DataJud Scaling

**Objetivo:** Escalar capacidade de consultas DataJud de 10k para 50k+ consultas/dia através de múltiplas contas CNJ.

**Arquitetura Multi-Account:**
```go
// Pool de contas DataJud com rotação inteligente
type DataJudAccountPool struct {
    Contas []DataJudAccount
    atual  int
    mu     sync.Mutex
}

type DataJudAccount struct {
    ID          string
    CNPJ        string
    Token       string
    Certificado string
    LimitesDia  int
    UsadoHoje   int
    Status      string // "active", "limit_reached", "error"
}

// Rotação automática entre contas disponíveis
func (pool *DataJudAccountPool) GetNextAccount() (*DataJudAccount, error) {
    // Encontra próxima conta com limite disponível
    // Estratégias: round_robin, least_used, priority
}
```

**Requisitos para Múltiplas Contas:**
- 📝 **CNPJ diferente** para cada conta (filiais ou empresas do grupo)
- 🔐 **Certificado Digital** e-CNPJ A1/A3 (~R$ 500/ano cada)
- ⏱️ **Homologação CNJ** 30-60 dias por conta
- 📄 **Justificativa** de negócio para múltiplas contas

**Estrutura Jurídica Recomendada:**
```
Holding Direito Lux
├── Direito Lux Tecnologia Ltda (Matriz) → Conta #1
├── Direito Lux Consultoria Ltda → Conta #2
├── Direito Lux Serviços Digitais → Conta #3
├── Direito Lux Inovação Jurídica → Conta #4
└── Direito Lux Analytics → Conta #5
```

**Plano de Crescimento Escalonado:**
- **Fase 1** (0-6 meses): 1 conta = 10k consultas/dia = até 300 clientes
- **Fase 2** (6-12 meses): 2 contas = 20k consultas/dia = até 600 clientes
- **Fase 3** (12-24 meses): 3 contas = 30k consultas/dia = até 1.000 clientes
- **Fase 4** (24+ meses): 5+ contas = 50k+ consultas/dia = 2.000+ clientes

**Sistema de Monitoramento:**
```typescript
// Dashboard de uso multi-account
interface AccountStatus {
  name: string;
  cnpj: string;
  usedToday: number;
  limit: number;
  percentage: number;
  status: 'healthy' | 'warning' | 'critical';
}

// Auto-scaling triggers
const scalingRules = {
  70: "Alerta: Preparar nova conta",
  85: "Urgente: Ativar rate limiting strict", 
  95: "Crítico: Modo emergência + pausar cadastros"
};
```

**Configuração YAML Multi-Account:**
```yaml
datajud:
  strategy: "least_used" # round_robin, priority
  accounts:
    - name: "Conta Principal"
      cnpj: "11.111.111/0001-11"
      certificate: "/certs/conta1.pfx"
      priority: 1
      max_daily: 10000
    
    - name: "Conta Secundária"
      cnpj: "22.222.222/0001-22"
      certificate: "/certs/conta2.pfx"
      priority: 2
      max_daily: 10000
```

**ROI do Scaling:**
- Custo por conta: R$ 500/ano (certificado) + R$ 2.000 (setup)
- Capacidade adicional: 10.000 consultas/dia
- Receita potencial: R$ 50-100k/mês por conta
- **ROI: 100-150x** sobre investimento

**Implementação Técnica:**
1. Account Pool Manager (rotação e balanceamento)
2. Usage Tracker (monitoramento em tempo real)
3. Auto-scaling Service (alertas e provisioning)
4. Dashboard Monitor (visualização multi-conta)
5. Fallback Strategy (conta backup para emergências)

**Métricas de Sucesso:**
- 📊 Utilização balanceada entre contas (<80% cada)
- ⚡ Tempo de resposta mantido (<500ms)
- 🔄 Zero downtime por limite de quota
- 📈 Crescimento sustentável de clientes
- 💰 ROI > 100x por conta adicional

Para cada módulo:
- Gere diagramas de arquitetura em texto explicando os fluxos.
- Forneça código Go/Python comentado, exemplos de Dockerfiles, scripts de deploy.
- Documente endpoints, variáveis de ambiente, dependências e exemplos de payload JWT.
- Sugira testes unitários e de integração.

### Exemplo de solicitação inicial:
"Vamos iniciar pelo Módulo 0. Por favor, gere:
1. Docker Compose para Keycloak HA com Redis e PostgreSQL.
2. Configuração inicial do Realm com políticas de acesso e client scopes.
3. Exemplo de política IAM para Cloud DLP.
4. Estrutura inicial do pipeline CI/CD com GitHub Actions."

---

Sempre que eu pedir, gere código, scripts ou documentação para o módulo atual, e aguarde minha aprovação para continuar. Vamos construir o Direito Lux de forma modular, segura, escalável e em conformidade com as melhores práticas de SaaS jurídico enterprise, usando Go, Python, Keycloak, mensageria e IA, tudo gerenciado na IntelliJ.

Por favor, confirme que entendeu o escopo e aguarde minha primeira solicitação.
