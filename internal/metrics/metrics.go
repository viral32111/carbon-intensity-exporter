package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	METRICS_NAMESPACE = "carbon"
)

var (
	REGION_LABELS         = []string{"region", "operator"}
	GENERATION_MIX_LABELS = []string{"fuel"}
	INTENSITY_LABELS      = []string{"index"}
)

var (
	GenerationMix     = NewGauge("generation_mix", "", append(REGION_LABELS, GENERATION_MIX_LABELS...))
	IntensityForecast = NewGauge("intensity_forecast", "", append(REGION_LABELS, INTENSITY_LABELS...))

	PeriodFrom = NewMonotonicGauge("period_from", "", REGION_LABELS)
	PeriodTo   = NewMonotonicGauge("period_to", "", REGION_LABELS)
)

func NewGauge(name string, help string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_%s", METRICS_NAMESPACE, name),
		Help: help,
	}, labels)
}

func Register(registry *prometheus.Registry) {
	collectors := []prometheus.Collector{
		GenerationMix,
		IntensityForecast,
		PeriodFrom.Collector(),
		PeriodTo.Collector(),
	}

	for _, collector := range collectors {
		registry.MustRegister(collector)
	}
}
