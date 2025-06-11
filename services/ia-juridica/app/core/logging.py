"""
Logging configuration for Direito Lux IA Jurídica
"""

import logging
import sys
from typing import Any, Dict

import structlog
from structlog.stdlib import LoggerFactory


def setup_logging(log_level: str = "INFO") -> None:
    """Setup structured logging with structlog"""
    
    # Configure structlog
    structlog.configure(
        processors=[
            structlog.stdlib.filter_by_level,
            structlog.stdlib.add_logger_name,
            structlog.stdlib.add_log_level,
            structlog.stdlib.PositionalArgumentsFormatter(),
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.UnicodeDecoder(),
            structlog.processors.JSONRenderer()
        ],
        context_class=dict,
        logger_factory=LoggerFactory(),
        wrapper_class=structlog.stdlib.BoundLogger,
        cache_logger_on_first_use=True,
    )
    
    # Configure standard logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=getattr(logging, log_level.upper()),
    )
    
    # Set specific logger levels
    logging.getLogger("chromadb").setLevel(logging.WARNING)
    logging.getLogger("sentence_transformers").setLevel(logging.WARNING)
    logging.getLogger("transformers").setLevel(logging.WARNING)
    logging.getLogger("torch").setLevel(logging.WARNING)


def get_logger(name: str) -> structlog.BoundLogger:
    """Get a structured logger instance"""
    return structlog.get_logger(name)


class LoggingMixin:
    """Mixin to add logging capabilities to classes"""
    
    @property
    def logger(self) -> structlog.BoundLogger:
        """Get logger for this class"""
        return get_logger(self.__class__.__name__)
    
    def log_info(self, message: str, **kwargs: Any) -> None:
        """Log info message with context"""
        self.logger.info(message, **kwargs)
    
    def log_warning(self, message: str, **kwargs: Any) -> None:
        """Log warning message with context"""
        self.logger.warning(message, **kwargs)
    
    def log_error(self, message: str, **kwargs: Any) -> None:
        """Log error message with context"""
        self.logger.error(message, **kwargs)
    
    def log_debug(self, message: str, **kwargs: Any) -> None:
        """Log debug message with context"""
        self.logger.debug(message, **kwargs)