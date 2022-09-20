package crypto

import (
	"testing"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/openlyinc/pointy"
)

func TestCryptoService_InitCryptoService(t *testing.T) {
	type args struct {
		keyFile string
	}
	tests := []struct {
		name    string
		c       *CryptoService
		args    args
		wantErr bool
	}{
		{
			name: "Pos #1",
			c:    &CryptoService{},
			args: args{
				keyFile: "key",
			},
			wantErr: false,
		},
		{
			name: "Neg #1",
			c:    &CryptoService{},
			args: args{
				keyFile: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.InitCryptoService(tt.args.keyFile); (err != nil) != tt.wantErr {
				t.Errorf("CryptoService.InitCryptoService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCryptoService_Hash(t *testing.T) {
	type args struct {
		m *models.Metrics
	}
	tests := []struct {
		args     args
		wantHash string
		name     string
		c        CryptoService
		want     int
		wantErr  bool
	}{
		{
			name: "Pos #1",
			c:    CryptoService{},
			args: args{
				&models.Metrics{
					ID:    "hash",
					MType: "counter",
					Delta: pointy.Int64(10),
				},
			},
			want:     15,
			wantErr:  false,
			wantHash: "558847d8a28ce76a2f5679e7d9ca34fb88055959e338924530b0ffddeca35ace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Hash(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("CryptoService.Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CryptoService.Hash() = %v, want %v", got, tt.want)
			}
			if tt.args.m.Hash != tt.wantHash {
				t.Errorf("CryptoService.Hash() = %v, want %v", tt.args.m.Hash, tt.wantHash)
			}
		})
	}
}

func TestCryptoService_CheckHash(t *testing.T) {
	type args struct {
		m models.Metrics
	}
	tests := []struct {
		args args
		name string
		c    CryptoService
		want bool
	}{
		{
			name: "Pos #1",
			c:    CryptoService{},
			args: args{
				models.Metrics{
					ID:    "hash",
					MType: "counter",
					Delta: pointy.Int64(10),
					Hash:  "558847d8a28ce76a2f5679e7d9ca34fb88055959e338924530b0ffddeca35ace",
				},
			},
			want: true,
		},
		{
			name: "Neg #1",
			c:    CryptoService{},
			args: args{
				models.Metrics{
					ID:    "hash",
					MType: "counter",
					Delta: pointy.Int64(10),
					Hash:  "not_hash",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.CheckHash(tt.args.m); got != tt.want {
				t.Errorf("CryptoService.CheckHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCryptoService_IsServiceEnable(t *testing.T) {
	tests := []struct {
		name string
		c    CryptoService
		want bool
	}{
		{
			name: "Pos #1",
			c: CryptoService{
				IsEnable: true,
			},
			want: true,
		},
		{
			name: "Neg #1",
			c: CryptoService{
				IsEnable: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.IsServiceEnable(); got != tt.want {
				t.Errorf("CryptoService.IsServiceEnable() = %v, want %v", got, tt.want)
			}
		})
	}
}
