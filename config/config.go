package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`

	MaxOpenConns           int           `mapstructure:"max_open_conns"`
	MaxIdleConns           int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMinutes int           `mapstructure:"conn_max_lifetime_minutes"`
	ConnMaxLifetime        time.Duration `mapstructure:"-"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
	Group   string   `mapstructure:"group"`
}

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
}

func Load() (*Config, error) {
	var cfg Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config") // ./config/config.yaml

	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()

	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("kafka.brokers", "KAFKA_BROKERS")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using env/flags only: %v", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.Database.ConnMaxLifetime = time.Duration(cfg.Database.ConnMaxLifetimeMinutes) * time.Minute

	log.Printf("Config loaded: server.port=%d, db.host=%s, kafka.topic=%s",
		cfg.Server.Port, cfg.Database.Host, cfg.Kafka.Topic)

	return &cfg, nil
}
