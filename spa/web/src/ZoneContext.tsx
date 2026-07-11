import React, { createContext, useContext, useState, useEffect } from "react";
import { DiseaseZone, getDiseaseZones } from "./api";

interface ZoneCtx {
  zones: DiseaseZone[];
  selectedZone: DiseaseZone | null;
  setSelectedZoneId: (id: number | null) => void;
  loading: boolean;
}

const ZoneContext = createContext<ZoneCtx>({
  zones: [],
  selectedZone: null,
  setSelectedZoneId: () => {},
  loading: false,
});

export function ZoneProvider({ children }: { children: React.ReactNode }) {
  const [zones, setZones] = useState<DiseaseZone[]>([]);
  const [selectedZoneId, setSelectedZoneId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getDiseaseZones()
      .then((zs) => setZones(zs))
      .finally(() => setLoading(false));
  }, []);

  const selectedZone = selectedZoneId
    ? zones.find((z) => z.id === selectedZoneId) || null
    : null;

  return (
    <ZoneContext.Provider value={{ zones, selectedZone, setSelectedZoneId, loading }}>
      {children}
    </ZoneContext.Provider>
  );
}

export function useZone() {
  return useContext(ZoneContext);
}

export default ZoneContext;
