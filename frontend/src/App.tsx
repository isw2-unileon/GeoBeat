import { AppField } from "./components/app-field";
import { AppDrawer } from "./components/app-drawer";
import { AppDialog } from "./components/app-dialog";
import { DailyModeMap } from "./components/map/DailyModeMap";
import { InverseModeMap } from "./components/map/InverseModeMap";
import { Toaster } from "./components/ui/sonner";

import { modes } from "./data/placeholder-data";
import { useState } from "react";
import { useEffect } from "react";
import { getDaily } from "./services/daily";
import { Attempts } from "./components/attempts";
import { Daily, GameStatus, STATUS } from "@/types/gameTypes";
import { getDataFromISO } from "./lib/getCountryData";
import { PopUp } from "./components/attempts-popup";
import { GuessProvider } from "./context/GuessProvider";
import { DailyModeTitle } from "./components/daily-title";
import { InverseModeTitle } from "./components/inverse-title";
import { LoadingText } from "./components/loading-text";
import { getInverse } from "./services/inverse";

export default function App() {
  const [mode, setMode] = useState<string>(modes[0]);
  const [daily, setDaily] = useState<Daily>(null);
  const [countryISO, setCountryISO] = useState<string>("UNKOWN");
  const [gameStatus, setGameStatus] = useState<GameStatus>({
    attempts: 0,
    status: STATUS.PLAYING,
  });

  useEffect(() => {
    const load = async () => {
      const data = await getDaily();
      setDaily(data);
    };

    load();
  }, []);

  useEffect(() => {
    if (daily) {
      setCountryISO(daily.country);
      setGameStatus({
        attempts: daily.attempts,
        status: daily.status,
      });
    }
  }, [daily]);

  const country = getDataFromISO(countryISO).name;

  const [forceLoad, setForceLoad] = useState(false);
  useEffect(() => {
    const timer = setTimeout(() => {
      setForceLoad(true);
    }, 3000);

    return () => clearTimeout(timer);
  }, []);

  const ready = countryISO !== "UNKOWN" || forceLoad;

  let content;
  switch (mode) {
    case modes[0]:
      if (ready) {
        content = (
          <>
            <DailyModeTitle />
            <DailyModeMap countryISO={countryISO} />
          </>
        );
      } else {
        content = (
          <>
            <DailyModeTitle />
            <LoadingText />
          </>
        );
      }

      break;

    case modes[1]:
      content = (
        <>
          <InverseModeTitle />
          <InverseModeMap
            countryISO={countryISO}
            setCountryISO={setCountryISO}
          />
        </>
      );
      break;
  }

  return (
    <main className="relative min-h-screen flex flex-col">
      <AppDialog />
      {content}
      <GuessProvider>
        {/* Desktop */}
        <div className="hidden md:block">
          <AppField
            country={country}
            mode={mode}
            setMode={setMode}
            gameStatus={gameStatus}
            setGameStatus={setGameStatus}
          />
        </div>
        {/* Mobile */}
        <div className="md:hidden">
          <AppDrawer
            country={country}
            mode={mode}
            setMode={setMode}
            gameStatus={gameStatus}
            setGameStatus={setGameStatus}
          />
        </div>
        <PopUp status={gameStatus.status} />
        <Attempts gameStatus={gameStatus} />
      </GuessProvider>
      <Toaster position={"top-center"} />
    </main>
  );
}
