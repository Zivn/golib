package http

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
	"time"
)

var (
	// DefaultBuckets prometheus buckets in seconds.
	DefaultBuckets = []float64{0.3, 1.2, 5.0}
)

const (
	reqsName    = "http_requests_total"
	latencyName = "http_request_duration_seconds"
)

// Prometheus is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
//
// Usage: pass its `ServeHTTP` to a route or globally.
type Prometheus struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// New returns a new prometheus middleware.
//
// If buckets are empty then `DefaultBuckets` are set.
func New(name, env string, buckets ...float64) *Prometheus {
	p := Prometheus{}
	p.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name, "env": env},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(p.reqs)

	if len(buckets) == 0 {
		buckets = DefaultBuckets
	}

	p.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": name, "env": env},
		Buckets:     buckets,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(p.latency)

	return &p
}

func (p *Prometheus) ServeHTTP(ctx context.Context) {
	start := time.Now()
	ctx.Next()
	r := ctx.Request()
	statusCode := strconv.Itoa(ctx.GetStatusCode())

	p.reqs.WithLabelValues(statusCode, r.Method, r.URL.Path).
		Inc()

	p.latency.WithLabelValues(statusCode, r.Method, r.URL.Path).
		Observe(float64(time.Since(start).Nanoseconds()) / 1000000000)
}

func (p *Prometheus) MetricsHandler() context.Handler {
	return iris.FromStd(promhttp.Handler())
}
