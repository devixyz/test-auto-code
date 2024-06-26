package config

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	DBHost         string `mapstructure:"POSTGRES_HOST"`
	DBUserName     string `mapstructure:"POSTGRES_USER"`
	DBUserPassword string `mapstructure:"POSTGRES_PASSWORD"`
	DBName         string `mapstructure:"POSTGRES_DB"`
	DBPort         string `mapstructure:"POSTGRES_PORT"`
	ServerPort     string `mapstructure:"PORT"`
	Mode           string `mapstructure:"MODE"`

	ClientOrigin string `mapstructure:"CLIENT_ORIGIN"`
	Domain       string `mapstructure:"DOMAIN"`

	TokenSecret    string        `mapstructure:"TOKEN_SECRET"`
	TokenExpiresIn time.Duration `mapstructure:"TOKEN_EXPIRED_IN"`
	TokenMaxAge    int           `mapstructure:"TOKEN_MAXAGE"`

	EmailFrom string `mapstructure:"EMAIL_FROM"`
	SMTPHost  string `mapstructure:"SMTP_HOST"`
	SMTPPass  string `mapstructure:"SMTP_PASS"`
	SMTPPort  int    `mapstructure:"SMTP_PORT"`
	SMTPUser  string `mapstructure:"SMTP_USER"`

	AccessTokenPrivateKey  string        `mapstructure:"ACCESS_TOKEN_PRIVATE_KEY"`
	AccessTokenPublicKey   string        `mapstructure:"ACCESS_TOKEN_PUBLIC_KEY"`
	RefreshTokenPrivateKey string        `mapstructure:"REFRESH_TOKEN_PRIVATE_KEY"`
	RefreshTokenPublicKey  string        `mapstructure:"REFRESH_TOKEN_PUBLIC_KEY"`
	AccessTokenExpiresIn   time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRED_IN"`
	RefreshTokenExpiresIn  time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRED_IN"`
	AccessTokenMaxAge      int           `mapstructure:"ACCESS_TOKEN_MAXAGE"`
	RefreshTokenMaxAge     int           `mapstructure:"REFRESH_TOKEN_MAXAGE"`

	// gpt apikey
	ApiKey        string   `mapstructure:"API_KEY"`
	ApiURL        string   `mapstructure:"API_URL"`
	Listen        string   `mapstructure:"LISTEN"`
	Proxy         string   `mapstructure:"PROXY"`
	AdminEmail    []string `mapstructure:"ADMIN_EMAIL"`
	AdminPassword string   `mapstructure:"ADMIN_PASSWORD"`

	// xminio apikey
	MinioAccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	MinioBucketUrl string `mapstructure:"MINIO_BUCKET_URL"`
	MinioBucket    string `mapstructure:"MINIO_BUCKET"`

	// redis config
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPassword string `mapstructure:"REDIS_PWD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
}

var Env Config

func LoadEnv(path string) error {
	var err error

	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigType("yaml")
		viper.SetConfigName("settings-dev")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&Env)
	return nil
}
