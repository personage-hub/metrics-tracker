package main

import (
	"context"
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
		name    string
		request string
		storage *internal.MemStorage
		method  string
		want    want
	}{
		{
			name:    "Success Update gauge value",
			request: "/update/gauge/someMetric/534.4",
			storage: s,
			method:  http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Fail Update gauge value",
			request: "/update/gauge/",
			storage: s,
			method:  http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:    "Fail Update gauge Method get not allowed",
			request: "/update/gauge/someMetric/534.4",
			storage: s,
			method:  http.MethodGet,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:    "Success Counter gauge value",
			request: "/update/counter/someMetric/527",
			storage: s,
			method:  http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Fail Update Counter value",
			request: "/update/counter/",
			storage: s,
			method:  http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:    "Fail Update Counter Method get not allowed",
			request: "/update/counter/someMetric/534.4",
			storage: s,
			method:  http.MethodGet,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, nil)
			response := httptest.NewRecorder()
			updateMetric(response, request.WithContext(NewContextWithValue(request.Context(), "storage", tt.storage)))
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
		metricName  string
		metricValue string
		want        want
	}{
		{
			name:        "Success Update gauge value",
			metricName:  "someMetric",
			metricValue: "230.001",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Fail Update gauge value",
			metricName:  "someMetric",
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
			updateMetric(response, request.WithContext(NewContextWithValue(request.Context(), "storage", s)))
			result := response.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)
			resultValue, _ := s.GetGaugeMetric(tt.metricName)
			wantValue, _ := strconv.ParseFloat(tt.metricValue, 64)
			assert.Equal(t, resultValue, wantValue)
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
		metricName  string
		metricValue string
		want        want
	}{
		{
			name:        "Success Update counter value",
			metricName:  "someMetric",
			metricValue: "230",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Fail Update counter value",
			metricName:  "someMetric",
			metricValue: "inconvertible",
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
			updateMetric(response, request.WithContext(NewContextWithValue(request.Context(), "storage", s)))
			result := response.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)
			resultValue, _ := s.GetCounterMetric(tt.metricName)
			wantValue, _ := strconv.ParseInt(tt.metricValue, 10, 64)
			assert.Equal(t, resultValue, wantValue)
			defer result.Body.Close()
		})
	}
}

func NewContextWithValue(ctx context.Context, key, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}
