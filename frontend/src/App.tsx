import { AppField } from "./components/app-field";
import { AppDrawer } from "./components/app-drawer";
import { AppDialog } from "./components/app-dialog";
import { DailyModeMap } from "./components/map/DailyModeMap";
import { FreeModeMap } from "./components/map/FreeModeMap";
import { Toaster } from "./components/ui/sonner";
import { modes } from "./data/placeholder-data";

import { useState } from "react";
import { useEffect } from "react";
import { getDaily } from "./services/daily";
import { Attempts } from "./components/attempts";
import { Daily } from "./services/daily";
import getCountryData from "./lib/getCountryData";

export type GameStatus = {
  // TODO where is defined
  attempts: number;
  status: string;
};

export default function App() {
  const [mode, setMode] = useState<string>(modes[0]);
  const [daily, setDaily] = useState<Daily>(null);

  useEffect(() => {
    const load = async () => {
      const data = await getDaily();
      setDaily(data);
    };

    load();
  }, []);

  const [countryISO, setCountryISO] = useState<string>("ES");
  useEffect(() => {
    if (daily?.country) {
      setCountryISO(daily.country);
    }
  }, [daily]);

  const [gameStatus, setGameStatus] = useState<GameStatus>({
    attempts: 0,
    status: "none",
  });
  useEffect(() => {
    if (daily?.attempts) {
      setGameStatus({
        attempts: daily.attempts,
        status: daily.status,
      });
    }
  }, [daily]);

  const country = getCountryData(countryISO).name;

  let content;
  switch (mode) {
    case modes[0]:
      content = <DailyModeMap countryISO={countryISO} />;
      break;

    case modes[1]:
      content = (
        <FreeModeMap countryISO={countryISO} setCountry={setCountryISO} />
      );
      break;
  }

  return (
    <main className="relative min-h-screen flex flex-col">
      <DailyModeTitle />
      <AppDialog />
      {content}
      {/* Desktop */}
      <div className="hidden md:block">
        <AppField
          country={country}
          setMode={setMode}
          setGameStatus={setGameStatus}
        />
      </div>
      {/* Mobile */}
      <div className="md:hidden">
        <AppDrawer country={country} />
      </div>
      <Attempts gameStatus={gameStatus} />
      <Toaster position={"top-center"} />
      <div className="hidden">
        <CorrectPopUp />
      </div>
    </main>
  );
}

function DailyModeTitle() {
  return (
    <h1
      className="md:absolute md:top-6 md:left-14 md:text-5xl md:translate-x-0
                absolute top-2 left-1/2 -translate-x-1/2 text-outline
                text-2xl text-center text-blue-600 font-semibold font-[sans] animate-fade-in-down z-1"
    >
      DAILY MODE
    </h1>
  );
}

function CorrectPopUp() {
  return (
    <label className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 animate-pop-fade text-6xl">
      ✅
    </label>
  );
}
