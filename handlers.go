package k8sheartbeat

import (
	"context"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	limiter "github.com/ulule/limiter/v3"
	"k8s.io/client-go/kubernetes"
)

// AppHandler contains the route handlers for the application
type AppHandler struct {
	clientset      *kubernetes.Clientset
	authTokenBasic string
	rateLimiter    *limiter.Limiter
}

func NewAppHandler(client *kubernetes.Clientset, authToken string, rateLimiter *limiter.Limiter) *AppHandler {
	return &AppHandler{clientset: client, authTokenBasic: authToken, rateLimiter: rateLimiter}
}

func (app *AppHandler) Serve(addr string) error {
	r := mux.NewRouter()
	r.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	// Health Check endpoint
	r.HandleFunc("/healthz", app.health).Methods("GET", "HEAD")

	// Deployment health check
	s := r.PathPrefix("/api").Methods("GET", "HEAD").Subrouter()
	s.Use(rateLimitMiddleware(app.rateLimiter))
	s.Use(authBasicMiddleware(app.authTokenBasic))
	s.Handle("/healthz/{namespace}/deployment/{component}", app.componentHandler())

	return http.ListenAndServe(addr, r)
}

func (app *AppHandler) healthcheckerForNamespace(namespace string) HealthChecker {
	return NewK8sHealthChecker(namespace, app.clientset)
}

func (app *AppHandler) componentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		component := vars["component"]

		health := app.healthcheckerForNamespace(namespace)
		err := health.HealthCheck(context.Background(), component)
		if err != nil {
			log.Errorf("Health check failed for %s: %s", component, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func (app *AppHandler) health(w http.ResponseWriter, r *http.Request) {
	health := app.healthcheckerForNamespace("default")
	if err := health.Ping(); err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
