package service

import "testing"

type randomCountryTestCase struct {
	name            string
	numberCountries int
	wantErr         bool
}

func Test_getRandomCountry(t *testing.T) {
	tests := []randomCountryTestCase{
		{
			name:            "valid number of countries",
			numberCountries: 5,
			wantErr:         false,
		},
		{
			name:            "zero countries requested",
			numberCountries: 0,
			wantErr:         true,
		},
		{
			name:            "negative number of countries requested",
			numberCountries: -3,
			wantErr:         true,
		},
		{
			name:            "number of countries requested exceeds supported",
			numberCountries: len(supportedCountries) + 1,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runRandomCountryAssertion(t, tt)
		})
	}
}

func runRandomCountryAssertion(t *testing.T, tt randomCountryTestCase) {
	t.Helper()

	got, gotErr := GetRandomCountry(tt.numberCountries)

	if gotErr != nil {
		if !tt.wantErr {
			t.Errorf("getRandomCountry() failed: %v", gotErr)
		}
		return
	}
	if tt.wantErr {
		t.Fatal("getRandomCountry() succeeded unexpectedly")
	}

	if len(got) != tt.numberCountries {
		t.Fatalf("getRandomCountry() = %v, want %v", len(got), tt.numberCountries)
	}

	seen := make(map[string]bool)
	for _, country := range got {
		if seen[country] {
			t.Errorf("getRandomCountry() returned duplicate country: %s", country)
		}
		seen[country] = true

		if !isCountrySupported(country) {
			t.Errorf("getRandomCountry() returned unsupported country: %s", country)
		}
	}
}

func isCountrySupported(target string) bool {
	for _, c := range supportedCountries {
		if c == target {
			return true
		}
	}
	return false
}
