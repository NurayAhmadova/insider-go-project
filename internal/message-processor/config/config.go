package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DSN         string        `mapstructure:"dsn"`
	RedisAddr   string        `mapstructure:"redis_addr"`
	WebhookURL  string        `mapstructure:"webhook_url"`
	AuthKey     string        `mapstructure:"auth_key"`
	BatchSize   int32         `mapstructure:"batch_size"`
	PingTimeout time.Duration `mapstructure:"ping_timeout"`
	HTTPAddr    string        `mapstructure:"http_addr"`
}

func LoadConfig() (cfg Config, err error) {
	v := viper.New()

	v.SetConfigName("data")
	v.SetConfigType("yml")
	v.AddConfigPath("./cmd")

	if err := v.ReadInConfig(); err != nil {
		return cfg, err
	}

	v.AutomaticEnv()

	err = v.Unmarshal(&cfg)
	return cfg, err
}
