# 🔌 API Reference - Direito Lux

## 📊 Status da API

**🌐 Base URL (DEV):** `http://104.154.62.30`  
**📋 Health Check:** `GET /health`  
**🔐 Authentication:** JWT Bearer Token (em implementação)  
**📦 Content-Type:** `application/json`  

## 🏥 Health & Status

### `GET /health`
Verifica status da aplicação e dependências.

**Response:**
```json
{
  "status": "healthy",
  "mode": "full",
  "time": 1749687881
}
```

**Status Codes:**
- `200` - Aplicação saudável
- `503` - Aplicação com problemas

**Modes:**
- `full` - Aplicação completa com banco
- `demo` - Modo demonstração sem dependências

## 🔐 Authentication (em implementação)

### `POST /api/v1/auth/login`
Autentica usuário e retorna JWT token.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "tenant_id": "uuid-tenant"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "tenant_id": "uuid-tenant",
    "roles": ["user"]
  }
}
```

### `POST /api/v1/auth/refresh`
Renova token JWT usando refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### `POST /api/v1/auth/forgot-password`
Inicia processo de recuperação de senha.

**Request:**
```json
{
  "email": "user@example.com"
}
```

## 🏢 Tenant Management (Admin Only)

### `POST /api/v1/tenants`
Cria novo tenant (escritório/empresa).

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request:**
```json
{
  "name": "Escritório Silva & Associados",
  "email": "contato@silva-advogados.com.br",
  "phone": "+5511999999999",
  "document": "12.345.678/0001-90",
  "plan_id": "uuid-plan",
  "settings": {
    "timezone": "America/Sao_Paulo",
    "language": "pt-BR",
    "features": {
      "ai_enabled": true,
      "billing_enabled": true
    }
  }
}
```

**Response:**
```json
{
  "id": "uuid-tenant",
  "name": "Escritório Silva & Associados",
  "email": "contato@silva-advogados.com.br",
  "status": "active",
  "plan": {
    "id": "uuid-plan",
    "name": "professional",
    "limits": {
      "max_users": 5,
      "max_clients": 500,
      "max_cases": 1000
    }
  },
  "created_at": "2024-12-12T00:00:00Z",
  "updated_at": "2024-12-12T00:00:00Z"
}
```

### `GET /api/v1/tenants`
Lista todos os tenants (paginado).

**Query Parameters:**
- `page` (int): Página (default: 1)
- `limit` (int): Itens por página (default: 20, max: 100)
- `status` (string): Filtrar por status (`active`, `suspended`, `inactive`)
- `plan` (string): Filtrar por plano

**Response:**
```json
{
  "data": [
    {
      "id": "uuid-tenant",
      "name": "Escritório Silva & Associados",
      "email": "contato@silva-advogados.com.br",
      "status": "active",
      "plan_name": "professional",
      "users_count": 3,
      "created_at": "2024-12-12T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```

### `GET /api/v1/tenants/{id}`
Obtém detalhes de um tenant específico.

**Response:**
```json
{
  "id": "uuid-tenant",
  "name": "Escritório Silva & Associados",
  "email": "contato@silva-advogados.com.br",
  "phone": "+5511999999999",
  "document": "12.345.678/0001-90",
  "status": "active",
  "plan": {
    "id": "uuid-plan",
    "name": "professional",
    "display_name": "Professional",
    "price": 299.90,
    "currency": "BRL",
    "limits": {
      "max_users": 5,
      "max_clients": 500,
      "max_cases": 1000,
      "max_storage_gb": 50,
      "ai_requests_month": 1000
    }
  },
  "usage": {
    "users": 3,
    "clients": 127,
    "cases": 45,
    "storage_gb": 12.5,
    "ai_requests_month": 234
  },
  "settings": {
    "timezone": "America/Sao_Paulo",
    "language": "pt-BR",
    "features": {
      "ai_enabled": true,
      "billing_enabled": true
    }
  },
  "created_at": "2024-12-12T00:00:00Z",
  "updated_at": "2024-12-12T00:00:00Z"
}
```

### `PUT /api/v1/tenants/{id}`
Atualiza informações do tenant.

**Request:**
```json
{
  "name": "Escritório Silva, Santos & Associados",
  "phone": "+5511888888888",
  "plan_id": "uuid-new-plan",
  "settings": {
    "features": {
      "ai_enabled": false
    }
  }
}
```

### `GET /api/v1/tenants/{id}/usage`
Obtém estatísticas de uso detalhadas do tenant.

**Response:**
```json
{
  "tenant_id": "uuid-tenant",
  "period": "2024-12",
  "usage": {
    "users": {
      "current": 3,
      "limit": 5,
      "percentage": 60.0
    },
    "clients": {
      "current": 127,
      "limit": 500,
      "percentage": 25.4
    },
    "cases": {
      "current": 45,
      "limit": 1000,
      "percentage": 4.5
    },
    "storage": {
      "current_gb": 12.5,
      "limit_gb": 50,
      "percentage": 25.0
    },
    "ai_requests": {
      "current_month": 234,
      "limit_month": 1000,
      "percentage": 23.4,
      "daily_average": 7.8
    },
    "api_calls": {
      "current_month": 1567,
      "limit_month": 10000,
      "percentage": 15.67
    }
  },
  "billing": {
    "current_amount": 299.90,
    "next_billing_date": "2025-01-12",
    "status": "active"
  }
}
```

## 👤 User Profile

### `GET /api/v1/profile`
Obtém perfil do usuário autenticado.

**Response:**
```json
{
  "id": "uuid-user",
  "email": "user@silva-advogados.com.br",
  "name": "João Silva",
  "avatar_url": "https://example.com/avatar.jpg",
  "tenant": {
    "id": "uuid-tenant",
    "name": "Escritório Silva & Associados"
  },
  "roles": ["lawyer", "admin"],
  "permissions": [
    "tenants.read",
    "cases.write",
    "clients.write"
  ],
  "preferences": {
    "language": "pt-BR",
    "timezone": "America/Sao_Paulo",
    "theme": "light"
  },
  "last_login": "2024-12-12T10:30:00Z",
  "created_at": "2024-11-01T00:00:00Z"
}
```

### `PUT /api/v1/profile`
Atualiza perfil do usuário.

**Request:**
```json
{
  "name": "João da Silva",
  "avatar_url": "https://example.com/new-avatar.jpg",
  "preferences": {
    "theme": "dark",
    "language": "en-US"
  }
}
```

## 📋 Plans & Subscriptions

### Available Plans (Seeded Data)
Planos criados automaticamente via migration:

```json
[
  {
    "id": "uuid-starter",
    "name": "starter",
    "display_name": "Starter",
    "description": "Ideal para advogados autônomos",
    "price": 99.90,
    "currency": "BRL",
    "billing_cycle": "monthly",
    "features": {
      "basic_features": true,
      "email_support": true
    },
    "limits": {
      "max_users": 1,
      "max_clients": 50,
      "max_cases": 100,
      "max_storage_gb": 10,
      "max_api_calls_month": 1000,
      "ai_requests_month": 100,
      "messages_month": 500
    }
  },
  {
    "id": "uuid-professional",
    "name": "professional",
    "display_name": "Professional",
    "description": "Para pequenos escritórios",
    "price": 299.90,
    "currency": "BRL",
    "billing_cycle": "monthly",
    "features": {
      "all_features": true,
      "priority_support": true,
      "api_access": true
    },
    "limits": {
      "max_users": 5,
      "max_clients": 500,
      "max_cases": 1000,
      "max_storage_gb": 50,
      "max_api_calls_month": 10000,
      "ai_requests_month": 1000,
      "messages_month": 5000,
      "allow_custom_domain": true
    }
  },
  {
    "id": "uuid-enterprise",
    "name": "enterprise",
    "display_name": "Enterprise",
    "description": "Para grandes escritórios",
    "price": 999.90,
    "currency": "BRL",
    "billing_cycle": "monthly",
    "features": {
      "all_features": true,
      "dedicated_support": true,
      "api_access": true,
      "white_label": true
    },
    "limits": {
      "max_users": -1,
      "max_clients": -1,
      "max_cases": -1,
      "max_storage_gb": 500,
      "max_api_calls_month": -1,
      "ai_requests_month": 10000,
      "messages_month": -1,
      "allow_custom_domain": true
    }
  }
]
```

## 🚫 Error Responses

### Standard Error Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Dados inválidos fornecidos",
    "details": [
      {
        "field": "email",
        "message": "Email é obrigatório"
      },
      {
        "field": "password",
        "message": "Senha deve ter no mínimo 8 caracteres"
      }
    ],
    "request_id": "uuid-request",
    "timestamp": "2024-12-12T10:30:00Z"
  }
}
```

### Common Error Codes
| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Dados de entrada inválidos |
| `UNAUTHORIZED` | 401 | Token ausente ou inválido |
| `FORBIDDEN` | 403 | Sem permissão para esta ação |
| `NOT_FOUND` | 404 | Recurso não encontrado |
| `CONFLICT` | 409 | Conflito (ex: email já existe) |
| `RATE_LIMITED` | 429 | Muitas requisições |
| `INTERNAL_ERROR` | 500 | Erro interno do servidor |
| `SERVICE_UNAVAILABLE` | 503 | Serviço temporariamente indisponível |

## 🔧 Development & Testing

### Environment Variables
```bash
# Base URL por ambiente
DEV_URL=http://104.154.62.30
STAGING_URL=https://homolog.direito-lux.com.br
PROD_URL=https://app.direito-lux.com.br
```

### cURL Examples
```bash
# Health check
curl http://104.154.62.30/health

# Login (quando implementado)
curl -X POST http://104.154.62.30/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'

# Get tenants with auth
curl http://104.154.62.30/api/v1/tenants \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Rate Limiting
| Endpoint | Rate Limit | Window |
|----------|------------|--------|
| `/health` | 100 req/min | 1 minute |
| `/api/v1/auth/*` | 10 req/min | 1 minute |
| `/api/v1/*` | 1000 req/hour | 1 hour |

### Pagination
Endpoints que retornam listas suportam paginação:

**Query Parameters:**
- `page` (int): Número da página (começando em 1)
- `limit` (int): Itens por página (max: 100)
- `sort` (string): Campo para ordenação
- `order` (string): `asc` ou `desc`

**Response Headers:**
```
X-Total-Count: 150
X-Page: 1
X-Per-Page: 20
X-Total-Pages: 8
Link: <http://api/endpoint?page=2>; rel="next"
```

## 📚 Additional Resources

- **OpenAPI Spec:** `/docs/swagger.yaml` (em desenvolvimento)
- **Postman Collection:** `/postman/Direito-Lux-API.postman_collection.json`
- **Environment Setup:** `/docs/CONFIGURACAO-AMBIENTE.md`
- **Database Schema:** `/docs/MIGRATIONS-E-PERSISTENCIA.md`

---

**🔌 API REST completa e documentada para o Direito Lux!**

*Última atualização: 12 de Junho de 2024*