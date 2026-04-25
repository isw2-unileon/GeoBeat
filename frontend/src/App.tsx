import {Map, Source, Layer, FillLayerSpecification} from '@vis.gl/react-maplibre';
import 'maplibre-gl/dist/maplibre-gl.css';

import { AppField } from './components/app-field';
import { AppDrawer } from './components/app-drawer';
import { AppDialog } from './components/app-dialog';

import { useState } from 'react';
import type { Feature, Geometry, GeoJsonProperties } from "geojson";

type ViewState = {
  longitude: number;
  latitude: number;
  zoom: number;
};

export default function App() {

  const [country, setCountry] = useState<string>('(Select a country by clicking on it)')

  return (
      <main className="relative min-h-screen flex flex-col">
        <DailyModeTitle />
        <AppDialog />
        <ContentMap setCountry={setCountry}/>
        {/* Desktop */}
        <div className='hidden md:block'>
          <AppField country={country} />
        </div>
        {/* Mobile */}
        <div className='md:hidden'>
          <AppDrawer country={country} />
        </div>
      </main>
  )
}

function ContentMap({ setCountry }: { setCountry: React.Dispatch<React.SetStateAction<string>> }) {

  const [countryFeatures, setCountyFeatures] = useState<Feature<Geometry, GeoJsonProperties>[]>([]);

  const countryLayer: FillLayerSpecification = {
    id: 'country-layer',
    type: 'fill',
    source: 'countries',
    paint: {
      'fill-color': '#2d643c',
      'fill-opacity': 0.4
    }
  };

  const selectedCountyLayer: FillLayerSpecification = {
    id: 'selected-country-layer',
    type: 'fill',
    source: 'selection',
    paint: {
      'fill-color': '#5145ac',
      'fill-opacity': 0.4,
    }
  }

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
      const country: string = features[0]?.properties.name
      console.log(country);
      setCountry(country);
      setCountyFeatures(
        features.map((f) => ({
          type: "Feature",
          geometry: f.geometry,
          properties: f.properties ?? {}
        }))
      );
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
    <Source
      id='selection'
      type='geojson'
      data={{
        type: "FeatureCollection",
        features: countryFeatures
      }}
    >
      <Layer {...selectedCountyLayer} />
    </Source>
  </Map>
}

function DailyModeTitle() {
  return (
  <h1 className="md:absolute md:top-6 md:left-14 md:text-5xl md:translate-x-0
                absolute top-2 left-1/2 -translate-x-1/2 text-outline
                text-2xl text-center text-blue-600 font-semibold font-[sans] animate-fade-in-down z-1">
    DAILY MODE
  </h1>
  )
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