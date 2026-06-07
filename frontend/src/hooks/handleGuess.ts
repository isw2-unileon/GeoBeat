import { useDailyGuess } from "@/context/GuessContext";
import { notify } from "@/lib/notifier";
import { makeAttempt } from "@/services/daily";
import { GameStatus, STATUS } from "@/types/gameTypes";

export function useHandleGuess() {
  const { guess, setGuess } = useDailyGuess();

  return async function handleGuess(
    normalized_genre: string | null,
    genre: string | null,
    setGameStatus: React.Dispatch<React.SetStateAction<GameStatus>>,
  ) {
    if (!normalized_genre) {
      notify.error("Please enter a valid genre");
      return;
    }

    const attemptResult = await makeAttempt(1, normalized_genre);
    if (attemptResult) {
      const gameStatus = attemptResult.gameStatus;

      setGameStatus({
        attempts: gameStatus.attempts,
        status: gameStatus.status,
      });

      localStorage.setItem(
        `guess${gameStatus.attempts}`,
        genre ?? normalized_genre,
      );
      if (attemptResult.hint) {
        localStorage.setItem(`hint${gameStatus.attempts}`, attemptResult.hint);
        notify.news("New hint added, hover the left squares to see it!");
      }
      if (gameStatus.status == STATUS.WON) {
        localStorage.removeItem(`hint${gameStatus.attempts}`);
      }
      setGuess(guess + 1);
    }
  };
}
