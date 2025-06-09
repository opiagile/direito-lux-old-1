# Módulo 1: Núcleo Auth/Admin Go + Keycloak

## Visão Geral

Este módulo implementa o núcleo de autenticação e administração do Direito Lux, integrando Go com Keycloak para gerenciamento de identidade multi-tenant.

## Arquitetura

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Frontend      │────▶│   Go Backend    │────▶│    Keycloak     │
│  (React/Vue)    │     │   (Gin/GORM)    │     │  (Auth Server)  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │                          │
                               ▼                          ▼
                        ┌─────────────────┐     ┌─────────────────┐
                        │   PostgreSQL    │     │   Redis Cache   │
                        │   (App Data)     │     │  (JWT/Sessions) │
                        └─────────────────┘     └─────────────────┘
```

## Estrutura do Projeto

```
direito-lux/
├── cmd/
│   └── api/
│       └── main.go              # Entry point da aplicação
├── internal/
│   ├── auth/
│   │   └── keycloak_client.go   # Cliente Keycloak Admin
│   ├── config/
│   │   └── config.go            # Configurações da aplicação
│   ├── domain/
│   │   └── models.go            # Modelos de domínio (Tenant, User, etc)
│   ├── handlers/
│   │   ├── auth_handler.go      # Endpoints de autenticação
│   │   └── tenant_handler.go    # Endpoints de tenant
│   ├── middleware/
│   │   ├── auth.go              # Middleware JWT
│   │   ├── cors.go              # CORS
│   │   ├── logger.go            # Logging
│   │   ├── recovery.go          # Panic recovery
│   │   └── request_id.go        # Request ID tracking
│   ├── repository/
│   │   └── repositories.go      # Camada de dados
│   └── services/
│       └── tenant_service.go    # Lógica de negócios
└── pkg/
    ├── logger/
    │   └── logger.go            # Logger centralizado
    └── utils/                   # Utilitários compartilhados
```

## Principais Funcionalidades

### 1. Multi-tenancy
- Cada tenant (escritório/profissional) tem isolamento completo de dados
- Grupos no Keycloak para segregação de usuários
- Tenant ID extraído do JWT token

### 2. Autenticação JWT
- Validação de tokens com Keycloak
- Cache de tokens no Redis
- Blacklist de tokens revogados
- Refresh token automático

### 3. RBAC (Role-Based Access Control)
- Roles: admin, lawyer, secretary, client
- Middleware para verificação de roles
- Permissões granulares por tenant

### 4. Gerenciamento de Tenants
- Criação de novos tenants com admin
- Planos: Starter, Professional, Enterprise
- Limites de uso por plano
- Tracking de uso (API calls, AI requests, etc)

## Configuração

1. Copie o arquivo de configuração de exemplo:
```bash
cp config.yaml.example config.yaml
```

2. Atualize o `config.yaml` com suas configurações:
```yaml
keycloak:
  clientSecret: "PBaNyvNLoJSCmW0VaxuQtE4VZGJhqPtF"  # Do seu Keycloak
```

3. Instale as dependências:
```bash
go mod download
```

4. Execute as migrações (automáticas ao iniciar):
```bash
go run cmd/api/main.go
```

## Endpoints da API

### Autenticação
- `POST /api/v1/auth/login` - Login de usuário
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/forgot-password` - Reset de senha

### Tenants (Admin only)
- `POST /api/v1/tenants` - Criar novo tenant
- `GET /api/v1/tenants` - Listar tenants
- `GET /api/v1/tenants/:id` - Detalhes do tenant
- `PUT /api/v1/tenants/:id` - Atualizar tenant
- `GET /api/v1/tenants/:id/usage` - Uso do tenant

### Profile
- `GET /api/v1/profile` - Perfil do usuário
- `PUT /api/v1/profile` - Atualizar perfil

## Exemplo de Criação de Tenant

```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "escritorio-silva",
    "display_name": "Escritório Silva & Associados",
    "plan_id": "uuid-do-plano",
    "admin_user": {
      "email": "admin@escritoriosilva.com",
      "first_name": "João",
      "last_name": "Silva",
      "password": "senha-segura-123"
    },
    "settings": {
      "language": "pt-BR",
      "timezone": "America/Sao_Paulo",
      "currency_code": "BRL"
    }
  }'
```

## Segurança

1. **JWT Validation**: Todos os tokens são validados com Keycloak
2. **Rate Limiting**: Implementado por tenant (TODO)
3. **Audit Logging**: Todas as ações são registradas
4. **Data Isolation**: Queries sempre filtradas por tenant_id
5. **Password Policy**: Configurada no Keycloak

## Próximos Passos

- [ ] Implementar rate limiting por tenant
- [ ] Adicionar métricas com Prometheus
- [ ] Implementar webhooks para eventos
- [ ] Adicionar testes de integração
- [ ] Documentação OpenAPI/Swagger

## Desenvolvimento

Para executar em modo desenvolvimento:
```bash
go run cmd/api/main.go
```

Para build de produção:
```bash
go build -o bin/api cmd/api/main.go
```

## Troubleshooting

1. **Erro de conexão com Keycloak**: Verifique se o Keycloak está rodando e o realm está configurado
2. **Erro de conexão com PostgreSQL**: Verifique as credenciais no config.yaml
3. **Token inválido**: Verifique se o client secret está correto