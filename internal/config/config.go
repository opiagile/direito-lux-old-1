package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server          ServerConfig
	Database        DatabaseConfig
	Redis           RedisConfig
	Keycloak        KeycloakConfig
	JWT             JWTConfig
	Logger          LoggerConfig
	ConsultaService ConsultaServiceConfig
}

type ServerConfig struct {
	Port         string
	Mode         string // debug, release, test
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

type KeycloakConfig struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
	AdminUser    string
	AdminPass    string
}

type JWTConfig struct {
	PublicKeyPath   string
	CacheDuration   time.Duration
	ClockSkewLeeway time.Duration
}

type LoggerConfig struct {
	Level      string
	Encoding   string // json or console
	OutputPath string
}

type ConsultaServiceConfig struct {
	Port string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("DIREITO_LUX")

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.readTimeout", "15s")
	viper.SetDefault("server.writeTimeout", "15s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "direito_lux")
	viper.SetDefault("database.sslmode", "disable")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.poolSize", 10)
	viper.SetDefault("redis.minIdleConns", 5)

	// Keycloak defaults
	viper.SetDefault("keycloak.baseURL", "http://localhost:8080")
	viper.SetDefault("keycloak.realm", "direito-lux")
	viper.SetDefault("keycloak.clientID", "direito-lux-app")
	viper.SetDefault("keycloak.adminUser", "admin")
	viper.SetDefault("keycloak.adminPass", "admin")

	// JWT defaults
	viper.SetDefault("jwt.cacheDuration", "5m")
	viper.SetDefault("jwt.clockSkewLeeway", "5m")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.encoding", "json")
	viper.SetDefault("logger.outputPath", "stdout")

	// Consulta Service defaults
	viper.SetDefault("consultaService.port", "9002")
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release" || os.Getenv("GIN_MODE") == "release"
}

// LoadConfig loads configuration with defaults for consulta service
func LoadConfig() *Config {
	config, err := Load()
	if err != nil {
		// Return default config if file doesn't exist
		return &Config{
			Logger: LoggerConfig{
				Level:    "info",
				Encoding: "json",
			},
			ConsultaService: ConsultaServiceConfig{
				Port: "9002",
			},
		}
	}
	return config
}
