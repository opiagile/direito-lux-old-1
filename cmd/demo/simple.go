package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// CORS
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
		})
	})

	// API info
	router.GET("/api/v1/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "Direito Lux - SaaS Jurídico Enterprise",
			"version":     "1.0.0",
			"status":      "Módulo 1 Implementado",
			"modules": []string{
				"✅ Núcleo Auth/Admin Go + Keycloak",
				"🔄 API Gateway, Health, OPA",
				"📋 Consulta Jurídica + Circuit Breaker",
				"📋 IA Jurídica (RAG + Avaliação)",
			},
		})
	})

	// Main page
	router.GET("/", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Direito Lux - SaaS Jurídico</title>
    <meta charset="UTF-8">
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; 
            margin: 0; padding: 20px; background: #f5f7fa; line-height: 1.6; 
        }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
            color: white; padding: 30px; border-radius: 15px; margin-bottom: 30px; text-align: center; 
        }
        .card { 
            background: white; padding: 25px; margin: 15px 0; border-radius: 10px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1); 
        }
        .btn { 
            background: #007bff; color: white; padding: 10px 20px; border: none; 
            border-radius: 5px; cursor: pointer; margin: 5px; font-size: 14px; 
        }
        .btn:hover { background: #0056b3; }
        .success { color: #28a745; font-weight: bold; }
        .endpoint { 
            background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 8px; 
            border-left: 4px solid #007bff; 
        }
        .method { 
            background: #007bff; color: white; padding: 3px 8px; border-radius: 3px; 
            font-family: monospace; font-size: 12px; margin-right: 10px; 
        }
        .url { font-family: monospace; color: #6f42c1; font-weight: bold; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        pre { background: #f1f3f4; padding: 15px; border-radius: 5px; overflow-x: auto; font-size: 12px; }
        .status-good { color: #28a745; }
        .status-pending { color: #ffc107; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏛️ Direito Lux</h1>
            <h3>SaaS Jurídico Enterprise - Módulo 1 Implementado</h3>
            <p>Backend em Go + Keycloak + PostgreSQL + Redis</p>
        </div>

        <div class="grid">
            <div class="card">
                <h2>📊 Status dos Serviços</h2>
                <p><span class="status-good">✅ API Go</span> - Porta 9000</p>
                <p><span class="status-good">✅ Keycloak</span> - Porta 8080</p>
                <p><span class="status-good">✅ PostgreSQL</span> - Banco de dados</p>
                <p><span class="status-good">✅ Redis</span> - Cache JWT</p>
                <p><span class="status-good">✅ Nginx</span> - Load balancer</p>
            </div>

            <div class="card">
                <h2>🔧 Endpoints da API</h2>
                
                <div class="endpoint">
                    <p><span class="method">GET</span> <span class="url">/health</span></p>
                    <p>Health check da API</p>
                    <button class="btn" onclick="testEndpoint('/health')">Testar</button>
                </div>

                <div class="endpoint">
                    <p><span class="method">GET</span> <span class="url">/api/v1/info</span></p>
                    <p>Informações da API</p>
                    <button class="btn" onclick="testEndpoint('/api/v1/info')">Testar</button>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>🔗 Links Importantes</h2>
            <div style="display: flex; flex-wrap: wrap; gap: 10px;">
                <a href="http://localhost:8080/admin" target="_blank" class="btn">🔐 Keycloak Admin</a>
                <a href="http://localhost:8080/realms/direito-lux/account" target="_blank" class="btn">👤 Account Console</a>
                <a href="http://localhost:8080/realms/direito-lux" target="_blank" class="btn">📋 Realm Info</a>
                <a href="https://github.com/opiagile/direito-lux" target="_blank" class="btn">📚 GitHub</a>
            </div>
        </div>

        <div class="card">
            <h2>📈 Módulos do Projeto</h2>
            <p><span class="status-good">✅ Módulo 1:</span> Núcleo Auth/Admin Go + Keycloak</p>
            <p><span class="status-pending">🔄 Módulo 2:</span> API Gateway, Health, OPA</p>
            <p><span class="status-pending">📋 Módulo 3:</span> Consulta Jurídica + Circuit Breaker</p>
            <p><span class="status-pending">📋 Módulo 4:</span> IA Jurídica (RAG + Avaliação)</p>
        </div>

        <div class="card">
            <h2>📝 Resultado dos Testes</h2>
            <div id="results" style="min-height: 100px; background: #f8f9fa; padding: 15px; border-radius: 5px;">
                <p>Clique nos botões "Testar" acima para ver as respostas da API.</p>
            </div>
        </div>
    </div>

    <script>
        async function testEndpoint(url) {
            const resultsDiv = document.getElementById('results');
            resultsDiv.innerHTML = '<p>🔄 Testando ' + url + '...</p>';
            
            try {
                const response = await fetch(url);
                const data = await response.json();
                
                resultsDiv.innerHTML = '<h4>✅ Resposta de ' + url + '</h4><pre>' + 
                    JSON.stringify(data, null, 2) + '</pre>';
            } catch (error) {
                resultsDiv.innerHTML = '<h4>❌ Erro ao testar ' + url + '</h4><p>' + 
                    error.message + '</p>';
            }
        }

        // Test automatic health check on page load
        document.addEventListener('DOMContentLoaded', function() {
            console.log('Direito Lux API Demo - Módulo 1 carregado');
        });
    </script>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	router.Run(":9001")
}