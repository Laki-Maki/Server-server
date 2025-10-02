package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port               string
	DatabaseURL        string
	LogLevel           string
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ShutdownTimeout    time.Duration
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetConfigType("yaml")

	v.AutomaticEnv()

	// значения по умолчанию
	v.SetDefault("PORT", "8080")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("SERVER_READ_TIMEOUT", 10)
	v.SetDefault("SERVER_WRITE_TIMEOUT", 10)
	v.SetDefault("SHUTDOWN_TIMEOUT", 10)

	_ = v.ReadInConfig() // игнорируем ошибку если файла нет

	// если задан DATABASE_URL — используем его, иначе собираем из частей
	dbURL := v.GetString("DATABASE_URL")
	if dbURL == "" {
		host := v.GetString("DB_HOST")
		port := v.GetString("DB_PORT")
		user := v.GetString("DB_USER")
		pass := v.GetString("DB_PASSWORD")
		dbname := v.GetString("DB_NAME")
		// если хоть одна часть пуста, dbURL останется пустым - вызывающий код должен заметить
		if host != "" && port != "" && user != "" && dbname != "" {
			dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				user, pass, host, port, dbname)
		}
	}

	cfg := &Config{
		Port:               v.GetString("PORT"),
		DatabaseURL:        dbURL,
		LogLevel:           v.GetString("LOG_LEVEL"),
		ServerReadTimeout:  time.Second * time.Duration(v.GetInt("SERVER_READ_TIMEOUT")),
		ServerWriteTimeout: time.Second * time.Duration(v.GetInt("SERVER_WRITE_TIMEOUT")),
		ShutdownTimeout:    time.Second * time.Duration(v.GetInt("SHUTDOWN_TIMEOUT")),
	}

	return cfg, nil
}
