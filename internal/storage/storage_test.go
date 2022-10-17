package storage

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
)

func TestRepository_InsertMetric(t *testing.T) {
	type args struct {
		ctx context.Context
		m   models.Metrics
	}
	tests := []struct {
		name    string
		r       *Repository
		args    args
		wantErr bool
	}{
		{
			name: "Pos #1",
			r: &Repository{
				JSONDB: []models.Metrics{},
			},
			args: args{
				ctx: context.TODO(),
				m:   models.NewMetric("test", "counter", pointy.Int64(10), nil, ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.InsertMetric(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("Repository.InsertMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.m != tt.r.JSONDB[0] {
				t.Error("Repository.InsertMetric() error ")
			}
		})
	}
}

func TestRepository_GetMetric(t *testing.T) {
	type args struct {
		data models.Metrics
	}
	tests := []struct {
		name    string
		r       *Repository
		args    args
		want    models.Metrics
		wantErr bool
	}{
		{
			name: "pos #1",
			r: &Repository{
				JSONDB: []models.Metrics{
					{
						ID:    "test",
						MType: "counter",
						Delta: pointy.Int64(123),
					},
				},
			},
			args: args{
				data: models.Metrics{
					ID:    "test",
					MType: "counter",
				},
			},
			want: models.Metrics{
				ID:    "test",
				MType: "counter",
				Delta: pointy.Int64(123),
			},
			wantErr: false,
		},
		{
			name: "pos #2",
			r: &Repository{
				JSONDB: []models.Metrics{
					{
						ID:    "test",
						MType: "counter",
						Delta: pointy.Int64(123),
					},
				},
			},
			args: args{
				data: models.Metrics{
					ID:    "test1",
					MType: "counter",
				},
			},
			want: models.Metrics{
				ID:    "test1",
				MType: "counter",
				Delta: new(int64),
				Value: pointy.Float64(0.0),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.GetMetric(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.GetMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_InsertData(t *testing.T) {
	type args struct {
		ctx     context.Context
		typeVar string
		name    string
		value   string
		hash    string
	}
	tests := []struct {
		name string
		r    *Repository
		args args
		want int
	}{
		{
			name: "Pos counter #1",
			r: &Repository{
				JSONDB: []models.Metrics{},
			},
			args: args{
				ctx:     context.TODO(),
				typeVar: "counter",
				name:    "test",
				value:   "10",
				hash:    "",
			},
			want: 200,
		},
		{
			name: "Pos gauge #1",
			r: &Repository{
				JSONDB: []models.Metrics{},
			},
			args: args{
				ctx:     context.TODO(),
				typeVar: "gauge",
				name:    "test",
				value:   "10.01",
				hash:    "",
			},
			want: 200,
		},
		{
			name: "Neg counter #1",
			r: &Repository{
				JSONDB: []models.Metrics{},
			},
			args: args{
				ctx:     context.TODO(),
				typeVar: "counter",
				name:    "test",
				value:   "not_a_number",
				hash:    "",
			},
			want: 400,
		},
		{
			name: "Neg gauge #1",
			r: &Repository{
				JSONDB: []models.Metrics{},
			},
			args: args{
				ctx:     context.TODO(),
				typeVar: "gauge",
				name:    "test",
				value:   "not_a_number",
				hash:    "",
			},
			want: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.InsertData(tt.args.ctx, tt.args.typeVar, tt.args.name, tt.args.value, tt.args.hash); got != tt.want {
				t.Errorf("Repository.InsertData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_AppendMetric(t *testing.T) {
	type args struct {
		m models.Metrics
	}
	tests := []struct {
		name string
		r    *Repository
		args args
	}{
		{
			name: "pos1",
			r: &Repository{
				JSONDB: []models.Metrics{
					models.NewMetric("test", "counter", pointy.Int64(10), nil, ""),
				},
			},
			args: args{
				models.NewMetric("test", "counter", pointy.Int64(10), nil, ""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.AppendMetric(tt.args.m)
		})
	}
}

func TestRepository_SaveData(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name string
		r    *Repository
		args args
	}{
		{
			name: "pos1",
			r: &Repository{
				JSONDB: []models.Metrics{
					models.NewMetric("test", "counter", pointy.Int64(10), nil, ""),
				},
			},
			args: args{
				"test.db",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldDB := tt.r.GetAll(context.TODO())
			tt.r.SaveData(tt.args.file)
			tt.r.Restore(tt.args.file)
			assert.Equal(t, tt.r.JSONDB, oldDB)
			os.Remove("test.db")
		})
	}
}

func TestRepository_PingDB(t *testing.T) {
	tests := []struct {
		name string
		r    Repository
		want bool
	}{
		{
			name: "pos",
			r:    Repository{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.PingDB(); got != tt.want {
				t.Errorf("Repository.PingDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_BatchInsert(t *testing.T) {
	type args struct {
		ctx        context.Context
		dataModels []models.Metrics
	}
	tests := []struct {
		name    string
		r       Repository
		args    args
		wantErr bool
	}{
		{
			name: "pos",
			r:    Repository{},
			args: args{
				ctx:        context.TODO(),
				dataModels: []models.Metrics{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.BatchInsert(tt.args.ctx, tt.args.dataModels); (err != nil) != tt.wantErr {
				t.Errorf("Repository.BatchInsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_GetCurrentCommit(t *testing.T) {
	tests := []struct {
		name string
		r    Repository
		want float64
	}{
		{
			name: "pos",
			r: Repository{
				[]models.Metrics{
					models.NewMetric("RandomValue", "gauge", nil, pointy.Float64(10.0), ""),
				},
			},
			want: 10.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.GetCurrentCommit(); got != tt.want {
				t.Errorf("Repository.GetCurrentCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRepo(t *testing.T) {
	tests := []struct {
		name string
		want Repository
	}{
		{
			name: "pos",
			want: Repository{
				JSONDB: []models.Metrics{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRepo()
			assert.Equal(t, got, tt.want)
		})
	}
}
