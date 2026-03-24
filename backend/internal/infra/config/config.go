package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Datasources DatasourcesConfig `mapstructure:"datasources"`
	LLM         LLMConfig         `mapstructure:"llm"`
	Alert       AlertConfig       `mapstructure:"alert"`
	Log         LogConfig         `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type DatasourcesConfig struct {
	Prometheus    PrometheusConfig    `mapstructure:"prometheus"`
	Loki          LokiConfig          `mapstructure:"loki"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Kubernetes    KubernetesConfig    `mapstructure:"kubernetes"`
}

type PrometheusConfig struct {
	URL string `mapstructure:"url"`
}

type LokiConfig struct {
	URL string `mapstructure:"url"`
}

type ElasticsearchConfig struct {
	URLs []string `mapstructure:"urls"`
}

type KubernetesConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig"`
}

type LLMConfig struct {
	BaseURL     string        `mapstructure:"base_url"`
	APIKey      string        `mapstructure:"api_key"`
	Model       string        `mapstructure:"model"`
	MaxTokens   int           `mapstructure:"max_tokens"`
	Temperature float32       `mapstructure:"temperature"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

type AlertConfig struct {
	Pipeline PipelineConfig `mapstructure:"pipeline"`
}

type PipelineConfig struct {
	DedupWindow         time.Duration `mapstructure:"dedup_window"`
	CorrelationWindow   time.Duration `mapstructure:"correlation_window"`
	FlappingThreshold   int           `mapstructure:"flapping_threshold"`
	FlappingWindow      time.Duration `mapstructure:"flapping_window"`
	AutoAnalyzeCritical bool          `mapstructure:"auto_analyze_critical"`
}

type LogConfig struct {
	AnomalyDetection AnomalyDetectionConfig `mapstructure:"anomaly_detection"`
}

type AnomalyDetectionConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	CheckInterval   time.Duration `mapstructure:"check_interval"`
	StddevThreshold float64       `mapstructure:"stddev_threshold"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Override LLM API key from env
	if apiKey := v.GetString("LLM_API_KEY"); apiKey != "" {
		cfg.LLM.APIKey = apiKey
	}

	return &cfg, nil
}
