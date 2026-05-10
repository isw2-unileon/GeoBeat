import { Source, Layer, FillLayerSpecification } from '@vis.gl/react-maplibre';
import { BaseGlobe } from './base/BaseGlobe';
import positions from '@/data/country-centroids.json';

type DailyMapProps = {
  country: string;
}


export function DailyModeMap({ country }: DailyMapProps) {

    const selectedCountyLayer: FillLayerSpecification = {
    id: 'selected-country-layer',
    type: 'fill',
    source: 'countries',
    filter: ['==', ['get', 'name'], country],
    paint: {
      'fill-color': '#5145ac',
      'fill-opacity': 0.4,
    }
  }

  const pos = positions.find(item => item.name == country) ?? {longitude: 0, latitude: 0}
  console.log(pos)

  return (
    <BaseGlobe longitude={pos.longitude} latitude={pos.latitude}>
      <Source
        id="countries"
        type="geojson"
        data="/data/countries.geojson"
      >
        <Layer {...selectedCountyLayer}/>
      </Source>
    </BaseGlobe>
  )
}