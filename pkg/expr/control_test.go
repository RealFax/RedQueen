package expr_test

import (
	"github.com/RealFax/RedQueen/pkg/expr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsZero(t *testing.T) {
	tests := []struct {
		name   string
		expect bool
		want   bool
	}{
		{"TestZeroInt", expr.IsZero(0), true},
		{"TestNonZeroInt", expr.IsZero(5), false},
		{"TestZeroString", expr.IsZero(""), true},
		{"TestNonZeroString", expr.IsZero("Hello"), false},
		{"TestZeroFloat", expr.IsZero(0.0), true},
		{"TestNonZeroFloat", expr.IsZero(3.14), false},
		{"TestZeroBool", expr.IsZero(false), true},
		{"TestNonZeroBool", expr.IsZero(true), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expect != tt.want {
				t.Errorf("IsZero() = %v, want %v", tt.expect, tt.want)
			}
		})
	}
}

func TestIf(t *testing.T) {
	type args struct {
		f    bool
		then any
		end  any
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{"TestIfTrueInt", args{f: true, then: 5, end: 10}, 5},
		{"TestIfFalseInt", args{f: false, then: 5, end: 10}, 10},
		{"TestIfTrueString", args{f: true, then: "Hello", end: "World"}, "Hello"},
		{"TestIfFalseString", args{f: false, then: "Hello", end: "World"}, "World"},
		{"TestIfTrueBool", args{f: true, then: true, end: false}, true},
		{"TestIfFalseBool", args{f: false, then: true, end: false}, false},
		{"TestIfTrueSlice", args{f: true, then: []int{1, 2, 3}, end: []int{4, 5, 6}}, []int{1, 2, 3}},
		{"TestIfFalseSlice", args{f: false, then: []int{1, 2, 3}, end: []int{4, 5, 6}}, []int{4, 5, 6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expr.If(tt.args.f, tt.args.then, tt.args.end); !assert.Equal(t, tt.want, got) {
				t.Errorf("If() = %v, want %v", got, tt.want)
			}

		})
	}
}
