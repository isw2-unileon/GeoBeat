import { GameStatus } from "@/types/gameTypes";
import { useInverseGuess } from "@/context/GuessContext";
import { makeInverseAttempt } from "@/services/inverse";

export function useHandleInverseGuess() {
  const { inverseGuess, setInverseGuess } = useInverseGuess();

  return async function handleInvrseGuess(
    countryISO: string,
    country: string,
    setGameStatus: React.Dispatch<React.SetStateAction<GameStatus>>,
  ) {
    const attemptResult = await makeInverseAttempt(countryISO);
    if (attemptResult) {
      setGameStatus({
        attempts: attemptResult.attempts,
        status: attemptResult.status,
      });

      localStorage.setItem(`inverse_guess${attemptResult.attempts}`, country);

      setInverseGuess(inverseGuess + 1);
    }
  };
}
