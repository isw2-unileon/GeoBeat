import { Map, MapLayerMouseEvent } from "@vis.gl/react-maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import { ReactNode, useEffect, useState } from "react";

type BaseGlobeProps = {
  onClick?: (event: MapLayerMouseEvent) => void;
  longitude?: number;
  latitude?: number;
  children?: ReactNode;
};

const DEFAULT_VIEW = {
  longitude: -100,
  latitude: 40,
  zoom: 2.5,
};

export function BaseGlobe({
  children,
  onClick,
  longitude,
  latitude,
}: BaseGlobeProps) {
  const [viewState, setViewState] = useState({
    longitude: longitude ?? DEFAULT_VIEW.longitude,
    latitude: latitude ?? DEFAULT_VIEW.latitude,
    zoom: DEFAULT_VIEW.zoom,
  });

  useEffect(() => {
    setViewState((prev) => ({
      ...prev,
      longitude: longitude ?? DEFAULT_VIEW.longitude,
      latitude: latitude ?? DEFAULT_VIEW.latitude,
    }));
  }, [longitude, latitude]);

  return (
    <Map
      {...viewState}
      onMove={(evt) => setViewState(evt.viewState)}
      style={{ width: "100vw", height: "100vh" }}
      projection={"globe"}
      mapStyle="https://tiles.openfreemap.org/styles/positron"
      onClick={onClick}
      maxZoom={4}
      minZoom={2.5}
      dragRotate={false}
    >
      {children}
    </Map>
  );
}
