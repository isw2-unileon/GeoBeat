import { ReactNode, useState } from "react";
import { GuessContext } from "./GuessContext";

export function GuessProvider({ children }: { children: ReactNode }) {
  const [guess, setGuess] = useState<number>(0);

  return (
    <GuessContext.Provider value={{ guess, setGuess }}>
      {children}
    </GuessContext.Provider>
  );
}
