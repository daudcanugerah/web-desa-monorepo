package config

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// AppConfig holds application configuration
type AppConfig struct {
	Name         string        `mapstructure:"name"`
	Version      string        `mapstructure:"version"`
	HTTPAddr     string        `mapstructure:"http_addr"`
	DomainAddr   string        `mapstructure:"domain_addr"`
	Timezone     string        `mapstructure:"timezone"`
	Proxy        string        `mapstructure:"proxy_url"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// Config holds all configuration
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Postgres   PostgresConfig   `mapstructure:"postgres"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Backup     BackupConfig     `mapstructure:"backup"`
	Otel       OtelConfig       `mapstructure:"otel"`
	Debug      DebugConfig      `mapstructure:"debug"`
	CORS       CORSConfig       `mapstructure:"cors"`
	RateLimit  RateLimitConfig  `mapstructure:"ratelimit"`
	FileUpload FileUploadConfig `mapstructure:"fileupload"`
	Metabase   MetabaseConfig   `mapstructure:"metabase"`
	SMTP       SMTPConfig       `mapstructure:"smtp"`
	Gallery    GalleryConfig    `mapstructure:"gallery"`
}

// GetProfilerAddress returns profiler address
func (c *Config) GetProfilerAddress() string {
	if c.Debug.ProfilerAddress == "" {
		return "localhost:21780"
	}
	return c.Debug.ProfilerAddress
}

// SetUpTimezone sets up timezone
func SetUpTimezone(tz string) error {
	if tz != "" {
		var err error
		time.Local, err = time.LoadLocation(tz)
		if err != nil {
			return errtrace.Wrap(fmt.Errorf("error loading location '%s': %v", tz, err))
		}
	}

	return nil
}

// SetProxy sets up HTTP proxy
func SetProxy(proxy string) (*http.Client, error) {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %v", err)
	}

	httpClient := &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return httpClient, nil
}

// InitConfig initializes configuration
func InitConfig(cfgFile string) (*Config, error) {
	var config Config

	viper.SetConfigType("toml")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return nil, errtrace.Wrap(fmt.Errorf("find home dir error: %w", err))
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("read init config error: %w", err))
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())

	if err := viper.Unmarshal(&config); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("parse config error: %w", err))
	}

	return &config, nil
}
