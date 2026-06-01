import { createContext, useContext } from "react";

type GuessContextType = {
  guess: number;
  setGuess: React.Dispatch<React.SetStateAction<number>>;
};

export const GuessContext = createContext<GuessContextType | undefined>(
  undefined,
);

export function useGuess() {
  const context = useContext(GuessContext);

  if (!context) {
    throw new Error("useGuess must be used within GuessProvider");
  }

  return context;
}
