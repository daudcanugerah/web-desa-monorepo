package cmd

import (
	"webdesa/api/config"
	"webdesa/api/system"
	"context"
	"fmt"
	"net/http"
	"os"
	"shared/otel"
	debug "shared/profiler"
	"time"

	"github.com/spf13/cobra"
)

var (
	ctx               context.Context = context.Background()
	mainOtel          otel.Otel
	shudtdownFuncList = []func(ctx context.Context) error{}
	systemConfig      *config.Config
)

var (
	defaultHttpClient = &http.Client{Timeout: time.Second * 60}
	proxyClient       = &http.Client{Timeout: time.Second * 60}
)

var (
	cfgFileInput        string
	noProxyInput        bool
	profileModeInput    []string
	logLevelInput       string
	profileAddressInput string
)

var rootCmd = &cobra.Command{
	Use:   "desa-api",
	Short: "Desa API - Village Information System",
	Long:  "REST API for managing village information with RBAC authorization",
}

func Initialize(ctx context.Context) error {
	var err error

	if systemConfig, err = config.InitConfig(cfgFileInput); err != nil {
		return err
	}

	// setup otel
	otelConfig := otel.SetupOption{
		EnableMetric:   systemConfig.Otel.EnableMetric,
		EnableTrace:    systemConfig.Otel.EnableTrace,
		EnableLog:      systemConfig.Otel.EnableLog,
		ServiceName:    system.APP_NAME,
		ServiceVersion: system.APP_VERSION,
	}

	shutdownFunc, err := otel.SetupOTelSDK(ctx, otelConfig)
	if err != nil {
		return err
	}

	mainOtel = otel.NewOtel("main", logLevelInput)
	shudtdownFuncList = append(shudtdownFuncList, shutdownFunc)

	// setup profiler
	systemProfiler := debug.SetProfiler(profileModeInput)
	shudtdownFuncList = append(shudtdownFuncList, func(ctx context.Context) error {
		for _, v := range systemProfiler {
			v.Stop()
		}
		return nil
	})

	// setup timezone
	if err := config.SetUpTimezone(systemConfig.App.Timezone); err != nil {
		return err
	}

	// setup proxy
	if !noProxyInput && systemConfig.App.Proxy != "" {
		proxyClient, err = config.SetProxy(systemConfig.App.Proxy)
		if err != nil {
			return err
		}
	}

	return nil
}

func Execute() int {
	defer func() {
		for _, v := range shudtdownFuncList {
			v(ctx)
		}
	}()

	cobra.OnInitialize(func() {
		if err := Initialize(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "%+v", err)
			os.Exit(1)
		}
	})

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		if !mainOtel.IsLogAvailable() {
			fmt.Fprintf(os.Stderr, "error when running command %+v", err)
			return 1
		}

		mainOtel.Log.Error(ctx, fmt.Sprintf("app exit error: %s", err), "stacktrace", fmt.Errorf("%+v", err))
		return 1
	}

	return 0
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFileInput, "config", "", "", "config file (default is $HOME/config.toml or ./config.toml)")
	rootCmd.PersistentFlags().StringSliceVarP(&profileModeInput, "run-profile", "", []string{}, "run app profile using pprof, available mode cpu,ram,mutex,block,gorountine,trace (if available)")
	rootCmd.PersistentFlags().BoolVarP(&noProxyInput, "no-proxy", "", false, "not use proxy")
	rootCmd.PersistentFlags().StringVarP(&profileAddressInput, "run-realtime-profile", "", "", "http address for running realtime pprof")
	rootCmd.PersistentFlags().StringVarP(&logLevelInput, "log-level", "", "info", "log level, available level debug,info,warn,error")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
