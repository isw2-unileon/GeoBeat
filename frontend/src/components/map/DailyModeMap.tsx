import { Source, Layer, FillLayerSpecification } from "@vis.gl/react-maplibre";
import { BaseGlobe } from "./base/BaseGlobe";
import { getDataFromISO } from "@/lib/getCountryData";

type DailyMapProps = {
  countryISO: string;
};

export function DailyModeMap({ countryISO }: DailyMapProps) {
  const country_data = getDataFromISO(countryISO);

  const selectedCountyLayer: FillLayerSpecification = {
    id: "selected-country-layer",
    type: "fill",
    source: "countries",
    filter: ["==", ["get", "name"], country_data.name],
    paint: {
      "fill-color": "#5145ac",
      "fill-opacity": 0.4,
    },
  };

  return (
    <BaseGlobe
      longitude={country_data.longitude}
      latitude={country_data.latitude}
    >
      <Source id="countries" type="geojson" data="/data/countries.geojson">
        <Layer {...selectedCountyLayer} />
      </Source>
    </BaseGlobe>
  );
}
