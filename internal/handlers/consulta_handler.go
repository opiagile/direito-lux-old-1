package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opiagile/direito-lux/internal/services"
	"github.com/opiagile/direito-lux/pkg/logger"
)

type ConsultaHandler struct {
	service *services.ConsultaService
	logger  logger.Logger
}

func NewConsultaHandler(service *services.ConsultaService, log logger.Logger) *ConsultaHandler {
	return &ConsultaHandler{
		service: service,
		logger:  log,
	}
}

// ConsultaProcessoRequest representa a requisição de consulta de processo
type ConsultaProcessoRequest struct {
	NumeroProcesso string `json:"numero_processo" binding:"required" example:"1234567-89.2023.1.23.4567"`
	Tribunal       string `json:"tribunal" binding:"required" example:"TJSP"`
}

// ConsultaLegislacaoRequest representa a requisição de consulta de legislação
type ConsultaLegislacaoRequest struct {
	Tema       string `json:"tema" binding:"required" example:"direito civil"`
	Jurisdicao string `json:"jurisdicao" binding:"required" example:"federal"`
}

// ConsultaJurisprudenciaRequest representa a requisição de consulta de jurisprudência
type ConsultaJurisprudenciaRequest struct {
	Tema     string `json:"tema" binding:"required" example:"responsabilidade civil"`
	Tribunal string `json:"tribunal" binding:"required" example:"STJ"`
}

// @Summary Consultar processo judicial
// @Description Busca informações de um processo judicial específico
// @Tags consultas
// @Accept json
// @Produce json
// @Param request body ConsultaProcessoRequest true "Dados da consulta"
// @Success 200 {object} domain.ConsultaProcesso
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /consultas/processos [post]
func (h *ConsultaHandler) ConsultarProcesso(c *gin.Context) {
	var req ConsultaProcessoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Erro no bind da requisição", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	requestID := c.GetHeader("X-Request-ID")
	h.logger.Info("Processando consulta de processo",
		"request_id", requestID,
		"numero_processo", req.NumeroProcesso,
		"tribunal", req.Tribunal)

	consulta, err := h.service.ConsultarProcesso(c.Request.Context(), req.NumeroProcesso, req.Tribunal)
	if err != nil {
		h.logger.Error("Erro ao consultar processo",
			"request_id", requestID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "consultation_failed",
			"message": "Falha na consulta do processo",
		})
		return
	}

	h.logger.Info("Consulta de processo realizada com sucesso",
		"request_id", requestID,
		"consulta_id", consulta.ID)

	c.JSON(http.StatusOK, consulta)
}

// @Summary Consultar legislação
// @Description Busca legislação relevante por tema e jurisdição
// @Tags consultas
// @Accept json
// @Produce json
// @Param request body ConsultaLegislacaoRequest true "Dados da consulta"
// @Success 200 {object} domain.ConsultaLegislacao
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /consultas/legislacao [post]
func (h *ConsultaHandler) ConsultarLegislacao(c *gin.Context) {
	var req ConsultaLegislacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Erro no bind da requisição", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	requestID := c.GetHeader("X-Request-ID")
	h.logger.Info("Processando consulta de legislação",
		"request_id", requestID,
		"tema", req.Tema,
		"jurisdicao", req.Jurisdicao)

	consulta, err := h.service.ConsultarLegislacao(c.Request.Context(), req.Tema, req.Jurisdicao)
	if err != nil {
		h.logger.Error("Erro ao consultar legislação",
			"request_id", requestID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "consultation_failed",
			"message": "Falha na consulta da legislação",
		})
		return
	}

	h.logger.Info("Consulta de legislação realizada com sucesso",
		"request_id", requestID,
		"consulta_id", consulta.ID)

	c.JSON(http.StatusOK, consulta)
}

// @Summary Consultar jurisprudência
// @Description Busca jurisprudência relevante por tema e tribunal
// @Tags consultas
// @Accept json
// @Produce json
// @Param request body ConsultaJurisprudenciaRequest true "Dados da consulta"
// @Success 200 {object} domain.ConsultaJurisprudencia
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /consultas/jurisprudencia [post]
func (h *ConsultaHandler) ConsultarJurisprudencia(c *gin.Context) {
	var req ConsultaJurisprudenciaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Erro no bind da requisição", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	requestID := c.GetHeader("X-Request-ID")
	h.logger.Info("Processando consulta de jurisprudência",
		"request_id", requestID,
		"tema", req.Tema,
		"tribunal", req.Tribunal)

	consulta, err := h.service.ConsultarJurisprudencia(c.Request.Context(), req.Tema, req.Tribunal)
	if err != nil {
		h.logger.Error("Erro ao consultar jurisprudência",
			"request_id", requestID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "consultation_failed",
			"message": "Falha na consulta da jurisprudência",
		})
		return
	}

	h.logger.Info("Consulta de jurisprudência realizada com sucesso",
		"request_id", requestID,
		"consulta_id", consulta.ID)

	c.JSON(http.StatusOK, consulta)
}

// @Summary Status da consulta
// @Description Retorna o status de uma consulta específica
// @Tags consultas
// @Produce json
// @Param id path string true "ID da consulta"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /consultas/status/{id} [get]
func (h *ConsultaHandler) StatusConsulta(c *gin.Context) {
	consultaID := c.Param("id")
	
	requestID := c.GetHeader("X-Request-ID")
	h.logger.Info("Consultando status",
		"request_id", requestID,
		"consulta_id", consultaID)

	// Em produção, buscaria no banco de dados
	// Por ora, retorna status mockado
	c.JSON(http.StatusOK, gin.H{
		"consulta_id": consultaID,
		"status":      "concluida",
		"timestamp":   "2023-06-10T15:30:00Z",
	})
}