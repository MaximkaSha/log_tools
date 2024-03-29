package utils

import (
	"encoding/hex"
	"errors"
	"testing"
)

func TestCheckIfStringIsNumber(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "positive #1",
			args: args{"100"},
			want: true,
		},
		{
			name: "positive #2",
			args: args{"100.0001"},
			want: true,
		},
		{
			name: "negative",
			args: args{"not a number"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckIfStringIsNumber(tt.args.v); got != tt.want {
				t.Errorf("CheckIfStringIsNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat64ToByte(t *testing.T) {
	type args struct {
		v float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive #1",
			args: args{100.0},
			want: "0000000000005940",
		},
		{
			name: "positive #2",
			args: args{100.0001},
			want: "b22e6ea301005940",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Float64ToByte(tt.args.v); hex.EncodeToString(got) != tt.want {
				t.Errorf("Float64ToByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckError(t *testing.T) {
	tests := []struct {
		name string
		args error
	}{
		{
			name: "positive #1",
			args: errors.New("Err..."),
		},
		{
			name: "positive #2",
			args: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CheckError(tt.args)
		})
	}
}
