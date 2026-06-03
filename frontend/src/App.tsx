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
import { Daily, GameStatus, STATUS } from "@/types/gameTypes";
import getCountryData from "./lib/getCountryData";
import { PopUp } from "./components/attempts-popup";
import { GuessProvider } from "./context/GuessProvider";
import { DailyModeTitle } from "./components/daily-title";

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
    status: STATUS.PLAYING,
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
        <FreeModeMap countryISO={countryISO} setCountryISO={setCountryISO} />
      );
      break;
  }

  return (
    <main className="relative min-h-screen flex flex-col">
      <DailyModeTitle />
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
