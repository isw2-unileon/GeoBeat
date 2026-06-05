import { createContext, useContext } from "react";

type DailyGuessContextType = {
  guess: number;
  setGuess: React.Dispatch<React.SetStateAction<number>>;
};

export const DailyGuessContext = createContext<
  DailyGuessContextType | undefined
>(undefined);

export function useDailyGuess() {
  const context = useContext(DailyGuessContext);
  if (!context) {
    throw new Error("useDailyGuess must be used within DailyGuessProvider");
  }
  return context;
}

type InverseGuessContextType = {
  inverseGuess: number;
  setInverseGuess: React.Dispatch<React.SetStateAction<number>>;
};

export const InverseGuessContext = createContext<
  InverseGuessContextType | undefined
>(undefined);

export function useInverseGuess() {
  const context = useContext(InverseGuessContext);
  if (!context) {
    throw new Error("useInverseGuess must be used within InverseGuessProvider");
  }
  return context;
}
