package prometheus

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "grafana_controller"

	TypeAlertNotification = "alert-notification"
	TypeDashboard         = "dashboard"
	TypeDataSource        = "datasource"
	TypeFolder            = "folder"
)

var (
	/*
		Controller Metrics
	*/
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

	/*
		Grafana Client Metrics
	*/
	GrafanaPostLatencyMilliseconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  namespace,
			Name:       "grafana_post_latency_ms",
			Help:       "Kubernetes Grafana Controllers Grafana Update Latency (milliseconds)",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"type"},
	)

	GrafanaPutLatencyMilliseconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  namespace,
			Name:       "grafana_put_latency_ms",
			Help:       "Kubernetes Grafana Controllers Grafana Update Latency (milliseconds)",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"type"},
	)

	GrafanaGetLatencyMilliseconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  namespace,
			Name:       "grafana_get_latency_ms",
			Help:       "Kubernetes Grafana Controllers Grafana Get Latency (milliseconds)",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"type"},
	)

	GrafanaDeleteLatencyMilliseconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  namespace,
			Name:       "grafana_delete_latency_ms",
			Help:       "Kubernetes Grafana Controllers Grafana Update Latency (milliseconds)",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"type"},
	)

	GrafanaWastedPutTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "wasted_put_total",
			Help:      "Kubernetes Grafana Controllers Grafana Wasted Put Total",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(ErrorTotal)
	prometheus.MustRegister(DeletedObjectTotal)
	prometheus.MustRegister(UpdatedObjectTotal)
	prometheus.MustRegister(ResyncDeletedTotal)

	prometheus.MustRegister(GrafanaPostLatencyMilliseconds)
	prometheus.MustRegister(GrafanaPutLatencyMilliseconds)
	prometheus.MustRegister(GrafanaDeleteLatencyMilliseconds)
	prometheus.MustRegister(GrafanaGetLatencyMilliseconds)
	prometheus.MustRegister(GrafanaWastedPutTotal)
}
