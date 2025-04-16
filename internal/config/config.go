package config

import (
	"log/slog"
	"os"
	"reflect"

	"github.com/spf13/viper"
)

type Config struct {
	Viper *viper.Viper
	Env   *configVar
}

type Env string

const (
	PROD    Env = "production"
	STAGING Env = "staging"
	DEV     Env = "development"
)

var ENV = DEV

type configVar struct {
	AppEnv      Env    `mapstructure:"APP_ENV"`
	AppURL      string `mapstructure:"APP_URL"`
	Port        int    `mapstructure:"PORT"`
	Host        string `mapstructure:"HOST"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	DBDatabase  string `mapstructure:"DATABASE_DB"`
	DBPassword  string `mapstructure:"DATABASE_PASSWORD"`
	DBUsername  string `mapstructure:"DATABASE_USERNAME"`
	DBPort      int    `mapstructure:"DATABASE_PORT"`
	DBHost      string `mapstructure:"DATABASE_HOST"`
	ClerkSDKKey string `mapstructure:"CLERK_SDK_KEY"`
}

func NewConfig() *Config {
	var v = viper.New()
	var e configVar

	v.AutomaticEnv()

	t := reflect.TypeOf(e)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		v.MustBindEnv(tag)
	}

	v.SetConfigFile(".env")
	if _, err := os.Stat(v.ConfigFileUsed()); os.IsNotExist(err) {
		slog.Warn("Config file not found, using environment variables only")
	} else {
		err := v.ReadInConfig()
		if err != nil {
			slog.Error("Error reading config file",
				"config-file", v.ConfigFileUsed(),
				"error", err)
		}
	}

	setDefaults(v)

	err := v.Unmarshal(&e)
	if err != nil {
		slog.Error("Config file can't be loaded", "error", err)
		os.Exit(1)
	}

	if e.AppEnv != "" {
		ENV = e.AppEnv
	}

	return &Config{
		Viper: v,
		Env:   &e,
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("PORT", "3000")
	v.SetDefault("DATABASE_PORT", "5432")
	v.SetDefault("DATABASE_HOST", "localhost")
	v.SetDefault("DATABASE_DB", "piscineDB")
	v.SetDefault("DATABASE_USERNAME", "postgres")
	v.SetDefault("DATABASE_URL", "postgres://postgres@localhost:5432/piscineDB?sslmode=disable")
}
