package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "status"},
	)

	configmapReadTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "configmap_read_total",
			Help: "Total number of configmap reads",
		},
	)
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpRequestsTotal.WithLabelValues("/", "200").Inc()
		w.Write([]byte(fmt.Sprintf("Backend running in namespace: %s\n", namespace)))
	})

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, "app-config", metav1.GetOptions{})
		if err != nil {
			httpRequestsTotal.WithLabelValues("/config", "500").Inc()
			http.Error(w, fmt.Sprintf("Failed to read ConfigMap: %v", err), http.StatusInternalServerError)
		}

		configmapReadTotal.Inc()
		httpRequestsTotal.WithLabelValues("/config", "200").Inc()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"app_name": "%s", "environment": "%s", "log_level": "%s"}`,
			cm.Data["APP_NAME"],
			cm.Data["ENVIRONMENT"],
			cm.Data["LOG_LEVEL"])
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		httpRequestsTotal.WithLabelValues("/health", "200").Inc()
		w.Write([]byte("OK"))
	})

	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Backend listenning to port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
