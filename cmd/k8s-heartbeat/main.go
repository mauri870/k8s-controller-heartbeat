package main

import (
	"fmt"

	"github.com/caarlos0/env/v7"
	k8sheartbeat "github.com/mauri870/k8s-heartbeat"
	log "github.com/sirupsen/logrus"
	limiter "github.com/ulule/limiter/v3"

	"github.com/ulule/limiter/v3/drivers/store/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type config struct {
	KubeConfig        string `env:"KUBECONFIG" envExpand:"true"`
	LogLevel          string `env:"LOG_LEVEL" envDefault:"INFO"`
	Port              int    `env:"PORT" envDefault:"8080"`
	RateLimitDuration string `env:"RATE_LIMIT" envDefault:"3600h"`
	AuthTokenBasic    string `env:"AUTH_TOKEN_BASIC" envDefault:"xxx"`
}

func main() {
	envCfg := config{}
	if err := env.Parse(&envCfg); err != nil {
		log.Fatal(err)
	}

	lvl, err := log.ParseLevel(envCfg.LogLevel)
	if err != nil {
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)

	config, err := clientcmd.BuildConfigFromFlags("", envCfg.KubeConfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// rate limiting
	rate, err := limiter.NewRateFromFormatted(envCfg.RateLimitDuration)
	if err != err {
		log.Fatal(err)
	}
	rateLimiter := limiter.New(memory.NewStore(), rate)

	app := k8sheartbeat.NewAppHandler(clientset, envCfg.AuthTokenBasic, rateLimiter)
	log.Infof("Listening on %s", envCfg.Port)
	log.Fatal(app.Serve(fmt.Sprintf(":%d", envCfg.Port)))
}
