import {Map, Source, Layer} from '@vis.gl/react-maplibre';
import {FillLayer} from '@vis.gl/react-maplibre';
import 'maplibre-gl/dist/maplibre-gl.css';

type ViewState = {
  longitude: number;
  latitude: number;
  zoom: number;
};

const countryLayer: FillLayer = {
  id: 'country-layer',
  type: 'fill',
  paint: {
    'fill-color': '#2d643c',
    'fill-opacity': 0.4
  }
};

export default function App() {

  return <Map
    initialViewState={{...dailyViewState()}}
    style={{width: '100vw', height: '100vh'}}
    projection={'globe'}
    mapStyle="https://tiles.openfreemap.org/styles/positron"
    onClick={(e) => {
    const features = e.target.queryRenderedFeatures(e.point, {
      layers: ['country-layer']
    });

    if (features.length > 0) {
      console.log(features[0].properties.name);
    }
  }}
  >
    <Source
      id="countries"
      type="geojson"
      data="https://raw.githubusercontent.com/datasets/geo-countries/master/data/countries.geojson"
    >
      <Layer {...countryLayer}/>
    </Source>
  </Map>
}

function dailyViewState(): ViewState {
  // Need to retieve daily country and associate country to longitude and latitude
  const longitude = -100;
  const latitude = 40;
  return {
    longitude,
    latitude,
    zoom: 2.5
  }
}