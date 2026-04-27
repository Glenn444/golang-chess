package config


import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver           string        `mapstructure:"DB_DRIVER"`
	DB_URL             string        `mapstructure:"DB_URL"`
	ServerAddress      string        `mapstructure:"SERVER_ADDRESS"`
	AcessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	TokenSymmetricKey  string        `mapstructure:"TokenSymmetricKey"`
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
	return
}
