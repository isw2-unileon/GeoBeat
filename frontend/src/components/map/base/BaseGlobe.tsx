import { Map, MapLayerMouseEvent } from '@vis.gl/react-maplibre';
import 'maplibre-gl/dist/maplibre-gl.css';
import { ReactNode } from 'react';

type BaseGlobeProps = {
  onClick?: (event: MapLayerMouseEvent) => void;
  longitude?: number;
  latitude?: number;
  children?: ReactNode;
}

const DEFAULT_VIEW = {
  longitude: -100,
  latitude: 40,
  zoom: 2.5
};

export function BaseGlobe({ children, onClick, longitude, latitude }: BaseGlobeProps) {
    return (
      <Map
        initialViewState={{
          longitude: longitude ?? DEFAULT_VIEW.longitude,
          latitude: latitude ?? DEFAULT_VIEW.latitude,
          zoom: DEFAULT_VIEW.zoom
        }}
        style={{width: '100vw', height: '100vh'}}
        projection={'globe'}
        mapStyle="https://tiles.openfreemap.org/styles/positron"
        onClick={onClick}
      >
        {children}
      </Map>
    )
}
