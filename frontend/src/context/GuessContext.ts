import { Genre } from "@/types/gameTypes";
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

type TimetrialGuessContextType = {
  timetrialGuess: number;
  setTimetrialGuess: React.Dispatch<React.SetStateAction<number>>;
};

export const TimetrialGuessContext = createContext<
  TimetrialGuessContextType | undefined
>(undefined);

export function useTimetrialGuess() {
  const context = useContext(TimetrialGuessContext);
  if (!context) {
    throw new Error(
      "useTimetrialGuess must be used within InverseGuessProvider",
    );
  }
  return context;
}

type GenresContextType = {
  genres: Genre[];
  setGenres: React.Dispatch<React.SetStateAction<Genre[]>>;
};

export const GenresContext = createContext<GenresContextType | undefined>(
  undefined,
);

export function useGenres() {
  const context = useContext(GenresContext);
  if (!context) {
    throw new Error("useGenres must be used within GenresProvider");
  }
  return context;
}
