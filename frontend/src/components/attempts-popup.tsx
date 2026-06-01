import { GameStatus, STATUS } from "@/types/gameTypes";
import { useGuess } from "@/context/GuessContext";
import { useEffect, useState } from "react";

type PopUpProps = {
  status: GameStatus["status"];
};

// TODO add shadcn hover component and memory of last guesses
export function PopUp({ status }: PopUpProps) {
  const [isHidden, setIsHidden] = useState<boolean>(true);
  const { guess } = useGuess();

  useEffect(() => {
    if (guess === 0) return;

    setIsHidden(false);

    const timer = setTimeout(() => {
      setIsHidden(true);
    }, 1500);

    return () => clearTimeout(timer);
  }, [guess]);

  return (
    !isHidden && (
      <div key={guess}>
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
