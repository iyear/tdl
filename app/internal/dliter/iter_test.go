package dliter

import (
	"reflect"
	"testing"
)

func TestPreSum(t *testing.T) {
	tests := []struct {
		dialogs []*Dialog
		want    []int
	}{
		{
			dialogs: []*Dialog{{Messages: []int{1, 2, 3}}, {Messages: []int{1, 2}}},
			want:    []int{0, 3, 5},
		},
		{
			dialogs: []*Dialog{{Messages: []int{1, 2, 3}}, {Messages: []int{1, 2, 3}}, {Messages: []int{1, 2, 3, 4}}},
			want:    []int{0, 3, 6, 10},
		},
		{
			dialogs: []*Dialog{{Messages: []int{1, 2, 3}}, {Messages: []int{1, 2, 3}}, {Messages: []int{1, 2, 3, 4}}, {Messages: []int{1}}},
			want:    []int{0, 3, 6, 10, 11},
		},
	}

	for _, tt := range tests {
		got := preSum(tt.dialogs)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("preSum() = %v, want %v", got, tt.want)
		}
	}
}

func TestIter_ij2n(t *testing.T) {
	tests := []struct {
		dialogs []*Dialog
		input   []struct {
			i, j int
		}
		want []int
	}{
		{
			dialogs: []*Dialog{{Messages: []int{1, 2, 3}}, {Messages: []int{1, 2}}},
			input: []struct {
				i, j int
			}{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}},
			want: []int{0, 1, 2, 3, 4},
		},
		{
			dialogs: []*Dialog{{Messages: []int{1, 2, 3}}, {Messages: []int{1, 2, 3}}, {Messages: []int{1, 2, 3, 4}}},
			input: []struct {
				i, j int
			}{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}, {2, 3}},
			want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
	}

	for _, tt := range tests {
		iter := &Iter{preSum: preSum(tt.dialogs), dialogs: tt.dialogs}

		for i, input := range tt.input {
			got := iter.ij2n(input.i, input.j)
			if got != tt.want[i] {
				t.Errorf("ij2n(%v, %v) = %v, want %v", input.i, input.j, got, tt.want[i])
			}
		}
	}
}
