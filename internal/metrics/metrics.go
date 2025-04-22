package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// responseWriter - обертка для http.ResponseWriter, которая сохраняет статус ответа
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader переопределяет метод WriteHeader для сохранения статуса
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write переопределяет метод Write для совместимости
func (rw *responseWriter) Write(b []byte) (int, error) {
	// Если статус не был установлен, считаем его 200 OK
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

type Metrics struct {
	// Технические метрики
	HTTPRequestsTotal *prometheus.CounterVec
	HTTPResponseTime  *prometheus.HistogramVec

	// Бизнесовые метрики
	PVZCreated        prometheus.Counter
	ReceptionsCreated prometheus.Counter
	ProductsAdded     prometheus.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPResponseTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_time_seconds",
				Help:    "HTTP response time in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5},
			},
			[]string{"method", "path"},
		),
		PVZCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "pvz_created_total",
				Help: "Total number of PVZ created",
			},
		),
		ReceptionsCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "receptions_created_total",
				Help: "Total number of receptions created",
			},
		),
		ProductsAdded: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "products_added_total",
				Help: "Total number of products added",
			},
		),
	}
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем обертку для ResponseWriter для получения статуса ответа
		rw := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		// Записываем метрики
		m.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(rw.status),
		).Inc()

		m.HTTPResponseTime.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)
	})
}
