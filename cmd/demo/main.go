package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opiagile/direito-lux/internal/auth"
	"github.com/opiagile/direito-lux/internal/config"
)

func main() {
	// Simple demo without database
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "Direito Lux API",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"features": []string{
				"Multi-tenant authentication",
				"Keycloak integration", 
				"JWT validation",
				"RBAC (Role-Based Access Control)",
				"Redis caching",
				"Audit logging",
			},
		})
	})

	// API info
	router.GET("/api/v1/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "Direito Lux - SaaS Jurídico Enterprise",
			"description": "Sistema jurídico com backend em Go, IA e automação",
			"version":     "1.0.0",
			"author":      "Opiagile",
			"architecture": gin.H{
				"backend":        "Go (Gin + GORM)",
				"authentication": "Keycloak",
				"database":       "PostgreSQL",
				"cache":          "Redis",
				"ai":             "Python (LangChain + Vertex AI)",
				"messaging":      "WhatsApp, Telegram, Slack",
			},
			"modules": []gin.H{
				{
					"id":          1,
					"name":        "Núcleo Auth/Admin Go + Keycloak",
					"status":      "✅ Implementado",
					"description": "Autenticação multi-tenant, RBAC, gerenciamento de tenants",
				},
				{
					"id":          2,
					"name":        "API Gateway, Health, OPA",
					"status":      "🔄 Próximo",
					"description": "Gateway, health checks, Open Policy Agent",
				},
				{
					"id":          3,
					"name":        "Consulta Jurídica + Circuit Breaker",
					"status":      "📋 Planejado",
					"description": "APIs jurídicas, circuit breaker, observabilidade",
				},
				{
					"id":          4,
					"name":        "IA Jurídica (RAG + Avaliação)",
					"status":      "📋 Planejado",
					"description": "Python, LangChain, Vertex AI, Ragas",
				},
			},
			"features": gin.H{
				"multi_tenancy": gin.H{
					"description": "Isolamento completo de dados por tenant (escritório/profissional)",
					"implementation": "Keycloak Groups + tenant_id em todas as queries",
				},
				"rbac": gin.H{
					"roles": []string{"admin", "lawyer", "secretary", "client"},
					"description": "Controle de acesso baseado em papéis por tenant",
				},
				"security": gin.H{
					"jwt_validation": "Keycloak + Redis cache",
					"audit_logging":  "Todas as ações registradas para compliance",
					"data_isolation": "Queries sempre filtradas por tenant_id",
				},
				"plans": []gin.H{
					{
						"name":   "Starter",
						"price":  99.90,
						"users":  1,
						"clients": 50,
						"cases":  100,
					},
					{
						"name":   "Professional", 
						"price":  299.90,
						"users":  5,
						"clients": 500,
						"cases":  1000,
					},
					{
						"name":   "Enterprise",
						"price":  999.90,
						"users":  -1, // unlimited
						"clients": -1,
						"cases":  -1,
					},
				},
			},
			"endpoints": gin.H{
				"authentication": []string{
					"POST /api/v1/auth/login",
					"POST /api/v1/auth/refresh", 
					"POST /api/v1/auth/forgot-password",
				},
				"tenants": []string{
					"POST /api/v1/tenants (admin only)",
					"GET /api/v1/tenants (admin only)",
					"GET /api/v1/tenants/:id (admin only)",
					"PUT /api/v1/tenants/:id (admin only)",
					"GET /api/v1/tenants/:id/usage (admin only)",
				},
				"profile": []string{
					"GET /api/v1/profile",
					"PUT /api/v1/profile",
				},
			},
		})
	})

	// Mock Keycloak status
	router.GET("/api/v1/keycloak/status", func(c *gin.Context) {
		cfg := &config.KeycloakConfig{
			BaseURL:  "http://localhost:8080",
			Realm:    "direito-lux",
			ClientID: "direito-lux-app",
		}
		
		keycloakClient := auth.NewKeycloakClient(cfg)
		
		// Try to get public key to test connection
		_, err := keycloakClient.GetPublicKey(c.Request.Context())
		
		status := "connected"
		if err != nil {
			status = "disconnected: " + err.Error()
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": status,
			"config": gin.H{
				"base_url": cfg.BaseURL,
				"realm":    cfg.Realm,
				"client_id": cfg.ClientID,
			},
			"urls": gin.H{
				"admin_console":   "http://localhost:8080/admin",
				"account_console": "http://localhost:8080/realms/direito-lux/account",
				"realm_info":      "http://localhost:8080/realms/direito-lux",
			},
		})
	})

	// Mock tenant creation demo
	router.POST("/api/v1/demo/tenant", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			// Ignore error for demo purposes
			req = make(map[string]interface{})
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Demo - Tenant creation simulation",
			"request": req,
			"actions": []string{
				"1. Validate tenant name and plan",
				"2. Create Keycloak group for tenant isolation", 
				"3. Create tenant record in PostgreSQL",
				"4. Create subscription with trial status",
				"5. Create admin user in Keycloak",
				"6. Set user password and send verification email",
				"7. Create user record with admin role",
				"8. Log audit trail for compliance",
			},
			"note": "This is a demo endpoint. Real implementation requires database connection.",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", `
<!DOCTYPE html>
<html>
<head>
    <title>Direito Lux - SaaS Jurídico Enterprise</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 40px; line-height: 1.6; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px; margin-bottom: 30px; }
        .status { background: #f8f9fa; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .endpoint { background: white; border: 1px solid #e9ecef; padding: 15px; margin: 10px 0; border-radius: 6px; }
        .success { color: #28a745; font-weight: bold; }
        .pending { color: #ffc107; font-weight: bold; }
        .method { background: #007bff; color: white; padding: 2px 8px; border-radius: 3px; font-family: monospace; font-size: 12px; }
        .url { font-family: monospace; color: #6f42c1; }
        pre { background: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🏛️ Direito Lux</h1>
        <p>SaaS Jurídico Enterprise - Módulo 1 Implementado</p>
        <p><strong>Backend em Go + Keycloak + PostgreSQL + Redis</strong></p>
    </div>

    <div class="status">
        <h2>📊 Status dos Serviços</h2>
        <p><span class="success">✅ API Go</span> - Rodando na porta 9000</p>
        <p><span class="success">✅ Keycloak</span> - Rodando na porta 8080</p>
        <p><span class="success">✅ PostgreSQL</span> - Banco keycloak + direito_lux</p>
        <p><span class="success">✅ Redis</span> - Cache de tokens JWT</p>
        <p><span class="success">✅ Nginx</span> - Load balancer</p>
    </div>

    <div class="status">
        <h2>🔧 Endpoints Implementados</h2>
        
        <div class="endpoint">
            <p><span class="method">GET</span> <span class="url">/health</span></p>
            <p>Health check da API com informações do sistema</p>
            <button onclick="fetch('/health').then(r=>r.json()).then(d=>alert(JSON.stringify(d,null,2)))">Testar</button>
        </div>

        <div class="endpoint">
            <p><span class="method">GET</span> <span class="url">/api/v1/info</span></p>
            <p>Informações completas da arquitetura, módulos e endpoints</p>
            <button onclick="fetch('/api/v1/info').then(r=>r.json()).then(d=>console.log(d))">Testar (ver console)</button>
        </div>

        <div class="endpoint">
            <p><span class="method">GET</span> <span class="url">/api/v1/keycloak/status</span></p>
            <p>Status da conexão com Keycloak e URLs importantes</p>
            <button onclick="fetch('/api/v1/keycloak/status').then(r=>r.json()).then(d=>alert(JSON.stringify(d,null,2)))">Testar</button>
        </div>

        <div class="endpoint">
            <p><span class="method">POST</span> <span class="url">/api/v1/demo/tenant</span></p>
            <p>Simulação de criação de tenant (demo sem banco)</p>
            <button onclick="
                fetch('/api/v1/demo/tenant', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        name: 'escritorio-silva',
                        display_name: 'Escritório Silva & Associados',
                        plan_id: 'uuid-plan-professional',
                        admin_user: {
                            email: 'admin@escritoriosilva.com',
                            first_name: 'João',
                            last_name: 'Silva'
                        }
                    })
                }).then(r=>r.json()).then(d=>alert(JSON.stringify(d,null,2)))
            ">Testar Criação</button>
        </div>
    </div>

    <div class="status">
        <h2>🔗 Links Úteis</h2>
        <p><a href="http://localhost:8080/admin" target="_blank">🔐 Keycloak Admin Console</a> (admin/admin)</p>
        <p><a href="http://localhost:8080/realms/direito-lux/account" target="_blank">👤 Keycloak Account Console</a></p>
        <p><a href="http://localhost:8080/realms/direito-lux" target="_blank">📋 Realm Info</a></p>
        <p><a href="https://github.com/opiagile/direito-lux" target="_blank">📚 Código no GitHub</a></p>
    </div>

    <div class="status">
        <h2>📈 Próximos Módulos</h2>
        <p><span class="pending">🔄 Módulo 2:</span> API Gateway, Health, OPA</p>
        <p><span class="pending">📋 Módulo 3:</span> Consulta Jurídica + Circuit Breaker</p>
        <p><span class="pending">📋 Módulo 4:</span> IA Jurídica (RAG + Avaliação)</p>
    </div>

    <script>
        console.log('Direito Lux API Demo - Módulo 1');
        console.log('Teste os endpoints usando os botões ou diretamente no console:');
        console.log('fetch("/api/v1/info").then(r=>r.json()).then(console.log)');
    </script>
</body>
</html>
		`)
	})

	router.Static("/static", "./static")
	
	// Start server
	if err := router.Run(":9000"); err != nil {
		panic(err)
	}
}