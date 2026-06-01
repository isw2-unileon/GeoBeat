import { Source, Layer, FillLayerSpecification } from "@vis.gl/react-maplibre";
import { BaseGlobe } from "./base/BaseGlobe";
import getCountryData from "@/lib/getCountryData";

type FreeMapProps = {
  countryISO: string;
  setCountryISO: React.Dispatch<React.SetStateAction<string>>;
};

export function FreeModeMap({ countryISO, setCountryISO }: FreeMapProps) {
  const countryLayer: FillLayerSpecification = {
    id: "country-layer",
    type: "fill",
    source: "countries",
    paint: {
      "fill-color": "#2d643c",
      "fill-opacity": 0.4,
    },
  };

  const selectedCountyLayer: FillLayerSpecification = {
    id: "selected-country-layer",
    type: "fill",
    source: "countries",
    filter: ["==", ["get", "name"], getCountryData(countryISO).name],
    paint: {
      "fill-color": "#5145ac",
      "fill-opacity": 0.4,
    },
  };

  return (
    <BaseGlobe
      onClick={(e) => {
        const features = e.target.queryRenderedFeatures(e.point, {
          layers: ["country-layer"],
        });

        if (features.length > 0) {
          const country: string =
            features[0]?.properties?.["ISO3166-1-Alpha-2"];
          setCountryISO(country);
        }
      }}
    >
      <Source id="countries" type="geojson" data="/data/countries.geojson">
        <Layer {...countryLayer} />
        <Layer {...selectedCountyLayer} />
      </Source>
    </BaseGlobe>
  );
}
