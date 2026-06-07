import { GameStatus, STATUS } from "@/types/gameTypes";
import {
  useDailyGuess,
  useInverseGuess,
  useTimetrialGuess,
} from "@/context/GuessContext";
import React, { useEffect, useRef, useState } from "react";

type PopUpProps = {
  status: GameStatus["status"];
};

// TODO add shadcn hover component and memory of last guesses
export function PopUp({ status }: PopUpProps) {
  const [isHidden, setIsHidden] = useState<boolean>(true);

  const { guess } = useDailyGuess();
  const { inverseGuess } = useInverseGuess();
  const { timetrialGuess } = useTimetrialGuess();

  const prevGuessRef = useRef<number | null>(null);
  const prevInverseRef = useRef<number | null>(null);
  const prevTimetrialRef = useRef<number | null>(null);

  // TODO discutir mejor método
  useEffect(() => {
    updateGuess(prevGuessRef, guess, setIsHidden);
  }, [guess]);

  useEffect(() => {
    updateGuess(prevInverseRef, inverseGuess, setIsHidden);
  }, [inverseGuess]);

  useEffect(() => {
    updateGuess(prevTimetrialRef, timetrialGuess, setIsHidden);
  }, [timetrialGuess]);

  return (
    !isHidden && (
      <div>
        {status === STATUS.WON ? (
          <label className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 animate-pop-fade text-6xl">
            ✅
          </label>
        ) : (
          <label className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 animate-pop-fade text-6xl">
            ❌
          </label>
        )}
      </div>
    )
  );
}

function updateGuess(
  prevGuessRef: React.RefObject<number | null>,
  guess: number,
  setIsHidden: React.Dispatch<React.SetStateAction<boolean>>,
) {
  if (prevGuessRef.current === null) {
    prevGuessRef.current = guess;
    return;
  }

  if (guess === prevGuessRef.current) return;

  prevGuessRef.current = guess;
  setIsHidden(false);

  const timer = setTimeout(() => {
    setIsHidden(true);
  }, 1500);

  return () => clearTimeout(timer);
}
