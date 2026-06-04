import { GameStatus, MAX_ATTEMPTS, STATUS } from "@/types/gameTypes";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";

type Props = {
  gameStatus: GameStatus;
};

export function Attempts({ gameStatus }: Props) {
  return (
    <div className="bg-gray-100 rounded-sm absolute md:top-30 md:left-15 top-20 left-3 flex flex-row">
      <GameSquares attempts={gameStatus.attempts} status={gameStatus.status} />
    </div>
  );
}

function GameSquares({ attempts, status }: GameStatus) {
  return (
    <>
      {[...Array(MAX_ATTEMPTS)].map((_, i) => {
        if (i >= attempts) {
          return <div key={i} className="w-8 h-8 m-2 rounded-sm bg-gray-200" />;
        }

        return (
          <HoverCard key={i}>
            <HoverCardTrigger>
              <div
                className={`w-8 h-8 m-2 rounded-sm ${
                  i === attempts - 1 && status === STATUS.WON
                    ? "bg-green-200"
                    : "bg-yellow-200"
                }`}
              />
            </HoverCardTrigger>
            <HoverCardContent className="w-fit">
              <div>
                <strong>Guess: </strong> {localStorage.getItem(`guess${i + 1}`)}
              </div>
              <div>
                {localStorage.getItem(`hint${i + 1}`) !== null ? (
                  <>
                    <strong>Hint: </strong>
                    {localStorage.getItem(`hint${i + 1}`)}
                  </>
                ) : (
                  ""
                )}
              </div>
            </HoverCardContent>
          </HoverCard>
        );
      })}
    </>
  );
}
