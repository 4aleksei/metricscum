package service

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/handlers/httpclientpool"
	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/stretchr/testify/assert"
)

func Test_NewHandlerStore(t *testing.T) {
	t.Run("Test NewHandlerStore", func(t *testing.T) {
		store := NewHandlerStore(nil, nil, nil, nil)
		assert.NotNil(t, store)
		num := store.jid
		nnum := store.newJid()
		assert.Equal(t, num+1, nnum)
	})
}

func Test_SetGauge(t *testing.T) {
	stor := memstorage.NewStore()
	serV := NewHandlerStore(stor, nil, nil, nil)

	tests := []struct {
		name      string
		valueName string
		value     float64
		wantVal   *valuemetric.ValueMetric
		wantErr   error
	}{
		{name: "Test kindIGauge64", valueName: "Test1", value: 44.44, wantVal: valuemetric.ConvertToFloatValueMetric(44.44), wantErr: nil},
		{name: "Test Overwrite", valueName: "Test1", value: 55.55, wantVal: valuemetric.ConvertToFloatValueMetric(55.55), wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := serV.SetGauge(context.Background(), tt.valueName, tt.value)
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("SetGauge = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				if got != *tt.wantVal {
					t.Errorf("SetGauge = %v, want_val %v ", got, *tt.wantVal)
				}
			}
		})
	}
}

func Test_SetCounter(t *testing.T) {
	stor := memstorage.NewStore()
	serV := NewHandlerStore(stor, nil, nil, nil)

	tests := []struct {
		name      string
		valueName string
		value     int64
		wantVal   *valuemetric.ValueMetric
		wantErr   error
	}{
		{name: "Test kindInt64", valueName: "Test1", value: 44, wantVal: valuemetric.ConvertToIntValueMetric(44), wantErr: nil},
		{name: "Test Add", valueName: "Test1", value: 55, wantVal: valuemetric.ConvertToIntValueMetric(55 + 44), wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := serV.SetCounter(context.Background(), tt.valueName, tt.value)
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("SetGauge = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				if got != *tt.wantVal {
					t.Errorf("SetGauge = %v, want_val %v ", got, *tt.wantVal)
				}
			}
		})
	}
}

func Test_SetGaugeMulti(t *testing.T) {
	stor := memstorage.NewStore()
	serV := NewHandlerStore(stor, nil, nil, nil)

	a := 100.5
	b := 100.1
	modval := make([]models.Metrics, 0)
	modval = append(modval, models.Metrics{ID: "Test1", MType: "gauge", Value: &a})
	modval = append(modval, models.Metrics{ID: "Test2", MType: "gauge", Value: &b})

	ff := make([]float64, 0)
	ff = append(ff, a)
	ff = append(ff, b)

	ss := make([]string, 0)
	ss = append(ss, "Test1")
	ss = append(ss, "Test2")

	tests := []struct {
		name      string
		valueName []string
		value     []float64
		wantVal   []models.Metrics
		wantErr   error
	}{
		{name: "Test FLOAT64 multi", valueName: ss, value: ff, wantVal: modval, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := serV.SetGaugeMulti(context.Background(), tt.valueName, tt.value)
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("SetGaugeMulti = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				assert.Equal(t, got, tt.wantVal)
			}
		})
	}
}

func Test_SendMetrics(t *testing.T) {
	cfg := &config.Config{
		Address:      "127.0.0.1:8081",
		RateLimit:    1,
		ContentJSON:  true,
		ContentBatch: 1,
	}

	l, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		_, _ = rw.Write([]byte(`OK`))
	}))
	server.Listener.Close()
	server.Listener = l
	server.Start()
	defer server.Close()

	stor := memstorage.NewStore()

	pool := httpclientpool.NewHandler(cfg)
	lo := logger.NewLogger(logger.Config{Level: "debug"})
	serV := NewHandlerStore(stor, pool, cfg, lo)

	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vI := valuemetric.ConvertToIntValueMetric(55)
	nameOfTest1 := "Test1"
	nameOfTest2 := "Test2"
	_, _ = stor.Add(context.Background(), nameOfTest1, *vI)
	_, _ = stor.Add(context.Background(), nameOfTest2, *vF)

	t.Run("TEST SendMetrics", func(t *testing.T) {
		err := serV.SendMetrics(context.Background())
		assert.Nil(t, err)
	})
}
