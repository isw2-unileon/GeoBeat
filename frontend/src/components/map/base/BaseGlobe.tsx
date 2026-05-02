import { Map, MapLayerMouseEvent } from '@vis.gl/react-maplibre';
import 'maplibre-gl/dist/maplibre-gl.css';
import { ReactNode } from 'react';

type BaseGlobeProps = {
  onClick?: (event: MapLayerMouseEvent) => void;
  children?: ReactNode;
}

type ViewState = {
  longitude: number;
  latitude: number;
  zoom: number;
};

export function BaseGlobe({ children, onClick }: BaseGlobeProps) {
    return (
      <Map
        initialViewState={{...dailyViewState()}}
        style={{width: '100vw', height: '100vh'}}
        projection={'globe'}
        mapStyle="https://tiles.openfreemap.org/styles/positron"
        onClick={onClick}
      >
        {children}
      </Map>
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
