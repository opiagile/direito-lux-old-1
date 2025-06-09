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

| Módulo | Novas Adições                          | Ferramentas/Exemplos                    |
|--------|----------------------------------------|------------------------------------------|
| 0      | Setup CI/CD, Keycloak HA, Vault        | GitHub Actions, ArgoCD, Docker Compose  |
| 1      | Núcleo Auth/Admin Go + Keycloak        | keycloak-admin-go, Redis, PostgreSQL     |
| 2      | API Gateway, Health, OPA               | Go, Open Policy Agent                    |
| 3      | Consulta Jurídica + Circuit Breaker    | Go, Hystrix, ELK, OpenTelemetry          |
| 4      | IA Jurídica (RAG + Avaliação)          | Python, LangChain, Vertex AI, Ragas      |
| 5      | Mensageria e Eventos                   | Go, Kafka, Avro, DLQ                     |
| 6      | Painel Admin Web (React/Vue.js)        | React, Keycloak JS Adapter               |
| 7      | Billing e Relatórios                   | Go, Stripe SDK, BigQuery                 |

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
