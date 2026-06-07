import { Source, Layer, FillLayerSpecification } from "@vis.gl/react-maplibre";
import { BaseGlobe } from "./base/BaseGlobe";
import { getDataFromISO } from "@/lib/getCountryData";
import { useEffect, useState } from "react";
import { COUNTRY } from "@/types/gameTypes";

type TimeTrialProps = {
  firstCountryISO: string | undefined;
  countryISO: string | undefined;
};

export function TimeTrialModeMap({
  firstCountryISO,
  countryISO,
}: TimeTrialProps) {
  const first_country_data = {
    longitude: 0,
    latitude: 0,
    name: COUNTRY.UNDEFINED,
  };

  const layer: FillLayerSpecification = {
    id: "selected-country-layer",
    type: "fill",
    source: "countries",
    filter: ["==", ["get", "name"], first_country_data.name],
    paint: {
      "fill-color": "#5145ac",
      "fill-opacity": 0.4,
    },
  };

  const [selectedCountyLayer, setSelectedCountryLayer] =
    useState<FillLayerSpecification>(layer);
  const [coords, setCoords] = useState({
    longitude: first_country_data.longitude,
    latitude: first_country_data.latitude,
  });

  useEffect(() => {
    const source =
      countryISO !== COUNTRY.UNDEFINED ? countryISO : firstCountryISO;

    let country_data = {
      longitude: 0,
      latitude: 0,
      name: COUNTRY.UNDEFINED,
    };

    if (source) {
      country_data = getDataFromISO(source);
    }

    setSelectedCountryLayer({
      id: "selected-country-layer",
      type: "fill",
      source: "countries",
      filter: ["==", ["get", "name"], country_data.name],
      paint: {
        "fill-color": "#5145ac",
        "fill-opacity": 0.4,
      },
    });

    setCoords({
      longitude: country_data.longitude,
      latitude: country_data.latitude,
    });
  }, [countryISO, firstCountryISO]);

  return (
    <BaseGlobe longitude={coords.longitude} latitude={coords.latitude}>
      <Source id="countries" type="geojson" data="/data/countries.geojson">
        <Layer {...selectedCountyLayer} />
      </Source>
    </BaseGlobe>
  );
}
