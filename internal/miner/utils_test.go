package miner

import "testing"

func TestHashFormatting(t *testing.T) {

	examples := []struct {
		hashRate string
		want     string
	}{
		{want: "  1.000 Hash/s", hashRate: "1"},
		{want: " 12.000 Hash/s", hashRate: "12"},
		{want: "123.000 Hash/s", hashRate: "123"},
		{want: "  1.234 Khash/s", hashRate: "1234"},
		{want: " 12.345 Khash/s", hashRate: "12345"},
		{want: "123.456 Khash/s", hashRate: "123456"},
		{want: "  1.234 Mhash/s", hashRate: "1234560"},
		{want: " 12.345 Mhash/s", hashRate: "12345600"},
		{want: "123.456 Mhash/s", hashRate: "123456000"},
		{want: "  1.234 Ghash/s", hashRate: "1234560000"},
		{want: " 12.345 Ghash/s", hashRate: "12345600000"},
		{want: "123.456 Ghash/s", hashRate: "123456000000"},
		{want: "  1.234 Thash/s", hashRate: "1234560000000"},
		{want: " 12.345 Thash/s", hashRate: "12345600000000"},
		{want: "123.456 Thash/s", hashRate: "123456000000000"},
		{want: "  1.234 Phash/s", hashRate: "1234560000000000"},
		{want: " 12.345 Phash/s", hashRate: "12345600000000000"},
		{want: "123.456 Phash/s", hashRate: "123456000000000000"},
		{want: "  1.234 Ehash/s", hashRate: "1234560000000000000"},
		{want: " 12.345 Ehash/s", hashRate: "12345600000000000000"},
		{want: "123.456 Ehash/s", hashRate: "123456000000000000000"},
		{want: "  1.234 Zhash/s", hashRate: "1234560000000000000000"},
		{want: " 12.345 Zhash/s", hashRate: "12345600000000000000000"},
		{want: "123.456 Zhash/s", hashRate: "123456000000000000000000"},
	}

	for _, tt := range examples {
		got := formatHashRate(tt.hashRate)
		if got != tt.want {
			t.Errorf("got %s want %s", got, tt.want)
		}
	}
}
