package config

import (
	"reflect"
	"testing"
)

func TestParseKeyFloatMap(t *testing.T) {
	tests := []struct {
		in   string
		want map[int]float64
		ok   bool
	}{
		{"", map[int]float64{}, true},
		{"0=1.23,1=0.98", map[int]float64{0: 1.23, 1: 0.98}, true},
		{" 0 = 1 , 2 = -0.5", map[int]float64{0: 1.0, 2: -0.5}, true},
		{"bad", nil, false},
	}
	for _, tt := range tests {
		got, err := parseKeyFloatMap(tt.in)
		if (err == nil) != tt.ok {
			t.Fatalf("parseKeyFloatMap(%q) ok=%v err=%v", tt.in, tt.ok, err)
		}
		if tt.ok && !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parseKeyFloatMap(%q) = %v; want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseKeyIntMap(t *testing.T) {
	tests := []struct {
		in   string
		want map[int]int
		ok   bool
	}{
		{"", map[int]int{}, true},
		{"0=128,1=250", map[int]int{0: 128, 1: 250}, true},
		{"0=8, 2=16", map[int]int{0: 8, 2: 16}, true},
		{"bad", nil, false},
	}
	for _, tt := range tests {
		got, err := parseKeyIntMap(tt.in)
		if (err == nil) != tt.ok {
			t.Fatalf("parseKeyIntMap(%q) ok=%v err=%v", tt.in, tt.ok, err)
		}
		if tt.ok && !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parseKeyIntMap(%q) = %v; want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseKeyBoolMap(t *testing.T) {
	tests := []struct {
		in   string
		want map[int]bool
		ok   bool
	}{
		{"", map[int]bool{}, true},
		{"0=true,1=false", map[int]bool{0: true, 1: false}, true},
		{"0=true, 2=true", map[int]bool{0: true, 2: true}, true},
		{"bad", nil, false},
	}
	for _, tt := range tests {
		got, err := parseKeyBoolMap(tt.in)
		if (err == nil) != tt.ok {
			t.Fatalf("parseKeyBoolMap(%q) ok=%v err=%v", tt.in, tt.ok, err)
		}
		if tt.ok && !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parseKeyBoolMap(%q) = %v; want %v", tt.in, got, tt.want)
		}
	}
}
