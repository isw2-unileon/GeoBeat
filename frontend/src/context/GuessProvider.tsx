import { ReactNode, useState } from "react";
import {
  DailyGuessContext,
  InverseGuessContext,
  TimetrialGuessContext,
} from "./GuessContext";

export function GuessProvider({ children }: { children: ReactNode }) {
  const [guess, setGuess] = useState(0);
  const [inverseGuess, setInverseGuess] = useState(0);
  const [timetrialGuess, setTimetrialGuess] = useState(0);

  return (
    <DailyGuessContext.Provider value={{ guess, setGuess }}>
      <InverseGuessContext.Provider value={{ inverseGuess, setInverseGuess }}>
        <TimetrialGuessContext.Provider
          value={{ timetrialGuess, setTimetrialGuess }}
        >
          {children}
        </TimetrialGuessContext.Provider>
      </InverseGuessContext.Provider>
    </DailyGuessContext.Provider>
  );
}
