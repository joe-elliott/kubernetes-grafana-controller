package prometheus

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "grafana_controller"
)

var (
	ErrorTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "error_total",
			Help:      "Kubernetes Grafana Controllers Errors",
		},
	)

	DeletedObjectTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "deleted_object_total",
			Help:      "Kubernetes Grafana Controllers Deleted Objects Counter",
		},
		[]string{"type"},
	)

	UpdatedObjectTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "updated_object_total",
			Help:      "Kubernetes Grafana Controllers Updated Objects Counter",
		},
		[]string{"type"},
	)

	ResyncDeletedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "resynced_deleted_total",
			Help:      "Kubernetes Grafana Controllers Resync Deleted Objects Counter",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(ErrorTotal)
	prometheus.MustRegister(DeletedObjectTotal)
	prometheus.MustRegister(UpdatedObjectTotal)
	prometheus.MustRegister(ResyncDeletedTotal)
}
