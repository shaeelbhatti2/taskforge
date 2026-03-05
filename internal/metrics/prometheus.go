package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Collector struct {
	jobRunsTotal   *prometheus.CounterVec
	jobRunDuration *prometheus.HistogramVec
	schedulerLag   prometheus.Gauge
	workerCount    *prometheus.GaugeVec
	dlqSize        prometheus.Gauge
}

func NewCollector() *Collector {
	return &Collector{
		jobRunsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "job_run_total",
			Help: "Total job runs by status",
		}, []string{"status"}),
		jobRunDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "job_run_duration_seconds",
			Help:    "Job run duration",
			Buckets: prometheus.DefBuckets,
		}, []string{"job_id"}),
		schedulerLag: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scheduler_tick_lag_seconds",
			Help: "Scheduler tick lag",
		}),
		workerCount: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "worker_count",
			Help: "Workers by status",
		}, []string{"status"}),
		dlqSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "dlq_size",
			Help: "Dead letter queue size",
		}),
	}
}

func (c *Collector) IncRun(status string) {
	c.jobRunsTotal.WithLabelValues(status).Inc()
}

func (c *Collector) ObserveDuration(jobID string, d time.Duration) {
	c.jobRunDuration.WithLabelValues(jobID).Observe(d.Seconds())
}

func (c *Collector) SetSchedulerLag(sec float64) {
	c.schedulerLag.Set(sec)
}

func (c *Collector) SetWorkers(status string, n float64) {
	c.workerCount.WithLabelValues(status).Set(n)
}

func (c *Collector) SetDLQSize(n float64) {
	c.dlqSize.Set(n)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
