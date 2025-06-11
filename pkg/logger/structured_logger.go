package logger

import (
	"fmt"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
}

// StructuredLogger implements Logger interface with support for multiple outputs
type StructuredLogger struct {
	zap *zap.Logger
}

// NewLogger creates a new structured logger with optional Logstash output
func NewLogger(level string) Logger {
	config := zap.NewProductionConfig()

	// Set log level
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		logLevel = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	// Custom encoder config for structured logging
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Output to stdout by default
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Build core logger
	cores := []zapcore.Core{}

	// Console output
	consoleEncoder := zapcore.NewJSONEncoder(config.EncoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		config.Level,
	)
	cores = append(cores, consoleCore)

	// Logstash output (if configured)
	logstashHost := os.Getenv("LOGSTASH_HOST")
	if logstashHost != "" {
		if logstashCore := createLogstashCore(logstashHost, config.EncoderConfig, config.Level); logstashCore != nil {
			cores = append(cores, logstashCore)
		}
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Add service context
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).With(
		zap.String("service", "direito-lux-consulta"),
		zap.String("module", "module-3"),
		zap.String("version", "1.0.0"),
		zap.String("environment", getEnv("GIN_MODE", "development")),
	)

	return &StructuredLogger{zap: logger}
}

// createLogstashCore creates a zapcore.Core that sends logs to Logstash via TCP
func createLogstashCore(logstashHost string, encoderConfig zapcore.EncoderConfig, level zap.AtomicLevel) zapcore.Core {
	// Try to connect to Logstash
	conn, err := net.DialTimeout("tcp", logstashHost, 5*time.Second)
	if err != nil {
		// If can't connect, return nil (will use only console output)
		fmt.Printf("Warning: Could not connect to Logstash at %s: %v\n", logstashHost, err)
		return nil
	}

	// Create TCP writer
	writer := &tcpWriter{conn: conn, address: logstashHost}

	// Create encoder for Logstash (JSON format)
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	return zapcore.NewCore(encoder, zapcore.AddSync(writer), level)
}

// tcpWriter implements zapcore.WriteSyncer for TCP connections
type tcpWriter struct {
	conn    net.Conn
	address string
}

func (w *tcpWriter) Write(p []byte) (n int, err error) {
	if w.conn == nil {
		// Try to reconnect
		conn, err := net.DialTimeout("tcp", w.address, 5*time.Second)
		if err != nil {
			return 0, err
		}
		w.conn = conn
	}

	return w.conn.Write(append(p, '\n'))
}

func (w *tcpWriter) Sync() error {
	return nil
}

// Logger interface implementation
func (l *StructuredLogger) Debug(msg string, fields ...interface{}) {
	l.zap.Debug(msg, l.parseFields(fields...)...)
}

func (l *StructuredLogger) Info(msg string, fields ...interface{}) {
	l.zap.Info(msg, l.parseFields(fields...)...)
}

func (l *StructuredLogger) Warn(msg string, fields ...interface{}) {
	l.zap.Warn(msg, l.parseFields(fields...)...)
}

func (l *StructuredLogger) Error(msg string, fields ...interface{}) {
	l.zap.Error(msg, l.parseFields(fields...)...)
}

func (l *StructuredLogger) Fatal(msg string, fields ...interface{}) {
	l.zap.Fatal(msg, l.parseFields(fields...)...)
}

func (l *StructuredLogger) With(fields ...interface{}) Logger {
	return &StructuredLogger{
		zap: l.zap.With(l.parseFields(fields...)...),
	}
}

// parseFields converts key-value pairs to zap.Field
func (l *StructuredLogger) parseFields(fields ...interface{}) []zap.Field {
	var zapFields []zap.Field

	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			zapFields = append(zapFields, zap.Any(key, value))
		}
	}

	return zapFields
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
