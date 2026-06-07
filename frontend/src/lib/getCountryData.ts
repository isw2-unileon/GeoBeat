import data from "@/data/country-centroids.json";
import { COUNTRY } from "@/types/gameTypes";

type countryData = {
  longitude: number;
  latitude: number;
  name: string;
};

export function getDataFromISO(ISO_name: string): countryData {
  const country_data = data.find((item) => item.ISO_name === ISO_name) ?? {
    longitude: 0,
    latitude: 0,
    name: COUNTRY.UNDEFINED,
  };

  return country_data;
}

export function getISOFromCountry(name: string): string {
  const country = data.find((item) => item.name === name);
  return country?.ISO_name ?? COUNTRY.UNDEFINED_ISO;
}
