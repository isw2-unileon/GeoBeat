import { Source, Layer, FillLayerSpecification } from "@vis.gl/react-maplibre";
import { BaseGlobe } from "./base/BaseGlobe";
import { getDataFromISO, getISOFromCountry } from "@/lib/getCountryData";
import { notify } from "@/lib/notifier";

type InverseMapProps = {
  countryISO: string;
  setCountryISO: React.Dispatch<React.SetStateAction<string>>;
};

export function InverseModeMap({ countryISO, setCountryISO }: InverseMapProps) {
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
    filter: ["==", ["get", "name"], getDataFromISO(countryISO).name],
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
          const country: string = features[0]?.properties?.name;
          setCountryISO(getISOFromCountry(country));
          notify.info("Selected: " + country);
          notify.info("ISO: " + getISOFromCountry(country));
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
