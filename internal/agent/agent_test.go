package agent

import (
	"testing"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
)

func TestAgent_CollectLogs(t *testing.T) {
	tests := []struct {
		name string
		a    *Agent
	}{
		{
			name: "Test change",
			a: &Agent{
				logDB:   []models.Metrics{},
				counter: 0,
				cfg:     Config{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			var oldVal float64
			var newVal float64
			tt.a.CollectLogs()
			for _, val := range tt.a.logDB {
				if val.ID == "RandomValue" {
					oldVal = *val.Value
				}
			}
			tt.a.CollectLogs()
			for _, val := range tt.a.logDB {
				if val.ID == "RandomValue" {
					newVal = *val.Value
				}
			}
			assert.NotEqualValues(newVal, oldVal, "must not be equal")
		})
	}
}

func TestAgent_AppendMetric(t *testing.T) {
	type args struct {
		m models.Metrics
	}
	tests := []struct {
		name string
		a    *Agent
		args args
	}{
		{
			name: "add metric",
			a: &Agent{
				logDB:   []models.Metrics{},
				counter: 0,
				cfg:     Config{},
			},
			args: args{models.Metrics{
				ID:    "test val",
				MType: "counter",
				Delta: pointy.Int64(10),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			tt.a.AppendMetric(tt.args.m)
			assert.Equal(pointy.Int64(10), tt.a.logDB[0].Delta, "Must be equal")
		})
	}
}

func TestConfig_isDefault(t *testing.T) {
	type args struct {
		flagName string
		envName  string
	}
	tests := []struct {
		name string
		c    *Config
		args args
		want bool
	}{
		{
			name: "false",
			c:    &Config{},
			args: args{
				flagName: "test",
				envName:  "false",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.isDefault(tt.args.flagName, tt.args.envName); got != tt.want {
				t.Errorf("Config.isDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_UmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		c       *Config
		args    args
		wantErr bool
	}{
		{
			name: "false #1",
			c:    &Config{},
			args: args{
				[]byte(`not a json`),
			},
			wantErr: true,
		},
		{
			name: "pos #1",
			c:    &Config{},
			args: args{
				[]byte(`{"address":"0.0.0.0:80","report_interval":"2s","poll_interval":"2s"}`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.UmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Config.UmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_SendLogsbyJSONBatch(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		a       Agent
		args    args
		wantErr bool
	}{
		{
			name: "pos1",
			a: Agent{
				logDB: []models.Metrics{
					models.NewMetric("test", "counter", pointy.Int64(10), nil, ""),
				},
			},
			args: args{
				url: "not a host",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.SendLogsbyJSONBatch(tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("Agent.SendLogsbyJSONBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_SendLogsbyJSON(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		a       Agent
		args    args
		wantErr bool
	}{
		{
			name: "pos1",
			a: Agent{
				logDB: []models.Metrics{
					models.NewMetric("test", "counter", pointy.Int64(10), nil, ""),
				},
			},
			args: args{
				url: "not a host",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.SendLogsbyJSON(tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("Agent.SendLogsbyJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_getPostStrByIndex(t *testing.T) {
	type args struct {
		i   int
		url string
	}
	tests := []struct {
		name string
		a    Agent
		args args
		want string
	}{
		{
			name: "pos1",
			a: Agent{
				logDB: []models.Metrics{
					models.NewMetric("test", "counter", pointy.Int64(10), nil, ""),
				},
			},
			args: args{
				url: "1.1.1.1/",
			},
			want: "1.1.1.1/counter/test/10",
		},
		{
			name: "pos2",
			a: Agent{
				logDB: []models.Metrics{
					models.NewMetric("test", "gauge", pointy.Int64(10), pointy.Float64(10), ""),
				},
			},
			args: args{
				url: "1.1.1.1/",
			},
			want: "1.1.1.1/gauge/test/10.000000",
		},
		{
			name: "neg1",
			a: Agent{
				logDB: []models.Metrics{
					models.NewMetric("test", "not a type", pointy.Int64(10), pointy.Float64(10), ""),
				},
			},
			args: args{
				url: "1.1.1.1/",
			},
			want: "type unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.getPostStrByIndex(tt.args.i, tt.args.url); got != tt.want {
				t.Errorf("Agent.getPostStrByIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_SendLogsbyPost(t *testing.T) {
	type args struct {
		sData string
	}
	tests := []struct {
		name    string
		a       *Agent
		args    args
		wantErr bool
	}{
		{
			name: "pos1",
			a: &Agent{
				logDB: []models.Metrics{
					models.NewMetric("test", "counter", pointy.Int64(10), nil, ""),
				},
			},
			args: args{
				sData: "not a host",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.SendLogsbyPost(tt.args.sData); (err != nil) != tt.wantErr {
				t.Errorf("Agent.SendLogsbyPost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
