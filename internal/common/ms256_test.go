package common

import (
	"crypto/sha256"
	"testing"
)

// 2022/01/03 07:34:34 <- JOINOK N6VxgLSpbni8kLbyUAjYXdHCPt2VEp 3p0000000 PoolData 39324 6E4A93F0F266B2352F71F41461BC1F00 12 0 111 0 -30 19892 3
// ************* SeedFullBytes is: [51 112 48 48 48 48 48 48 48 78 54 86 120 103 76 83 112 98 110 105 56 107 76 98 121 85 65 106 89 88 100 72 67 80 116 50 86 69 112 49 49 48 48 49]
// ************* SeedFullBytes is: 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001
// ************* buff is         : 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001
// ************* seedLen is      : 44
// ************* buff is        : 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp110010000
// ************* buff.Bytes() is: [51 112 48 48 48 48 48 48 48 78 54 86 120 103 76 83 112 98 110 105 56 107 76 98 121 85 65 106 89 88 100 72 67 80 116 50 86 69 112 49 49 48 48 49 48 48 48 48]
// ************* buff.Bytes() is: 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp110010000
// ************* encoded is     : [53 97 53 53 50 52 52 57 97 57 98 55 50 57 52 51 57 56 57 97 102 99 51 53 99 50 54 52 49 100 55 97 101 50 101 50 56 54 51 48 54 51 49 57 49 101 49 51 50 100 57 100 48 100 102 49 54 52 101 53 48 52 49 52]
// ************* encoded is     : 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
// ************* val is         : 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
// ************* val is         : 5a552449a9b72943989afc35c2641d7ae2e2863063191e132d9d0df164e50414
// ************* buff is        : 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp110010001
// ************* buff.Bytes() is: [51 112 48 48 48 48 48 48 48 78 54 86 120 103 76 83 112 98 110 105 56 107 76 98 121 85 65 106 89 88 100 72 67 80 116 50 86 69 112 49 49 48 48 49 48 48 48 49]
// ************* buff.Bytes() is: 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp110010001
// ************* encoded is     : [48 102 50 56 100 57 102 54 51 99 49 57 53 51 54 56 56 57 98 48 52 56 97 53 99 97 54 102 97 101 51 55 51 55 50 53 57 54 102 53 98 52 98 51 101 102 57 57 102 101 98 98 101 52 52 54 97 55 97 101 52 100 97 56]
// ************* encoded is     : 0f28d9f63c19536889b048a5ca6fae37372596f5b4b3ef99febbe446a7ae4da8
// ************* val is         : 0f28d9f63c19536889b048a5ca6fae37372596f5b4b3ef99febbe446a7ae4da8
// ************* val is         : 0f28d9f63c19536889b048a5ca6fae37372596f5b4b3ef99febbe446a7ae4da8
// ************* buff is        : 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp110010002
// ************* buff.Bytes() is: [51 112 48 48 48 48 48 48 48 78 54 86 120 103 76 83 112 98 110 105 56 107 76 98 121 85 65 106 89 88 100 72 67 80 116 50 86 69 112 49 49 48 48 49 48 48 48 50]
// ************* buff.Bytes() is: 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp110010002
// ************* encoded is     : [102 57 51 53 55 102 52 52 48 102 54 98 56 98 97 98 52 99 51 54 98 50 98 53 50 55 52 56 97 56 99 54 99 100 100 100 50 56 52 52 52 97 53 50 51 51 102 50 48 98 102 99 56 98 49 98 100 56 54 53 48 99 52 51]
// ************* encoded is     : f9357f440f6b8bab4c36b2b52748a8c6cddd28444a5233f20bfc8b1bd8650c43
// ************* val is         : f9357f440f6b8bab4c36b2b52748a8c6cddd28444a5233f20bfc8b1bd8650c43
// ************* val is         : f9357f440f6b8bab4c36b2b52748a8c6cddd28444a5233f20bfc8b1bd8650c43
// ************* buff is        : 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp110010003
// ************* buff.Bytes() is: [51 112 48 48 48 48 48 48 48 78 54 86 120 103 76 83 112 98 110 105 56 107 76 98 121 85 65 106 89 88 100 72 67 80 116 50 86 69 112 49 49 48 48 49 48 48 48 51]
// ************* buff.Bytes() is: 3p0000000N6VxgLSpbni8kLbyUAjYXdHCPt2VEp110010003

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

var seed string
var next string
var nextBytes []byte

func BenchmarkAAAFoo(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)
	seed = "1t0000NNNN4ExCj4NvjPUZBWzeHcoHWVBJoZf"
	next = "PEf1"
	hasher := NewFoo(seed)

	for n := 0; n < b.N; n++ {
		hasher.Hash(next)
	}
}

func BenchmarkAAABarNextBytes(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)
	seed = "1t0000NNNN4ExCj4NvjPUZBWzeHcoHWVBJoZfPEf1"
	nextBytes = []byte("PEf1")
	hasher := NewBar(seed)
	for n := 0; n < b.N; n++ {
		hasher.HashBytes(nextBytes)
	}
}

func BenchmarkNew256(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)

	hasher := sha256.New()
	hasher.Write([]byte("1t0000NNNN4ExCj4NvjPUZBWzeHcoHWVBJoZf"))
	for n := 0; n < b.N; n++ {
		hasher.Sum([]byte("PEf1"))
	}
}
