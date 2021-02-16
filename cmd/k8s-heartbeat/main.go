package main

import (
	"os"

	"github.com/mauri870/k8s-heartbeat"
	log "github.com/sirupsen/logrus"
	limiter "github.com/ulule/limiter/v3"

	"github.com/ulule/limiter/v3/drivers/store/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	kubeConfig     = getEnv("KUBECONFIG", "")
	logLevel       = getEnv("LOG_LEVEL", "INFO")
	port           = getEnv("PORT", "8080")
	rateLimit      = getEnv("RATE_LIMIT", "3600-H")
	authTokenBasic = getEnv("AUTH_TOKEN_BASIC", "xxx")
)

func init() {
	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)
}

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// rate limiting
	rate, err := limiter.NewRateFromFormatted(rateLimit)
	if err != err {
		log.Fatal(err)
	}
	rateLimiter := limiter.New(memory.NewStore(), rate)

	app := k8sheartbeat.NewAppHandler(clientset, authTokenBasic, rateLimiter)
	log.Infof("Listening on %s", port)
	log.Fatal(app.Serve(":" + port))
}

// getEnv gets an environment variable or a default
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}
