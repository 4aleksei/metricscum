package dbstorage

import (
	"context"
	"testing"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/store/mock"
	"github.com/golang/mock/gomock"
)

func Test_PingContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stor := mock.NewMockStore(ctrl)

	stor.EXPECT().
		Ping(gomock.Any()).
		Return(nil)

	db := &DBStorage{db: stor,
		l: nil}

	tests := []struct {
		name string
		want error
	}{
		{name: "Test Ping", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := db.PingContext(context.Background())
			if got != tt.want {
				t.Errorf("PingContext = %v ", got)
			}
		})
	}
}

func Test_AddMulti(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stor := mock.NewMockStore(ctrl)
	var a int64 = 100
	modval := make([]models.Metrics, 0)
	modval = append(modval, models.Metrics{ID: "TEst", MType: "counter", Delta: &a})
	stor.EXPECT().
		Upserts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	db := &DBStorage{db: stor,
		l: nil}

	tests := []struct {
		name string
		want error
	}{
		{name: "Test AddMulti", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := db.AddMulti(context.Background(), modval)
			if gotErr != tt.want {
				t.Errorf("AddMulti = %v , error = %v", got, gotErr)
			}
		})
	}
}

func Test_Upserts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stor := mock.NewMockStore(ctrl)

	stor.EXPECT().
		Ping(gomock.Any()).
		Return(nil)

	db := &DBStorage{db: stor,
		l: nil}

	tests := []struct {
		name string
		want error
	}{
		{name: "Test Ping", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := db.PingContext(context.Background())
			if got != tt.want {
				t.Errorf("PingContext = %v ", got)
			}
		})
	}
}
