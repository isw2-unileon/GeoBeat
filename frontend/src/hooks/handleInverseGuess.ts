import { COUNTRY, GameStatus } from "@/types/gameTypes";
import { useInverseGuess } from "@/context/GuessContext";
import { makeInverseAttempt } from "@/services/inverse";
import { notify } from "@/lib/notifier";

export function useHandleInverseGuess() {
  const { inverseGuess, setInverseGuess } = useInverseGuess();

  return async function handleInvrseGuess(
    countryISO: string,
    country: string,
    setGameStatus: React.Dispatch<React.SetStateAction<GameStatus>>,
  ) {
    if (countryISO === COUNTRY.UNDEFINED_ISO) {
      notify.error("Please select a country");
      return;
    }

    const attemptResult = await makeInverseAttempt(1, countryISO);
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
