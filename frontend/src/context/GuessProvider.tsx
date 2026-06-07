import { ReactNode, useEffect, useState } from "react";
import {
  DailyGuessContext,
  GenresContext,
  InverseGuessContext,
  TimetrialGuessContext,
} from "./GuessContext";
import { getGenres } from "@/services/genres";

export function GuessProvider({ children }: { children: ReactNode }) {
  const [guess, setGuess] = useState(0);
  const [inverseGuess, setInverseGuess] = useState(0);
  const [timetrialGuess, setTimetrialGuess] = useState(0);
  const [genres, setGenres] = useState([
    { name: "none", normalized_name: "none" },
  ]);

  useEffect(() => {
    async function load() {
      const data = await getGenres();
      if (data) setGenres(data);
    }

    load();
  }, []);

  return (
    <DailyGuessContext.Provider value={{ guess, setGuess }}>
      <InverseGuessContext.Provider value={{ inverseGuess, setInverseGuess }}>
        <TimetrialGuessContext.Provider
          value={{ timetrialGuess, setTimetrialGuess }}
        >
          <GenresContext.Provider value={{ genres, setGenres }}>
            {children}
          </GenresContext.Provider>
        </TimetrialGuessContext.Provider>
      </InverseGuessContext.Provider>
    </DailyGuessContext.Provider>
  );
}
