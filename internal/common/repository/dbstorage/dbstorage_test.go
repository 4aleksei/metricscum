package dbstorage

import (
	"context"
	"testing"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
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
		want error
		name string
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
		want error
		name string
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

func Test_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stor := mock.NewMockStore(ctrl)

	v := valuemetric.ConvertToFloatValueMetric(55.55)

	stor.EXPECT().
		Upsert(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(valuemetric.ErrBadTypeValue)

	l, _ := logger.NewLog("debug")
	db := &DBStorage{db: stor,
		l: l}

	tests := []struct {
		wantErr error
		name    string
		valName string
		value   valuemetric.ValueMetric
		want    valuemetric.ValueMetric
	}{
		{name: "Test Add", valName: "Test1", value: *v, want: *v, wantErr: valuemetric.ErrBadTypeValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := db.Add(context.Background(), tt.valName, tt.value)
			if gotErr != nil {
				if gotErr != tt.wantErr {
					t.Errorf("Error Add Ret Error = %v  , wantErr = %v", gotErr, tt.wantErr)
				}
			} else {
				if got != tt.want {
					t.Errorf("Error Add = %v  , want = %v", got, tt.want)
				}
			}
		})
	}
}

func Test_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stor := mock.NewMockStore(ctrl)

	v := valuemetric.ConvertToFloatValueMetric(55.55)

	stor.EXPECT().
		SelectValue(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(valuemetric.ErrBadTypeValue)

	l, _ := logger.NewLog("debug")
	db := &DBStorage{db: stor,
		l: l}

	tests := []struct {
		wantErr error
		name    string
		valName string
		value   valuemetric.ValueMetric
		want    valuemetric.ValueMetric
	}{
		{name: "Test Add", valName: "Test1", want: *v, wantErr: valuemetric.ErrBadTypeValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := db.Get(context.Background(), tt.valName)
			if gotErr != nil {
				if gotErr != tt.wantErr {
					t.Errorf("Error Get Ret Error = %v  , wantErr = %v", gotErr, tt.wantErr)
				}
			} else {
				if got != tt.want {
					t.Errorf("Error Get = %v  , want = %v", got, tt.want)
				}
			}
		})
	}
}

func Test_ReadAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stor := mock.NewMockStore(ctrl)

	stor.EXPECT().
		SelectValueAll(gomock.Any(), gomock.Any()).
		Return(valuemetric.ErrBadTypeValue)

	l, _ := logger.NewLog("debug")
	db := &DBStorage{db: stor,
		l: l}

	tests := []struct {
		wantErr error
		name    string
		valName string
		value   valuemetric.ValueMetric
		want    valuemetric.ValueMetric
	}{
		{name: "Test ReadAll", valName: "Test1", wantErr: valuemetric.ErrBadTypeValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := db.ReadAll(context.Background(), func(name string, val valuemetric.ValueMetric) error {
				return nil
			})

			if gotErr != nil {
				if gotErr != tt.wantErr {
					t.Errorf("Error ReadAll Ret Error = %v  , wantErr = %v", gotErr, tt.wantErr)
				}
			}
		})
	}
}

func Test_ReadAllClearCounters(t *testing.T) {
	db := &DBStorage{db: nil,
		l: nil}

	tests := []struct {
		wantErr error
		name    string
	}{
		{name: "Test ReadAllClearCounters", wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := db.ReadAllClearCounters(context.Background(), func(name string, val valuemetric.ValueMetric) error {
				return nil
			})

			if gotErr != nil {
				if gotErr != tt.wantErr {
					t.Errorf("Error ReadAllClearCounters Ret Error = %v  , wantErr = %v", gotErr, tt.wantErr)
				}
			}
		})
	}
}
