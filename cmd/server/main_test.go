package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/personage-hub/metrics-tracker/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestUpdateMetricFunc(t *testing.T) {
	s := internal.NewMemStorage()
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		request     string
		storage     *internal.MemStorage
		method      string
		metricType  MetricType
		metricName  string
		metricValue string
		want        want
	}{
		{
			name:        "Success Update gauge value",
			request:     "/update",
			storage:     s,
			method:      http.MethodPost,
			metricType:  MetricType("gauge"),
			metricName:  "someMetric",
			metricValue: "543.0",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Fail Update gauge value",
			request:     "/update",
			storage:     s,
			method:      http.MethodPost,
			metricType:  MetricType("gauge"),
			metricName:  "",
			metricValue: "",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:        "Fail Update gauge Method get not allowed",
			request:     "/update",
			storage:     s,
			metricType:  MetricType("gauge"),
			metricName:  "someMetric",
			metricValue: "543.0",
			method:      http.MethodGet,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:        "Success Counter gauge value",
			request:     "/update",
			storage:     s,
			method:      http.MethodPost,
			metricType:  MetricType("counter"),
			metricName:  "someMetric",
			metricValue: "527",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Fail Update Counter value",
			request:     "/update",
			storage:     s,
			metricType:  MetricType("counter"),
			metricName:  "",
			metricValue: "",
			method:      http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:        "Fail Update Counter Method get not allowed",
			request:     "/update",
			storage:     s,
			metricType:  MetricType("counter"),
			metricName:  "someMetric",
			metricValue: "527",
			method:      http.MethodGet,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, nil)
			response := httptest.NewRecorder()

			ctx := context.WithValue(request.Context(), chi.RouteCtxKey, chi.NewRouteContext())
			routeContext := chi.RouteContext(ctx)
			routeContext.URLParams.Add("metricType", string(tt.metricType))
			routeContext.URLParams.Add("metricName", tt.metricName)
			routeContext.URLParams.Add("metricValue", tt.metricValue)
			request = request.WithContext(ctx)

			updateMetric(response, request.WithContext(NewContextWithValue(request.Context(), storageKeyContextKey, tt.storage)))
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
		name        string
		metricType  MetricType
		metricName  string
		metricValue string
		want        want
	}{
		{
			name:        "Success Update gauge value",
			metricName:  "someMetric",
			metricType:  MetricType("gauge"),
			metricValue: "230.001",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Fail Update gauge value",
			metricName:  "someMetric",
			metricType:  MetricType("gauge"),
			metricValue: "inconvertible",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := internal.NewMemStorage()
			uri := "/update/gauge/" + tt.metricName + "/" + tt.metricValue
			request := httptest.NewRequest(http.MethodPost, uri, nil)
			response := httptest.NewRecorder()

			ctx := context.WithValue(request.Context(), chi.RouteCtxKey, chi.NewRouteContext())
			routeContext := chi.RouteContext(ctx)
			routeContext.URLParams.Add("metricType", string(tt.metricType))
			routeContext.URLParams.Add("metricName", tt.metricName)
			routeContext.URLParams.Add("metricValue", tt.metricValue)
			request = request.WithContext(ctx)

			updateMetric(response, request.WithContext(NewContextWithValue(request.Context(), storageKeyContextKey, s)))
			result := response.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)
			resultValue, _ := s.GetGaugeMetric(tt.metricName)
			wantValue, _ := strconv.ParseFloat(tt.metricValue, 64)
			assert.Equal(t, wantValue, resultValue)
			defer result.Body.Close()
		})
	}
}

func TestUpdateCounterMetricStorage(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		metricType  MetricType
		metricName  string
		metricValue string
		want        want
	}{
		{
			name:        "Success Update counter value",
			metricName:  "someMetric",
			metricType:  MetricType("counter"),
			metricValue: "230",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Fail Update counter value",
			metricName:  "someMetric",
			metricValue: "inconvertible",
			metricType:  MetricType("counter"),
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := internal.NewMemStorage()
			uri := "/update/counter/" + tt.metricName + "/" + tt.metricValue
			request := httptest.NewRequest(http.MethodPost, uri, nil)
			response := httptest.NewRecorder()

			ctx := context.WithValue(request.Context(), chi.RouteCtxKey, chi.NewRouteContext())
			routeContext := chi.RouteContext(ctx)
			routeContext.URLParams.Add("metricType", string(tt.metricType))
			routeContext.URLParams.Add("metricName", tt.metricName)
			routeContext.URLParams.Add("metricValue", tt.metricValue)
			request = request.WithContext(ctx)

			updateMetric(response, request.WithContext(NewContextWithValue(request.Context(), storageKeyContextKey, s)))
			result := response.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)
			resultValue, _ := s.GetCounterMetric(tt.metricName)
			wantValue, _ := strconv.ParseInt(tt.metricValue, 10, 64)
			assert.Equal(t, wantValue, resultValue)
			defer result.Body.Close()
		})
	}
}

func NewContextWithValue(ctx context.Context, key storageKey, value *internal.MemStorage) context.Context {
	return context.WithValue(ctx, key, value)
}
