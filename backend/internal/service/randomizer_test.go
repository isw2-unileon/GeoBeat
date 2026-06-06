package service

import "testing"

func Test_getRandomCountry(t *testing.T) {
	tests := []struct {
		name            string
		numberCountries int
		wantErr         bool
	}{
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
				t.Errorf("getRandomCountry() = %v, want %v", got, tt.numberCountries)
			}

			seen := make(map[string]bool)

			for _, country := range got {
				if seen[country] {
					t.Errorf("getRandomCountry() returned duplicate country: %s", country)
				}
				seen[country] = true

				found := false
				for _, c := range supportedCountries {
					if c == country {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("getRandomCountry() returned unsupported country: %s", country)
				}
			}
		})
	}
}
