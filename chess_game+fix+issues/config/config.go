package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/gin-contrib/cors"
)

type Config struct {
	DBDriver           string        `mapstructure:"DB_DRIVER"`
	DB_URL             string        `mapstructure:"DB_URL"`
	ServerAddress      string        `mapstructure:"SERVER_ADDRESS"`
	AcessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	TokenSymmetricKey  string        `mapstructure:"TokenSymmetricKey"`
	RESEND_API_KEY     string        `mapstructure:"RESEND_API_KEY"`
	CloudflareTURNKeyID    string `mapstructure:"CLOUDFLARE_TURN_KEY_ID"`
	CloudflareTURNAPIToken string `mapstructure:"CLOUDFLARE_TURN_API_TOKEN"`
	Environment        string        `mapstructure:"ENVIRONMENT"`
	PUBLIC_HOST         string `mapstructure:"PUBLIC_HOST"`
	VAPIDPublicKey      string `mapstructure:"VAPID_PUBLIC_KEY"`
	VAPIDPrivateKey     string `mapstructure:"VAPID_PRIVATE_KEY"`
	VAPIDSubject        string `mapstructure:"VAPID_SUBJECT"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// Environment variables take precedence over config file.
	viper.AutomaticEnv()

	// ReadInConfig is optional — in production (Docker), all values
	// come from environment variables and no config file is mounted.
	if fileErr := viper.ReadInConfig(); fileErr != nil {
		// Config file not found is OK; proceed with env vars only.
		if _, ok := fileErr.(viper.ConfigFileNotFoundError); !ok {
			// Some other file error — still try to proceed with env vars.
		}
	}

	err = viper.Unmarshal(&config)

	if config.DBDriver == "" {
		err = fmt.Errorf("DB_DRIVER is required but not set")
		return
	}
	if config.DB_URL == "" {
		err = fmt.Errorf("DB_URL is required but not set")
		return
	}
	if config.AcessTokenDuration == time.Duration(0) {
		err = fmt.Errorf("AcessTokenDuration is required but not set")
		return
	}
	if config.TokenSymmetricKey == "" {
		err = fmt.Errorf("TokenSymmetricKey is required but not set")
		return
	}

	return
}

func (c Config) CORSConfig() cors.Config {
    if c.Environment == "production" {
        return cors.Config{
            AllowOrigins:     []string{"https://chesske.com", "https://www.chesske.com"},
            AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowHeaders:     []string{"Authorization", "Content-Type"},
            AllowCredentials: true,
        }
    }

    // development — allow common Vite/React/Next ports
    return cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    }
}