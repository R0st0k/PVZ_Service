package metrics

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

var (
	metricsOnce sync.Once
	testMetrics *Metrics
)

func TestResponseWriter(t *testing.T) {
	t.Run("WriteHeader sets status", func(t *testing.T) {
		rw := &responseWriter{
			ResponseWriter: httptest.NewRecorder(),
			status:         0,
		}

		rw.WriteHeader(http.StatusNotFound)
		assert.Equal(t, http.StatusNotFound, rw.status)
	})

	t.Run("Write sets default status if not set", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: rec,
			status:         0,
		}

		_, err := rw.Write([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rw.status)
		assert.Equal(t, "test", rec.Body.String())
	})

	t.Run("Write preserves custom status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: rec,
			status:         http.StatusCreated,
		}

		_, err := rw.Write([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rw.status)
	})
}

func TestNewMetrics(t *testing.T) {
	metricsOnce.Do(func() {
		testMetrics = NewMetrics()
	})

	tests := []struct {
		name     string
		metric   interface{}
		expected prometheus.Collector
	}{
		{
			name:     "HTTPRequestsTotal",
			metric:   testMetrics.HTTPRequestsTotal,
			expected: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total"}, []string{"method", "path", "status"}),
		},
		{
			name:     "HTTPResponseTime",
			metric:   testMetrics.HTTPResponseTime,
			expected: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_response_time_seconds"}, []string{"method", "path"}),
		},
		{
			name:     "PVZCreated",
			metric:   testMetrics.PVZCreated,
			expected: prometheus.NewCounter(prometheus.CounterOpts{Name: "pvz_created_total"}),
		},
		{
			name:     "ReceptionsCreated",
			metric:   testMetrics.ReceptionsCreated,
			expected: prometheus.NewCounter(prometheus.CounterOpts{Name: "receptions_created_total"}),
		},
		{
			name:     "ProductsAdded",
			metric:   testMetrics.ProductsAdded,
			expected: prometheus.NewCounter(prometheus.CounterOpts{Name: "products_added_total"}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.IsType(t, tt.expected, tt.metric)
		})
	}
}

func TestMiddleware(t *testing.T) {
	metricsOnce.Do(func() {
		testMetrics = NewMetrics()
	})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("records metrics for successful request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		middleware := testMetrics.Middleware(handler)
		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())

		counter, err := testMetrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/test", "200")
		assert.NoError(t, err)
		assert.NotNil(t, counter)

		histogram, err := testMetrics.HTTPResponseTime.GetMetricWithLabelValues("GET", "/test")
		assert.NoError(t, err)
		assert.NotNil(t, histogram)
	})

	t.Run("records metrics for error response", func(t *testing.T) {
		errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		})

		req := httptest.NewRequest("GET", "/not-found", nil)
		rec := httptest.NewRecorder()

		middleware := testMetrics.Middleware(errorHandler)
		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		counter, err := testMetrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/not-found", "404")
		assert.NoError(t, err)
		assert.NotNil(t, counter)
	})

	t.Run("records response time", func(t *testing.T) {
		slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/slow", nil)
		rec := httptest.NewRecorder()

		start := time.Now()
		middleware := testMetrics.Middleware(slowHandler)
		middleware.ServeHTTP(rec, req)
		duration := time.Since(start)

		assert.GreaterOrEqual(t, duration.Seconds(), 0.1)
	})
}

func TestBusinessMetrics(t *testing.T) {
	metricsOnce.Do(func() {
		testMetrics = NewMetrics()
	})

	t.Run("PVZCreated can be incremented", func(t *testing.T) {
		before := testutil.ToFloat64(testMetrics.PVZCreated)
		testMetrics.PVZCreated.Inc()
		after := testutil.ToFloat64(testMetrics.PVZCreated)
		assert.Equal(t, before+1, after)
	})

	t.Run("ReceptionsCreated can be incremented", func(t *testing.T) {
		before := testutil.ToFloat64(testMetrics.ReceptionsCreated)
		testMetrics.ReceptionsCreated.Inc()
		after := testutil.ToFloat64(testMetrics.ReceptionsCreated)
		assert.Equal(t, before+1, after)
	})

	t.Run("ProductsAdded can be incremented", func(t *testing.T) {
		before := testutil.ToFloat64(testMetrics.ProductsAdded)
		testMetrics.ProductsAdded.Inc()
		after := testutil.ToFloat64(testMetrics.ProductsAdded)
		assert.Equal(t, before+1, after)
	})
}
