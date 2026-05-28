import data from "@/data/country-centroids.json";

type countryData = {
  longitude: number;
  latitude: number;
  name: string;
};

export default function getCountryData(ISO_name: string): countryData {
  const country_data = data.find((item) => item.ISO_name == ISO_name) ?? {
    longitude: 0,
    latitude: 0,
    name: "UNKNOWN",
  };

  return country_data;
}
