package common

import (
	"fmt"
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

func TestMultiStep256Search(t *testing.T) {
	examples := []struct {
		name    string
		seed    string
		next    string
		targets []string
		want    string
	}{
		{
			name:    "no match",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"11111111"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "",
		},
		{
			name:    "one target beginning",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"5a552449"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "5a552449",
		},
		{
			name:    "one target middle",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"c35c2641"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "c35c2641",
		},
		{
			name:    "one target end",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"64e50414"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "64e50414",
		},
		{
			name:    "two target beginning",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"5a55244", "5a552449"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "5a552449",
		},
		{
			name:    "two target middle",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"c35c264", "c35c2641"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "c35c2641",
		},
		{
			name:    "two target end",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"64e5041", "64e50414"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "64e50414",
		},
	}

	for _, tt := range examples {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewMultiStep256(tt.seed)
			hasher.Hash(tt.next)
			got := hasher.Search(tt.targets)
			want := tt.want

			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}

func TestMultiStep256LargestTarget(t *testing.T) {
	examples := []struct {
		name    string
		seed    string
		next    string
		targets []string
		want    string
	}{
		{
			name:    "8 match",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"5a552449", "5a5524490", "5a55244900", "5a552449000"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "5a552449",
		},
		{
			name:    "8 9 match",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"5a552449", "5a552449a", "5a552449a0", "5a552449a00"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "5a552449a",
		},
		{
			name:    "8 9 10 match",
			seed:    "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001",
			next:    "0000",
			targets: []string{"5a552449", "5a552449a", "5a552449a9", "5a552449a90"},
			// Hashed string: 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
			want: "5a552449a9",
		},
	}

	for _, tt := range examples {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewMultiStep256(tt.seed)
			hasher.Hash(tt.next)
			got := hasher.Search(tt.targets)
			want := tt.want

			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
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

func BenchmarkMsFastParallel(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)

	b.RunParallel(func(pb *testing.PB) {
		seed := "1t0000NNNN4ExCj4NvjPUZBWzeHcoHWVBJoZfPEf1"
		next := "1001"
		hasher := NewMultiStep256(seed)
		for pb.Next() {
			hasher.Hash(next)
		}
	})
}

// strings.Contains bumps time from ~185 ms/op to ~320 ns/op
// The lengh of the search string does not impact ns/op values
//     in any meaningful way
func BenchmarkTargetSearch(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)
	seed := "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001"
	hasher := NewMultiStep256(seed)
	for n := 0; n < b.N; n++ {
		hasher.Hash(fmt.Sprintf("%04d", n))
		hasher.Search([]string{"1234567890123456"})
	}
}

func BenchmarkTargetSearchParallel(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)
	b.RunParallel(func(pb *testing.PB) {

		seed := "3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001"
		hasher := NewMultiStep256(seed)
		count := 0
		for pb.Next() {
			count++
			hasher.Hash(fmt.Sprintf("%04d", count))
			hasher.Search([]string{"1234567890123456"})
		}
	})
}
