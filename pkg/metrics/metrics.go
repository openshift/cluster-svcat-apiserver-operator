package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

var (
	buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "openshift_cluster_openshift_apiserver_operator_build_info",
			Help: "A metric with a constant '1' value labeled by major, minor, git commit & git version from which OpenShift API Server was built.",
		},
		[]string{"major", "minor", "gitCommit", "gitVersion"},
	)

	apiserverEnabled = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_catalog_apiserver_enabled",
			Help: "Indicates whether Service Catalog apiserver is enabled",
		})
)

func init() {
	// do the MustRegister here
	prometheus.MustRegister(buildInfo)
	prometheus.MustRegister(apiserverEnabled)
}

// We will never want to panic our operator because of metric saving.
// Therefore, we will recover our panics here and error log them
// for later diagnosis but will never fail the operator.
func recoverMetricPanic() {
	if r := recover(); r != nil {
		klog.Errorf("Recovering from metric function - %v", r)
	}
}

func RegisterVersion(major, minor, gitCommit, gitVersion string) {
	defer recoverMetricPanic()
	buildInfo.WithLabelValues(major, minor, gitCommit, gitVersion).Set(1)
}

// APIServerEnabled - Indicates Service Catalog APIServer has been enabled
func APIServerEnabled() {
	defer recoverMetricPanic()
	apiserverEnabled.Inc()
}

// APIServerDisabled - Indicates Service Catalog APIServer has been disabled
func APIServerDisabled() {
	defer recoverMetricPanic()
	apiserverEnabled.Dec()
}
