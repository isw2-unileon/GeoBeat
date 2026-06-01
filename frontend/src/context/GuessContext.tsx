import { createContext, useContext, useState, ReactNode } from "react";

type GuessContextType = {
  guess: number;
  setGuess: React.Dispatch<React.SetStateAction<number>>;
};

const GuessContext = createContext<GuessContextType | undefined>(undefined);

export function GuessProvider({ children }: { children: ReactNode }) {
  const [guess, setGuess] = useState<number>(0);

  return (
    <GuessContext.Provider value={{ guess, setGuess }}>
      {children}
    </GuessContext.Provider>
  );
}

export function useGuess() {
  const context = useContext(GuessContext);

  if (!context) {
    throw new Error("useGuess must be used within GuessProvider");
  }

  return context;
}
