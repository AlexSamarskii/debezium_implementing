package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`

	MaxOpenConns           int           `mapstructure:"max_open_conns"`
	MaxIdleConns           int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMinutes int           `mapstructure:"conn_max_lifetime_minutes"`
	ConnMaxLifetime        time.Duration `mapstructure:"-"`
}

type ElasticsearchConfig struct {
	URL string `mapstructure:"url"`
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
	Database      DatabaseConfig      `mapstructure:"database"`
	Server        ServerConfig        `mapstructure:"server"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
}

func Load() (*Config, error) {
	var cfg Config

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using env: %v", err)
		viper.BindEnv("database.host", "DB_HOST")
		viper.BindEnv("database.port", "DB_PORT")
		viper.BindEnv("database.user", "DB_USER")
		viper.BindEnv("database.password", "DB_PASSWORD")
		viper.BindEnv("database.name", "DB_NAME")
		viper.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
		viper.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
		viper.BindEnv("database.conn_max_lifetime_minutes", "DB_CONN_MAX_LIFETIME_MINUTES")
		viper.BindEnv("server.port", "SERVER_PORT")
		viper.BindEnv("kafka.brokers", "KAFKA_BROKERS")
		viper.BindEnv("kafka.topic", "KAFKA_TOPIC")
		viper.BindEnv("kafka.group", "KAFKA_GROUP")
		viper.BindEnv("elasticsearch.url", "ELASTICSEARCH_URL")

		if brokersStr := viper.GetString("kafka.brokers"); brokersStr != "" && len(cfg.Kafka.Brokers) == 0 {
			cfg.Kafka.Brokers = strings.Split(brokersStr, ",")
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.Database.ConnMaxLifetime = time.Duration(cfg.Database.ConnMaxLifetimeMinutes) * time.Minute

	log.Printf("Config loaded: server.port=%d, db.host=%s, kafka.brokers=%s, kafka.topic=%s, elasticsearch.url=%s",
		cfg.Server.Port, cfg.Database.Host, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Elasticsearch.URL)

	return &cfg, nil
}
