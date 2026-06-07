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
import {
  COUNTRY,
  Daily,
  GameStatus,
  Inverse,
  STATUS,
  STATUS_TIME,
  Timetrial,
  TimetrialStatus,
} from "@/types/gameTypes";
import { getDataFromISO } from "./lib/getCountryData";
import { PopUp } from "./components/attempts-popup";
import { GuessProvider } from "./context/GuessProvider";
import { DailyModeTitle } from "./components/daily/daily-title";
import { InverseModeTitle } from "@/components/inverse/inverse-title";
import { getInverse } from "./services/inverse";
import { InverseField } from "./components/inverse/inverse-field";
import { InverseDrawer } from "./components/inverse/inverse-drawer";
import { TimetrialField } from "./components/timetrial/timetrial-field";
import { TimeTrialModeMap } from "./components/map/TimeTrialModeMap";
import { TimetrialModeTitle } from "./components/timetrial/timetrial-title";
import { TimetrialDrawer } from "./components/timetrial/timetrial-drawer";

export default function App() {
  const [mode, setMode] = useState<string>(modes[0]);

  // daily mode data
  const [daily, setDaily] = useState<Daily>(null);
  const [gameStatus, setGameStatus] = useState<GameStatus>({
    attempts: 0,
    status: STATUS.PLAYING,
  });
  const [countryISO, setCountryISO] = useState<string>(COUNTRY.UNDEFINED_ISO);

  // inverse mode data
  const [inverse, setInverse] = useState<Inverse>(null);
  const [inverseStatus, setInverseStatus] = useState<GameStatus>({
    attempts: 0,
    status: STATUS.PLAYING,
  });
  const [inverseISO, setInverseISO] = useState<string>(COUNTRY.UNDEFINED_ISO);

  // timetrial mode data
  const [timetrial, setTimetrial] = useState<Timetrial>({
    status: STATUS_TIME.PLAYING,
    target_country: COUNTRY.UNDEFINED,
    start_time: 0.0,
    duration_ms: 0.0,
  });
  const [timetrialStatus, setTimetrialStatus] = useState<TimetrialStatus>({
    attempt_status: { status: STATUS.PLAYING, attempts: -1 },
    status: STATUS_TIME.PLAYING,
    next_county: COUNTRY.UNDEFINED,
    duration: 0.0,
  });
  const timetrialISO = timetrial?.target_country;
  const timetrialStatusISO = timetrialStatus?.next_county;

  // Load information when page first loads
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

  const dailyMode = mode === modes[0];
  const inverseMode = mode === modes[1];
  const timetrialMode = mode === modes[2];

  return (
    <main className="relative min-h-screen flex flex-col">
      <AppDialog />
      {dailyMode && <DailyModeTitle />}
      {inverseMode && <InverseModeTitle />}
      {timetrialMode && <TimetrialModeTitle />}
      <div className={dailyMode ? "" : "hidden"}>
        <DailyModeMap countryISO={countryISO} />
      </div>
      <div className={inverseMode ? "" : "hidden"}>
        <InverseModeMap countryISO={inverseISO} setCountryISO={setInverseISO} />
      </div>
      <div className={timetrialMode ? "" : "hidden"}>
        <TimeTrialModeMap
          firstCountryISO={timetrialISO}
          countryISO={timetrialStatusISO}
        />
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
          {timetrialMode && (
            <TimetrialField
              mode={mode}
              setMode={setMode}
              timetrial={timetrial}
              setTimeTrial={setTimetrial}
              timetrialStatus={timetrialStatus}
              setTimetrialStatus={setTimetrialStatus}
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
          {timetrialMode && (
            <TimetrialDrawer
              mode={mode}
              setMode={setMode}
              timetrial={timetrial}
              setTimeTrial={setTimetrial}
              timetrialStatus={timetrialStatus}
              setTimetrialStatus={setTimetrialStatus}
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
        {timetrialMode && timetrialStatus !== null && (
          <>
            <PopUp status={timetrialStatus.attempt_status.status}></PopUp>
          </>
        )}
      </GuessProvider>
      <Toaster position={"top-center"} />
    </main>
  );
}
