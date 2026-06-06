import { useDailyGuess } from "@/context/GuessContext";
import { notify } from "@/lib/notifier";
import { makeAttempt } from "@/services/daily";
import { GameStatus } from "@/types/gameTypes";

export function useHandleGuess() {
  const { guess, setGuess } = useDailyGuess();

  return async function handleGuess(
    genre: string | null,
    setGameStatus: React.Dispatch<React.SetStateAction<GameStatus>>,
  ) {
    if (!genre) {
      notify.error("Please enter a valid genre");
      return;
    }

    const attemptResult = await makeAttempt(genre, 1);
    if (attemptResult) {
      const gameStatus = attemptResult.gameStatus;

      setGameStatus({
        attempts: gameStatus.attempts,
        status: gameStatus.status,
      });

      localStorage.setItem(`guess${gameStatus.attempts}`, genre);
      if (attemptResult.hint) {
        localStorage.setItem(`hint${gameStatus.attempts}`, attemptResult.hint);
        notify.news("New hint added, hover the left squares to see it!");
      }

      setGuess(guess + 1);
    }
  };
}
