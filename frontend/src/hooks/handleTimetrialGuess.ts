import { useTimetrialGuess } from "@/context/GuessContext";
import { getDataFromISO } from "@/lib/getCountryData";
import { notify } from "@/lib/notifier";
import { makeTimetrialAttempt } from "@/services/timetrial";
import { TimetrialStatus } from "@/types/gameTypes";
import React from "react";

export function useHandleTimetrialGuess() {
  const { timetrialGuess, setTimetrialGuess } = useTimetrialGuess();

  return async function handleTimetrialGuess(
    genre: string | null,
    setTimetrialStatus: React.Dispatch<React.SetStateAction<TimetrialStatus>>,
    setCountry: React.Dispatch<React.SetStateAction<string>>,
  ) {
    if (!genre) {
      notify.info("Please enter a valid genre");
      return;
    }

    const timetrialStatus = await makeTimetrialAttempt(1, genre);
    if (timetrialStatus) {
      setTimetrialStatus(timetrialStatus);
      setTimetrialGuess(timetrialGuess + 1);
      setCountry(getDataFromISO(timetrialStatus.next_county).name);
    }
  };
}
