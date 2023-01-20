package sys

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/gin-contrib/cors.v1"

	"github.com/68696c6c/goat/sys/database"
	"github.com/68696c6c/goat/sys/http"
	"github.com/68696c6c/goat/sys/log"
)

type Goat struct {
	DB   database.Service
	HTTP http.Service
	Log  log.Service
}

func Init() (Goat, error) {
	config, err := readConfig()
	if err != nil {
		return Goat{}, errors.Wrapf(err, "failed to read config")
	}
	logService, err := log.NewService(config.Log)
	if err != nil {
		return Goat{}, errors.Wrapf(err, "failed to initialize log service")
	}
	return Goat{
		DB:   database.NewService(config.DB, logService),
		HTTP: http.NewService(config.HTTP, logService),
		Log:  logService,
	}, nil
}

type Config struct {
	DB   database.Config
	HTTP http.Config
	Log  log.Config
}

const (
	keyBaseUrl              = "base_url"
	keyDbDebug              = "db_debug"
	keyDbHost               = "db_host"
	keyDbPort               = "db_port"
	keyDbDatabase           = "db_database"
	keyDbUsername           = "db_username"
	keyDbPassword           = "db_password"
	keyLogLevel             = "log_level"
	keyLogStacktrace        = "log_stacktrace"
	keyHttpDebug            = "http_debug"
	keyHttpHost             = "http_host"
	keyHttpPort             = "http_port"
	keyHttpAllowOrigins     = "http_allow_origins"
	keyHttpAllowHeaders     = "http_allow_headers"
	keyHttpAllowMethods     = "http_allow_methods"
	keyHttpAllowCredentials = "http_allow_credentials"
)

func readConfig() (Config, error) {
	viper.AutomaticEnv()

	// Enable viper.Get type inferring; needed by the utils.EnvOrDefault function.
	viper.SetTypeByDefaultValue(true)

	viper.SetDefault(keyDbDebug, false)
	viper.SetDefault(keyLogLevel, "info")
	viper.SetDefault(keyLogStacktrace, false)
	viper.SetDefault(keyHttpDebug, false)
	viper.SetDefault(keyHttpHost, "")
	viper.SetDefault(keyHttpPort, "80")
	viper.SetDefault(keyHttpAllowOrigins, "*")
	viper.SetDefault(keyHttpAllowMethods, "*")
	viper.SetDefault(keyHttpAllowHeaders, "*")
	viper.SetDefault(keyHttpAllowCredentials, true)

	level, err := zap.ParseAtomicLevel(viper.GetString(keyLogLevel))
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	baseUrl, err := url.Parse(viper.GetString(keyBaseUrl))
	if err != nil {
		return Config{}, errors.Wrapf(err, "env var '%s' is not set", keyBaseUrl)
	}

	return Config{
		DB: database.Config{
			Debug:    viper.GetBool(keyDbDebug),
			Host:     viper.GetString(keyDbHost),
			Port:     viper.GetInt(keyDbPort),
			Database: viper.GetString(keyDbDatabase),
			Username: viper.GetString(keyDbUsername),
			Password: viper.GetString(keyDbPassword),
		},
		HTTP: http.Config{
			BaseUrl: baseUrl,
			Debug:   viper.GetBool(keyHttpDebug),
			Host:    viper.GetString(keyHttpHost),
			Port:    viper.GetInt(keyHttpPort),
			CORS: cors.Config{
				// AllowAllOrigins:  false,
				AllowOrigins: viper.GetStringSlice(keyHttpAllowOrigins),
				// AllowOriginFunc:  nil,
				AllowMethods:     viper.GetStringSlice(keyHttpAllowMethods),
				AllowHeaders:     viper.GetStringSlice(keyHttpAllowHeaders),
				AllowCredentials: viper.GetBool(keyHttpAllowCredentials),
				// ExposeHeaders:    nil,
				// MaxAge:           0,
			},
		},
		Log: log.Config{
			Level:      level,
			Stacktrace: viper.GetBool(keyLogStacktrace),
		},
	}, nil
}
