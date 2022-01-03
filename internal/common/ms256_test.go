package common

import (
	"testing"
)

func TestMultiStep256(t *testing.T) {
	examples := []struct {
		seed string
		next string
		want string
	}{
		{
			seed: "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next: "0000",
			want: "5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414",
		},
		{
			seed: "3P0000555N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next: "0001",
			want: "5636892465063987f76296855e1843e360880c1068f86a012e8f270bcc36ef7f",
		},
		{
			seed: "1t0000NNNN4ExCj4NvjPUZBWzeHcoHWVBJoZfPEf11001",
			next: "0002",
			want: "ff681044e6b4c4fe68fe094a9ad920182ac7b7894991b3f7e6a63fa1aece9dcf",
		},
	}

	t.Run("basic hasher", func(t *testing.T) {
		for _, tt := range examples {
			got := MultiStep256Hash(tt.seed + tt.next)
			want := tt.want

			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		}
	})
	t.Run("fast hasher", func(t *testing.T) {
		for _, tt := range examples {
			hasher := NewMultiStep256(tt.seed)
			got := hasher.Hash(tt.next)
			want := tt.want

			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		}
	})
}

func BenchmarkMsBasic(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)
	val := "1t0000NNNN4ExCj4NvjPUZBWzeHcoHWVBJoZfPEf11001"
	for n := 0; n < b.N; n++ {
		MultiStep256Hash(val)
	}
}

func BenchmarkMsFast(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)
	seed := "1t0000NNNN4ExCj4NvjPUZBWzeHcoHWVBJoZfPEf1"
	next := "1001"
	hasher := NewMultiStep256(seed)

	for n := 0; n < b.N; n++ {
		hasher.Hash(next)
	}
}
