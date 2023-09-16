package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/personage-hub/metrics-tracker/internal/db"
	"github.com/personage-hub/metrics-tracker/internal/dumper"
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/personage-hub/metrics-tracker/internal/metrics"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricFunc(t *testing.T) {
	keeper := dumper.NewDumper("/tmp/temp.json")
	s, _ := storage.NewMemStorage(keeper, false)
	log, _ := logger.Initialize("info")
	database, _ := db.CreateAndConnect("")

	server := NewServer(s, database, log)
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		server  *Server
		method  string
		metric  metrics.Metrics
		want    want
	}{
		{
			name:    "Success Update gauge value",
			request: "/update",
			server:  server,
			method:  http.MethodPost,
			metric: metrics.Metrics{
				ID:    "someMetric",
				MType: "gauge",
				Value: func() *float64 { v := 543.01; return &v }(),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Fail Update gauge value",
			request: "/update",
			server:  server,
			method:  http.MethodPost,
			metric: metrics.Metrics{
				ID:    "someMetric",
				MType: "gauge",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Success Counter gauge value",
			request: "/update",
			server:  server,
			method:  http.MethodPost,
			metric: metrics.Metrics{
				ID:    "someMetric",
				MType: "counter",
				Delta: func() *int64 { v := int64(5456); return &v }(),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Fail Update Counter value",
			request: "/update",
			server:  server,
			method:  http.MethodPost,
			metric: metrics.Metrics{
				ID:    "someMetric",
				MType: "counter",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonMetric, _ := json.Marshal(tt.metric)
			request := httptest.NewRequest(tt.method, tt.request, bytes.NewBuffer(jsonMetric))
			response := httptest.NewRecorder()

			ctx := context.WithValue(request.Context(), chi.RouteCtxKey, chi.NewRouteContext())
			request = request.WithContext(ctx)

			server.updateMetricJSON(response, request)
			result := response.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			defer result.Body.Close()
		})
	}
}

func TestUpdateGaugeMetricStorage(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name   string
		metric string
		want   want
	}{
		{
			name:   "Success Update gauge value",
			metric: "{\n    \"id\": \"someMetric\",\n    \"type\": \"gauge\",\n    \"value\": 456.3\n}",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "Fail Update gauge value",
			metric: "{\n    \"id\": \"someMetric\",\n    \"type\": \"gauge\",\n    \"value\": \"inconvertible\"\n}",

			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := dumper.NewDumper("/tmp/temp.json")
			s, _ := storage.NewMemStorage(keeper, false)
			log, _ := logger.Initialize("info")
			database, _ := db.CreateAndConnect("")
			server := NewServer(s, database, log)
			uri := "/update/"
			request := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer([]byte(tt.metric)))
			response := httptest.NewRecorder()

			ctx := context.WithValue(request.Context(), chi.RouteCtxKey, chi.NewRouteContext())

			request = request.WithContext(ctx)

			server.updateMetricJSON(response, request)
			result := response.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)
			var m metrics.Metrics
			_ = easyjson.Unmarshal([]byte(tt.metric), &m)
			resultValue, _ := s.GetGaugeMetric(m.ID)
			assert.Equal(t, *m.Value, resultValue)
			defer result.Body.Close()
		})
	}
}

func TestUpdateCounterMetricStorage(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name   string
		metric string
		want   want
	}{
		{
			name:   "Success Update counter value",
			metric: "{\n    \"id\": \"someMetric\",\n    \"type\": \"counter\",\n    \"delta\": 456\n}",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "Fail Update counter value",
			metric: "{\n    \"id\": \"someMetric\",\n    \"type\": \"counter\",\n    \"delta\": \"inconvertible\"\n}",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := dumper.NewDumper("/tmp/temp.json")
			s, _ := storage.NewMemStorage(keeper, false)
			log, _ := logger.Initialize("info")
			database, _ := db.CreateAndConnect("")
			server := NewServer(s, database, log)
			uri := "/update/"
			request := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer([]byte(tt.metric)))
			response := httptest.NewRecorder()

			ctx := context.WithValue(request.Context(), chi.RouteCtxKey, chi.NewRouteContext())

			request = request.WithContext(ctx)

			server.updateMetricJSON(response, request)
			result := response.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)
			var m metrics.Metrics
			_ = easyjson.Unmarshal([]byte(tt.metric), &m)
			resultValue, _ := s.GetCounterMetric(m.ID)
			assert.Equal(t, *m.Delta, resultValue)
			defer result.Body.Close()
		})
	}
}
