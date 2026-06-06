import { DailyField } from "./components/daily/daily-field";
import { DailyDrawer } from "./components/daily/daily-drawer";
import { AppDialog } from "./components/app-dialog";
import { DailyModeMap } from "./components/map/DailyModeMap";
import { InverseModeMap } from "./components/map/InverseModeMap";
import { Toaster } from "./components/ui/sonner";
import { modes } from "./data/constats";
import { useState } from "react";
import { useEffect } from "react";
import { getDaily } from "./services/daily";
import { Attempts } from "./components/attempts";
import { COUNTRY, Daily, GameStatus, Inverse, STATUS } from "@/types/gameTypes";
import { getDataFromISO } from "./lib/getCountryData";
import { PopUp } from "./components/attempts-popup";
import { GuessProvider } from "./context/GuessProvider";
import { DailyModeTitle } from "./components/daily/daily-title";
import { InverseModeTitle } from "@/components/inverse/inverse-title";
import { LoadingText } from "./components/loading-text";
import { getInverse } from "./services/inverse";
import { InverseField } from "./components/inverse/inverse-field";
import { InverseDrawer } from "./components/inverse/inverse-drawer";

export default function App() {
  const [mode, setMode] = useState<string>(modes[0]);

  // Declare modes data
  const [daily, setDaily] = useState<Daily>(null);
  const [gameStatus, setGameStatus] = useState<GameStatus>({
    attempts: 0,
    status: STATUS.PLAYING,
  });
  const [countryISO, setCountryISO] = useState<string>(COUNTRY.UNDEFINED_ISO);

  const [inverse, setInverse] = useState<Inverse>(null);
  const [inverseStatus, setInverseStatus] = useState<GameStatus>({
    attempts: 0,
    status: STATUS.PLAYING,
  });
  const [inverseISO, setInverseISO] = useState<string>(COUNTRY.UNDEFINED_ISO);

  // Load daily infomation
  useEffect(() => {
    const load = async () => {
      setDaily(await getDaily(1));
      setInverse(await getInverse(1));
    };

    load();
  }, []);

  const country = getDataFromISO(countryISO).name;
  useEffect(() => {
    if (daily) {
      setCountryISO(daily.country);
      setGameStatus({
        attempts: daily.attempts,
        status: daily.status,
      });
    }
  }, [daily]);

  const [song, setSong] = useState("No song");
  useEffect(() => {
    if (inverse) {
      setInverseStatus({
        attempts: inverse.attempts,
        status: inverse.status,
      });
      setSong(inverse.song);
    }
  }, [inverse]);

  const [forceLoad, setForceLoad] = useState(false);
  useEffect(() => {
    const timer = setTimeout(() => {
      setForceLoad(true);
    }, 3000);

    return () => clearTimeout(timer);
  }, []);

  const ready = countryISO !== COUNTRY.UNDEFINED_ISO || forceLoad;
  const dailyMode = mode === modes[0];
  const inverseMode = mode === modes[1];

  return (
    <main className="relative min-h-screen flex flex-col">
      <AppDialog />
      {dailyMode && <DailyModeTitle />}
      {inverseMode && <InverseModeTitle />}
      <div className={dailyMode ? "" : "hidden"}>
        {ready ? <DailyModeMap countryISO={countryISO} /> : <LoadingText />}
      </div>
      <div className={inverseMode ? "" : "hidden"}>
        <InverseModeMap countryISO={inverseISO} setCountryISO={setInverseISO} />
      </div>
      <GuessProvider>
        {/* Desktop */}
        <div className="hidden md:block">
          {dailyMode && (
            <DailyField
              country={country}
              mode={mode}
              setMode={setMode}
              gameStatus={gameStatus}
              setGameStatus={setGameStatus}
            />
          )}
          {inverseMode && (
            <InverseField
              song={song}
              countryISO={inverseISO}
              mode={mode}
              setMode={setMode}
              inverseStatus={inverseStatus}
              setInverseStatus={setInverseStatus}
            />
          )}
        </div>
        {/* Mobile */}
        <div className="md:hidden">
          {dailyMode && (
            <DailyDrawer
              country={country}
              mode={mode}
              setMode={setMode}
              gameStatus={gameStatus}
              setGameStatus={setGameStatus}
            />
          )}
          {inverseMode && (
            <InverseDrawer
              song={song}
              countryISO={inverseISO}
              mode={mode}
              setMode={setMode}
              inverseStatus={inverseStatus}
              setInverseStatus={setInverseStatus}
            />
          )}
        </div>
        {dailyMode && (
          <>
            <Attempts gameStatus={gameStatus} mode={mode} />
            <PopUp status={gameStatus.status} />
          </>
        )}
        {inverseMode && (
          <>
            <Attempts gameStatus={inverseStatus} mode={mode} />
            <PopUp status={inverseStatus.status} />
          </>
        )}
      </GuessProvider>
      <Toaster position={"top-center"} />
    </main>
  );
}
