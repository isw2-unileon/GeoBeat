import { GameStatus, MAX_ATTEMPTS, STATUS } from "@/types/gameTypes";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";
import { modes } from "@/data/placeholder-data";

type Props = {
  gameStatus: GameStatus;
  mode: string;
};

export function Attempts({ gameStatus, mode }: Props) {
  return (
    <div className="bg-gray-100 rounded-sm absolute md:top-30 md:left-15 top-20 left-3 flex flex-row">
      <GameSquares gameStatus={gameStatus} mode={mode} />
    </div>
  );
}

type SquaresProps = {
  gameStatus: GameStatus;
  mode: string;
};

function GameSquares({ gameStatus, mode }: SquaresProps) {
  return (
    <>
      {[...Array(MAX_ATTEMPTS)].map((_, i) => {
        const storage = getStorage(mode, i);

        if (i >= gameStatus.attempts) {
          return <div key={i} className="w-8 h-8 m-2 rounded-sm bg-gray-200" />;
        }

        return (
          <HoverCard key={i}>
            <HoverCardTrigger>
              <div
                className={`w-8 h-8 m-2 rounded-sm ${
                  i === gameStatus.attempts - 1 &&
                  gameStatus.status === STATUS.WON
                    ? "bg-green-200"
                    : "bg-yellow-200"
                }`}
              />
            </HoverCardTrigger>

            <HoverCardContent className="w-fit">
              <div>
                <strong>Guess: </strong> {storage[0]}
              </div>

              {storage[1] !== null && (
                <div>
                  <strong>Hint: </strong> {storage[1]}
                </div>
              )}
            </HoverCardContent>
          </HoverCard>
        );
      })}
    </>
  );
}

function getStorage(mode: string, i: number) {
  switch (mode) {
    case modes[0]: {
      const hint =
        localStorage.getItem(`hint${i + 1}`) !== null
          ? localStorage.getItem(`hint${i + 1}`)
          : null;
      return [localStorage.getItem(`guess${i + 1}`), hint];
      break;
    }

    case modes[1]: {
      return [localStorage.getItem(`inverse_guess${i + 1}`), null];
      break;
    }

    default: {
      return [null, null];
    }
  }
}
