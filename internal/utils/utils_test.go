package utils

import "testing"

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
