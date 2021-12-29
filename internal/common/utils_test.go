package common

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
		got := FormatHashRate(tt.hashRate)
		if got != tt.want {
			t.Errorf("got %s want %s", got, tt.want)
		}
	}
}

func TestBalanceFormatting(t *testing.T) {

	examples := []struct {
		balance string
		want    string
	}{
		{want: "0.00000000 Noso", balance: "0"},
		{want: "0.00000001 Noso", balance: "1"},
		{want: "0.00000021 Noso", balance: "21"},
		{want: "0.00000321 Noso", balance: "321"},
		{want: "0.00004321 Noso", balance: "4321"},
		{want: "0.00054321 Noso", balance: "54321"},
		{want: "0.00654321 Noso", balance: "654321"},
		{want: "0.07654321 Noso", balance: "7654321"},
		{want: "0.87654321 Noso", balance: "87654321"},
		{want: "9.87654321 Noso", balance: "987654321"},
		{want: "19.87654321 Noso", balance: "1987654321"},
	}

	for _, tt := range examples {
		got := FormatBalance(tt.balance)
		if got != tt.want {
			t.Errorf("got %s want %s", got, tt.want)
		}
	}
}
