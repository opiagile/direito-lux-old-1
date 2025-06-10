package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/opiagile/direito-lux/internal/domain"
	"github.com/opiagile/direito-lux/pkg/circuitbreaker"
	"github.com/opiagile/direito-lux/pkg/logger"
)

type ConsultaService struct {
	logger   logger.Logger
	circuit  circuitbreaker.CircuitBreaker
	client   *http.Client
}

func NewConsultaService(log logger.Logger, cb circuitbreaker.CircuitBreaker) *ConsultaService {
	return &ConsultaService{
		logger: log,
		circuit: cb,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// ConsultarProcesso busca informações de processo judicial
func (s *ConsultaService) ConsultarProcesso(ctx context.Context, numeroProcesso string, tribunal string) (*domain.ConsultaProcesso, error) {
	consultaID := uuid.New().String()
	
	s.logger.Info("Iniciando consulta de processo", 
		"consulta_id", consultaID,
		"numero_processo", numeroProcesso,
		"tribunal", tribunal)
	
	// Executa com circuit breaker
	result, err := s.circuit.Execute(func() (interface{}, error) {
		return s.buscarProcessoExterno(ctx, numeroProcesso, tribunal)
	})
	
	if err != nil {
		s.logger.Error("Erro na consulta de processo", 
			"consulta_id", consultaID,
			"error", err)
		return nil, err
	}
	
	processo := result.(*domain.ProcessoJudicial)
	
	consulta := &domain.ConsultaProcesso{
		ID:              consultaID,
		NumeroProcesso:  numeroProcesso,
		Tribunal:        tribunal,
		Status:          "concluida",
		DataConsulta:    time.Now(),
		Processo:        processo,
	}
	
	s.logger.Info("Consulta de processo concluída",
		"consulta_id", consultaID,
		"status", consulta.Status)
	
	return consulta, nil
}

// ConsultarLegislacao busca legislação relevante
func (s *ConsultaService) ConsultarLegislacao(ctx context.Context, tema string, jurisdicao string) (*domain.ConsultaLegislacao, error) {
	consultaID := uuid.New().String()
	
	s.logger.Info("Iniciando consulta de legislação",
		"consulta_id", consultaID,
		"tema", tema,
		"jurisdicao", jurisdicao)
	
	// Executa com circuit breaker
	result, err := s.circuit.Execute(func() (interface{}, error) {
		return s.buscarLegislacaoExterna(ctx, tema, jurisdicao)
	})
	
	if err != nil {
		s.logger.Error("Erro na consulta de legislação",
			"consulta_id", consultaID,
			"error", err)
		return nil, err
	}
	
	leis := result.([]*domain.Lei)
	
	consulta := &domain.ConsultaLegislacao{
		ID:           consultaID,
		Tema:         tema,
		Jurisdicao:   jurisdicao,
		Status:       "concluida",
		DataConsulta: time.Now(),
		Leis:         leis,
	}
	
	s.logger.Info("Consulta de legislação concluída",
		"consulta_id", consultaID,
		"total_leis", len(leis))
	
	return consulta, nil
}

// ConsultarJurisprudencia busca jurisprudência relevante
func (s *ConsultaService) ConsultarJurisprudencia(ctx context.Context, tema string, tribunal string) (*domain.ConsultaJurisprudencia, error) {
	consultaID := uuid.New().String()
	
	s.logger.Info("Iniciando consulta de jurisprudência",
		"consulta_id", consultaID,
		"tema", tema,
		"tribunal", tribunal)
	
	// Executa com circuit breaker  
	result, err := s.circuit.Execute(func() (interface{}, error) {
		return s.buscarJurisprudenciaExterna(ctx, tema, tribunal)
	})
	
	if err != nil {
		s.logger.Error("Erro na consulta de jurisprudência",
			"consulta_id", consultaID,
			"error", err)
		return nil, err
	}
	
	decisoes := result.([]*domain.Decisao)
	
	consulta := &domain.ConsultaJurisprudencia{
		ID:           consultaID,
		Tema:         tema,
		Tribunal:     tribunal,
		Status:       "concluida",
		DataConsulta: time.Now(),
		Decisoes:     decisoes,
	}
	
	s.logger.Info("Consulta de jurisprudência concluída",
		"consulta_id", consultaID,
		"total_decisoes", len(decisoes))
	
	return consulta, nil
}

// Métodos privados para chamadas externas (simulados)

func (s *ConsultaService) buscarProcessoExterno(ctx context.Context, numeroProcesso, tribunal string) (*domain.ProcessoJudicial, error) {
	// Simula chamada para API externa (ex: CNJ, TJxx)
	// Em produção, seria uma chamada real para APIs jurídicas
	
	// Simula latência
	time.Sleep(time.Millisecond * 500)
	
	// Simula falha ocasional para testar circuit breaker
	if numeroProcesso == "0000000-00.0000.0.00.0000" {
		return nil, fmt.Errorf("processo não encontrado")
	}
	
	processo := &domain.ProcessoJudicial{
		Numero:    numeroProcesso,
		Tribunal:  tribunal,
		Classe:    "Ação Civil Pública",
		Assunto:   "Direito Civil",
		Status:    "Em andamento",
		DataAutuacao: time.Now().AddDate(0, -6, 0),
		Partes: []domain.Parte{
			{Nome: "João Silva", Tipo: "Autor"},
			{Nome: "Maria Santos", Tipo: "Réu"},
		},
		Movimentacoes: []domain.Movimentacao{
			{
				Data:      time.Now().AddDate(0, 0, -1),
				Descricao: "Julgamento da ação",
				Tipo:      "Decisão",
			},
		},
	}
	
	return processo, nil
}

func (s *ConsultaService) buscarLegislacaoExterna(ctx context.Context, tema, jurisdicao string) ([]*domain.Lei, error) {
	// Simula chamada para API de legislação
	time.Sleep(time.Millisecond * 300)
	
	leis := []*domain.Lei{
		{
			ID:          "lei-1",
			Numero:      "10.406/2002",
			Nome:        "Código Civil",
			Ementa:      "Institui o Código Civil.",
			DataPublicacao: time.Date(2002, 1, 10, 0, 0, 0, 0, time.UTC),
			Jurisdicao:  "Federal",
			Status:      "Vigente",
		},
		{
			ID:          "lei-2", 
			Numero:      "8.078/1990",
			Nome:        "Código de Defesa do Consumidor",
			Ementa:      "Dispõe sobre a proteção do consumidor.",
			DataPublicacao: time.Date(1990, 9, 11, 0, 0, 0, 0, time.UTC),
			Jurisdicao:  "Federal",
			Status:      "Vigente",
		},
	}
	
	return leis, nil
}

func (s *ConsultaService) buscarJurisprudenciaExterna(ctx context.Context, tema, tribunal string) ([]*domain.Decisao, error) {
	// Simula chamada para API de jurisprudência
	time.Sleep(time.Millisecond * 400)
	
	decisoes := []*domain.Decisao{
		{
			ID:          "decisao-1",
			Tribunal:    tribunal,
			Numero:      "REsp 1.234.567",
			Relator:     "Min. João Silva",
			DataJulgamento: time.Now().AddDate(0, -2, 0),
			Ementa:      "Direito Civil. Responsabilidade civil. Dano moral.",
			Resultado:   "Provido",
			Tema:        tema,
		},
		{
			ID:          "decisao-2",
			Tribunal:    tribunal,
			Numero:      "REsp 7.890.123",
			Relator:     "Min. Maria Santos",
			DataJulgamento: time.Now().AddDate(0, -1, 0),
			Ementa:      "Processo Civil. Competência. Foro privilegiado.",
			Resultado:   "Negado provimento",
			Tema:        tema,
		},
	}
	
	return decisoes, nil
}