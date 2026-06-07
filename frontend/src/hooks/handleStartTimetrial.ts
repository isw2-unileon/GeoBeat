import { getDataFromISO } from "@/lib/getCountryData";
import { notify } from "@/lib/notifier";
import { startGame } from "@/services/timetrial";
import { STATUS_TIME, Timetrial } from "@/types/gameTypes";
import React from "react";

export function useHandleStartTimetrial() {
  return async function handleStartTimetrial(
    setTimtetrial: React.Dispatch<React.SetStateAction<Timetrial>>,
    setHasStarted: React.Dispatch<React.SetStateAction<boolean>>,
    setCountry: React.Dispatch<React.SetStateAction<string>>,
  ) {
    const timetrial = await startGame(1);

    if (timetrial) {
      if (timetrial.status === STATUS_TIME.COMPLETED) {
        notify.news("It seems the game has already been completed");
      } else {
        notify.news("Game started!");
      }
      setTimtetrial(timetrial);
      setHasStarted(true);
      setCountry(getDataFromISO(timetrial.target_country).name);
    }
  };
}
