package service

import "testing"

func TestParseAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "integer", raw: "100", want: "100"},
		{name: "two decimals", raw: "10.50", want: "10.5"},
		{name: "one decimal", raw: "1.2", want: "1.2"},
		{name: "trimmed", raw: " 25.00 ", want: "25"},
		{name: "indonesian comma", raw: "10,50", want: "10.5"},
		{name: "indonesian one decimal", raw: "50,5", want: "50.5"},
		{name: "indonesian thousands", raw: "10.000,50", want: "10000.5"},
		{name: "zero", raw: "0", wantErr: true},
		{name: "negative", raw: "-1", wantErr: true},
		{name: "too many decimals", raw: "1.234", wantErr: true},
		{name: "too many indonesian decimals", raw: "1,234", wantErr: true},
		{name: "not a number", raw: "abc", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseAmount(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %s want %s", got.String(), tt.want)
			}
		})
	}
}
