package storage

import (
	"context"
	"testing"

	"github.com/MaximkaSha/log_tools/internal/models"
)

func TestRepository_insertCount(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name    string
		r       Repository
		args    models.Metrics
		wantErr bool
	}{
		{
			name: "positive",
			r:    NewRepo(),
			args: models.Metrics{
				ID:    "Test",
				MType: "counter",
				Delta: new(int64),
			},
			wantErr: false,
		},
		{
			name: "positive #2",
			r:    NewRepo(),
			args: models.Metrics{
				ID:    "Test",
				MType: "gauge",
				Value: new(float64),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			//	if tt.name == "postive #2" || tt.name == "negative #2" {
			//		tt.r.InsertData("counter", "Test", "100")
			//	}
			if err := tt.r.InsertMetric(ctx, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Repository.InsertMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

/*
func TestRepository_insertGouge(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name    string
		r       Repository
		args    args
		wantErr bool
	}{
		{
			name: "positive",
			r:    NewRepo(),
			args: args{
				name:  "Test",
				value: "123.000",
			},
			wantErr: false,
		},
		{
			name: "positive",
			r:    NewRepo(),
			args: args{
				name:  "Test",
				value: "not float",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.insertGouge(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Repository.insertGouge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_GetByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name  string
		r     Repository
		args  args
		want  string
		want1 bool
	}{
		{
			name: "positive",
			r:    NewRepo(),
			args: args{
				name: "Test",
			},
			want:  "100",
			want1: true,
		},
		{
			name: "negative",
			r:    NewRepo(),
			args: args{
				name: "NegativeTest",
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.InsertData("gauge", "Test", "100")
			got, got1 := tt.r.GetByName(tt.args.name)
			if got != tt.want {
				t.Errorf("Repository.GetByName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Repository.GetByName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestRepository_GetAll(t *testing.T) {
	tests := []struct {
		name string
		r    Repository
		want map[string]string
	}{
		{
			name: "positive",
			r: Repository{map[string]string{
				"Test": "100",
			},
				NewRepo().JSONDB},
			want: map[string]string{
				"Test": "100",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.GetAll(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.GetAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_InsertData(t *testing.T) {
	type args struct {
		typeVar string
		name    string
		value   string
	}
	tests := []struct {
		name string
		r    Repository
		args args
		want int
	}{
		{
			name: "positive",
			r:    NewRepo(),
			args: args{
				typeVar: "counter",
				name:    "Test",
				value:   "123",
			},
			want: 200,
		},
		{
			name: "positive #2",
			r:    NewRepo(),
			args: args{
				typeVar: "gauge",
				name:    "Test",
				value:   "123",
			},
			want: 200,
		},
		{
			name: "negative #1",
			r:    NewRepo(),
			args: args{
				typeVar: "counter",
				name:    "Test",
				value:   "error",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "negative #2",
			r:    NewRepo(),
			args: args{
				typeVar: "gauge",
				name:    "Test",
				value:   "error",
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.InsertData(tt.args.typeVar, tt.args.name, tt.args.value); got != tt.want {
				t.Errorf("Repository.InsertData() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
