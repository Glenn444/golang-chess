package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver           string        `mapstructure:"DB_DRIVER"`
	DB_URL             string        `mapstructure:"DB_URL"`
	ServerAddress      string        `mapstructure:"SERVER_ADDRESS"`
	AcessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	TokenSymmetricKey  string        `mapstructure:"TokenSymmetricKey"`
	RESEND_API_KEY     string        `mapstructure:"RESEND_API_KEY"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path) //name of config file (without extension)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

    // catch missing required fields
    if config.DBDriver == "" {
        err = fmt.Errorf("DB_DRIVER is required but not set")
        return
    }
    if config.DB_URL == "" {
        err = fmt.Errorf("DB_URL is required but not set")
        return
    }
	
	if config.AcessTokenDuration == time.Duration(0){
        err = fmt.Errorf("AcessTokenDuration is required but not set")
        return
    }
	if config.TokenSymmetricKey == ""{
		err = fmt.Errorf("TokenSymmetricKey is required but not set")
		return
	}

	return
}
