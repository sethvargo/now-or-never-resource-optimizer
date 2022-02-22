package rates

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExchange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		hand *ResourceAlloc
		want *Trade
	}{
		// Default exchanges
		{
			name: "one_shell",
			hand: &ResourceAlloc{S: 1},
			want: &Trade{
				R: []*ResourceAlloc{{S: 1}},
				V: 1,
			},
		},
		{
			name: "one_tool",
			hand: &ResourceAlloc{T: 1},
			want: &Trade{
				R: []*ResourceAlloc{{T: 1}},
				V: 2,
			},
		},
		{
			name: "one_demon",
			hand: &ResourceAlloc{D: 1},
			want: &Trade{
				R: []*ResourceAlloc{{D: 1}},
				V: 2,
			},
		},
		{
			name: "one_crystal",
			hand: &ResourceAlloc{C: 1},
			want: &Trade{
				R: []*ResourceAlloc{{C: 1}},
				V: 2,
			},
		},
		{
			name: "three_shell",
			hand: &ResourceAlloc{S: 3},
			want: &Trade{
				R: []*ResourceAlloc{{S: 3}},
				V: 5,
			},
		},
		{
			name: "one_shell_two_tool",
			hand: &ResourceAlloc{S: 1, T: 2},
			want: &Trade{
				R: []*ResourceAlloc{{S: 1, T: 2}},
				V: 7,
			},
		},
		{
			name: "one_tool_two_demon",
			hand: &ResourceAlloc{T: 1, D: 2},
			want: &Trade{
				R: []*ResourceAlloc{{T: 1, D: 2}},
				V: 9,
			},
		},
		{
			name: "one_shell_one_tool_one_demon",
			hand: &ResourceAlloc{S: 1, T: 1, D: 1},
			want: &Trade{
				R: []*ResourceAlloc{{S: 1, T: 1, D: 1}},
				V: 9,
			},
		},
		{
			name: "three_crystal",
			hand: &ResourceAlloc{C: 3},
			want: &Trade{
				R: []*ResourceAlloc{{C: 3}},
				V: 11,
			},
		},
		{
			name: "one_shell_one_tool_one_demon_one_crystal",
			hand: &ResourceAlloc{S: 1, T: 1, D: 1, C: 1},
			want: &Trade{
				R: []*ResourceAlloc{{S: 1, T: 1, D: 1, C: 1}},
				V: 12,
			},
		},
		{
			name: "two_demon_two_crystal",
			hand: &ResourceAlloc{D: 2, C: 2},
			want: &Trade{
				R: []*ResourceAlloc{{D: 2, C: 2}},
				V: 14,
			},
		},
		{
			name: "three_shell_two_tool",
			hand: &ResourceAlloc{S: 3, T: 2},
			want: &Trade{
				R: []*ResourceAlloc{
					{S: 1, T: 2},
					{S: 1},
					{S: 1},
				},
				V: 9,
			},
		},
		{
			name: "two_of_all",
			hand: &ResourceAlloc{S: 2, T: 2, D: 2, C: 2},
			want: &Trade{
				R: []*ResourceAlloc{
					{S: 1, T: 1, D: 1, C: 1},
					{S: 1, T: 1, D: 1, C: 1},
				},
				V: 24,
			},
		},
		{
			name: "three_of_all",
			hand: &ResourceAlloc{S: 3, T: 3, D: 3, C: 3},
			want: &Trade{
				R: []*ResourceAlloc{
					{S: 1, T: 1, D: 1},
					{S: 1, T: 1, D: 1},
					{S: 1, T: 1, D: 1},
					{C: 3},
				},
				V: 38,
			},
		},
		{
			name: "four_of_all",
			hand: &ResourceAlloc{S: 4, T: 4, D: 4, C: 4},
			want: &Trade{
				R: []*ResourceAlloc{
					{S: 1, T: 1, D: 1, C: 1},
					{S: 1, T: 1, D: 1},
					{S: 1, T: 1, D: 1},
					{S: 1, T: 1, D: 1},
					{C: 3},
				},
				V: 50,
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Exchange(tc.hand)
			if len(got) < 1 {
				t.Fatal("no exchanges (should never happen)")
			}

			if diff := cmp.Diff(tc.want, got[0]); diff != "" {
				t.Errorf("bad exchange (-want, +got):\n%s", diff)
			}
		})
	}
}
