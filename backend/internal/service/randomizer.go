package service

import (
	"errors"
	"math/rand/v2"
)

var supportedCountries = []string{
	"spain", "france", "italy", "germany", "united kingdom", "ireland",
	"portugal", "netherlands", "belgium", "switzerland", "austria",
	"sweden", "norway", "denmark", "finland", "iceland",
	"poland", "czech republic", "hungary", "slovakia", "romania",
	"bulgaria", "greece", "turkey", "russia", "ukraine", "belarus",
	"serbia", "croatia", "slovenia", "lithuania", "latvia", "estonia",

	"united states", "canada", "mexico", "brazil", "argentina",
	"colombia", "chile", "peru", "venezuela", "ecuador", "uruguay",
	"paraguay", "bolivia", "costa rica", "panama", "dominican republic",
	"puerto rico", "guatemala", "el salvador", "honduras", "jamaica",

	"japan", "south korea", "india", "indonesia", "philippines",
	"thailand", "vietnam", "malaysia", "singapore", "taiwan",
	"china", "pakistan", "bangladesh", "sri lanka", "nepal",
	"saudi arabia", "united arab emirates", "israel", "lebanon",

	"south africa", "egypt", "nigeria", "kenya", "morocco",
	"algeria", "tunisia", "ghana", "senegal", "tanzania",

	"australia", "new zealand",
}

func GetRandomCountry(numberCountries int) ([]string, error) {
	if numberCountries <= 0 {
		return nil, errors.New("number of countries must be greater than zero")
	}
	if numberCountries > len(supportedCountries) {
		return nil, errors.New("number of countries requested exceeds the number of supported countries")
	}

	perm := rand.Perm(len(supportedCountries))
	selectedCountries := make([]string, numberCountries)
	for i := 0; i < numberCountries; i++ {
		selectedCountries[i] = supportedCountries[perm[i]]
	}
	return selectedCountries, nil
}
