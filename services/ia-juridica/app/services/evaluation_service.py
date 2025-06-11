"""
Evaluation Service using Ragas for RAG quality assessment
"""

import asyncio
from typing import List, Dict, Any, Optional
from datetime import datetime
import json

from datasets import Dataset
from ragas import evaluate
from ragas.metrics import (
    answer_relevancy,
    context_precision,
    context_recall,
    faithfulness,
    answer_similarity,
    answer_correctness
)

from app.core.config import Settings
from app.core.logging import LoggingMixin


class EvaluationService(LoggingMixin):
    """Service for evaluating RAG system quality using Ragas"""
    
    def __init__(self, settings: Settings):
        self.settings = settings
        self._is_initialized = False
        self._is_healthy = False
        
        # Evaluation metrics
        self.metrics = [
            faithfulness,           # How faithful the answer is to the context
            answer_relevancy,       # How relevant the answer is to the question
            context_precision,      # How precise the retrieved context is
            context_recall,         # How well the context covers the answer
            answer_similarity,      # Semantic similarity of answers
            answer_correctness      # Correctness of the answer
        ]
        
        # Evaluation history
        self.evaluation_history = []
    
    async def initialize(self) -> None:
        """Initialize evaluation service"""
        try:
            self.log_info("Inicializando Evaluation Service (Ragas)")
            
            if not self.settings.evaluation_enabled:
                self.log_info("Avaliação desabilitada na configuração")
                self._is_initialized = True
                self._is_healthy = True
                return
            
            # Test Ragas functionality
            await self._test_ragas_setup()
            
            self._is_initialized = True
            self._is_healthy = True
            self.log_info("Evaluation Service inicializado com sucesso")
            
        except Exception as e:
            self._is_healthy = False
            self.log_error("Erro ao inicializar Evaluation Service", error=str(e))
            raise
    
    async def _test_ragas_setup(self) -> None:
        """Test Ragas setup with a minimal evaluation"""
        try:
            # Create a minimal test dataset
            test_data = {
                "question": ["What is artificial intelligence?"],
                "answer": ["Artificial intelligence is a field of computer science."],
                "contexts": [["AI is the simulation of human intelligence in machines."]],
                "ground_truth": ["AI is the simulation of human intelligence processes by machines."]
            }
            
            test_dataset = Dataset.from_dict(test_data)
            
            # Run a quick evaluation with one metric
            await asyncio.to_thread(
                evaluate,
                test_dataset,
                metrics=[faithfulness]
            )
            
            self.log_debug("Ragas test evaluation successful")
            
        except Exception as e:
            self.log_error("Ragas test setup failed", error=str(e))
            raise
    
    async def evaluate_rag_response(
        self,
        question: str,
        answer: str,
        contexts: List[str],
        ground_truth: Optional[str] = None
    ) -> Dict[str, Any]:
        """Evaluate a single RAG response"""
        try:
            if not self._is_initialized or not self.settings.evaluation_enabled:
                return {"status": "evaluation_disabled"}
            
            start_time = datetime.utcnow()
            
            # Prepare evaluation data
            eval_data = {
                "question": [question],
                "answer": [answer],
                "contexts": [contexts]
            }
            
            # Add ground truth if available
            if ground_truth:
                eval_data["ground_truth"] = [ground_truth]
                metrics_to_use = self.metrics
            else:
                # Use metrics that don't require ground truth
                metrics_to_use = [faithfulness, answer_relevancy, context_precision]
            
            # Create dataset
            dataset = Dataset.from_dict(eval_data)
            
            # Run evaluation
            self.log_info("Executando avaliação Ragas", 
                         question_length=len(question),
                         contexts_count=len(contexts))
            
            result = await asyncio.to_thread(
                evaluate,
                dataset,
                metrics=metrics_to_use
            )
            
            # Extract scores
            scores = {}
            for metric_name, score_value in result.items():
                if hasattr(score_value, 'iloc'):
                    # Handle pandas Series
                    scores[metric_name] = float(score_value.iloc[0]) if len(score_value) > 0 else 0.0
                else:
                    scores[metric_name] = float(score_value)
            
            # Calculate overall score
            overall_score = sum(scores.values()) / len(scores) if scores else 0.0
            
            processing_time = (datetime.utcnow() - start_time).total_seconds()
            
            evaluation_result = {
                "evaluation_id": f"eval_{int(datetime.utcnow().timestamp())}",
                "timestamp": datetime.utcnow().isoformat(),
                "question": question,
                "answer_length": len(answer),
                "contexts_count": len(contexts),
                "scores": scores,
                "overall_score": overall_score,
                "processing_time": processing_time,
                "has_ground_truth": ground_truth is not None
            }
            
            # Store in history
            self.evaluation_history.append(evaluation_result)
            
            # Keep only recent evaluations
            if len(self.evaluation_history) > 1000:
                self.evaluation_history = self.evaluation_history[-1000:]
            
            self.log_info("Avaliação Ragas concluída",
                         overall_score=overall_score,
                         processing_time=processing_time)
            
            return evaluation_result
            
        except Exception as e:
            self.log_error("Erro na avaliação Ragas", error=str(e))
            return {
                "error": str(e),
                "status": "evaluation_failed",
                "timestamp": datetime.utcnow().isoformat()
            }
    
    async def batch_evaluate(
        self,
        evaluation_data: List[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """Evaluate multiple RAG responses in batch"""
        try:
            if not self._is_initialized or not self.settings.evaluation_enabled:
                return {"status": "evaluation_disabled"}
            
            if not evaluation_data:
                return {"error": "No evaluation data provided"}
            
            start_time = datetime.utcnow()
            
            # Prepare batch data
            questions = []
            answers = []
            contexts_list = []
            ground_truths = []
            has_ground_truth = False
            
            for item in evaluation_data:
                questions.append(item["question"])
                answers.append(item["answer"])
                contexts_list.append(item["contexts"])
                
                if "ground_truth" in item:
                    ground_truths.append(item["ground_truth"])
                    has_ground_truth = True
            
            # Create batch dataset
            batch_data = {
                "question": questions,
                "answer": answers,
                "contexts": contexts_list
            }
            
            if has_ground_truth and len(ground_truths) == len(questions):
                batch_data["ground_truth"] = ground_truths
                metrics_to_use = self.metrics
            else:
                metrics_to_use = [faithfulness, answer_relevancy, context_precision]
            
            dataset = Dataset.from_dict(batch_data)
            
            self.log_info("Executando avaliação em lote",
                         batch_size=len(evaluation_data))
            
            # Run batch evaluation
            results = await asyncio.to_thread(
                evaluate,
                dataset,
                metrics=metrics_to_use
            )
            
            # Process results
            processed_results = {}
            for metric_name, scores in results.items():
                if hasattr(scores, 'tolist'):
                    processed_results[metric_name] = {
                        "scores": scores.tolist(),
                        "mean": float(scores.mean()),
                        "std": float(scores.std()),
                        "min": float(scores.min()),
                        "max": float(scores.max())
                    }
                else:
                    processed_results[metric_name] = scores
            
            # Calculate overall statistics
            all_scores = []
            for metric_data in processed_results.values():
                if isinstance(metric_data, dict) and "scores" in metric_data:
                    all_scores.extend(metric_data["scores"])
            
            overall_stats = {
                "mean": sum(all_scores) / len(all_scores) if all_scores else 0.0,
                "count": len(evaluation_data)
            }
            
            processing_time = (datetime.utcnow() - start_time).total_seconds()
            
            batch_result = {
                "batch_id": f"batch_{int(datetime.utcnow().timestamp())}",
                "timestamp": datetime.utcnow().isoformat(),
                "batch_size": len(evaluation_data),
                "metrics": processed_results,
                "overall_stats": overall_stats,
                "processing_time": processing_time,
                "has_ground_truth": has_ground_truth
            }
            
            self.log_info("Avaliação em lote concluída",
                         batch_size=len(evaluation_data),
                         overall_mean=overall_stats["mean"],
                         processing_time=processing_time)
            
            return batch_result
            
        except Exception as e:
            self.log_error("Erro na avaliação em lote", error=str(e))
            return {
                "error": str(e),
                "status": "batch_evaluation_failed",
                "timestamp": datetime.utcnow().isoformat()
            }
    
    async def get_evaluation_summary(
        self,
        days: int = 7
    ) -> Dict[str, Any]:
        """Get evaluation summary for the last N days"""
        try:
            if not self.evaluation_history:
                return {
                    "message": "No evaluation history available",
                    "total_evaluations": 0
                }
            
            # Filter recent evaluations
            cutoff_time = datetime.utcnow().timestamp() - (days * 24 * 3600)
            recent_evaluations = [
                eval_data for eval_data in self.evaluation_history
                if datetime.fromisoformat(eval_data["timestamp"]).timestamp() > cutoff_time
            ]
            
            if not recent_evaluations:
                return {
                    "message": f"No evaluations in the last {days} days",
                    "total_evaluations": len(self.evaluation_history)
                }
            
            # Calculate summary statistics
            all_scores = []
            metric_stats = {}
            
            for eval_data in recent_evaluations:
                if "scores" in eval_data:
                    for metric_name, score in eval_data["scores"].items():
                        if metric_name not in metric_stats:
                            metric_stats[metric_name] = []
                        metric_stats[metric_name].append(score)
                        all_scores.append(score)
            
            # Process metric statistics
            processed_metrics = {}
            for metric_name, scores in metric_stats.items():
                processed_metrics[metric_name] = {
                    "mean": sum(scores) / len(scores),
                    "min": min(scores),
                    "max": max(scores),
                    "count": len(scores)
                }
            
            summary = {
                "period_days": days,
                "total_evaluations": len(recent_evaluations),
                "overall_mean_score": sum(all_scores) / len(all_scores) if all_scores else 0.0,
                "metric_statistics": processed_metrics,
                "evaluation_frequency": len(recent_evaluations) / days,
                "last_evaluation": recent_evaluations[-1]["timestamp"] if recent_evaluations else None
            }
            
            return summary
            
        except Exception as e:
            self.log_error("Erro ao gerar resumo de avaliações", error=str(e))
            return {"error": str(e)}
    
    def is_healthy(self) -> bool:
        """Check if the evaluation service is healthy"""
        return self._is_healthy and self._is_initialized
    
    async def close(self) -> None:
        """Close evaluation service"""
        try:
            self.log_info("Fechando Evaluation Service")
            
            # Save evaluation history if needed
            if self.evaluation_history:
                self.log_info("Salvando histórico de avaliações",
                             count=len(self.evaluation_history))
            
            self._is_initialized = False
            self._is_healthy = False
            
        except Exception as e:
            self.log_error("Erro ao fechar Evaluation Service", error=str(e))