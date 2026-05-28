import { GameStatus } from "@/App";

type Props = {
  gameStatus: GameStatus;
};

export function Attempts({ gameStatus }: Props) {
  return (
    <div className="bg-gray-100 rounded-sm absolute top-30 left-15 flex flex-row">
      <GameSquares attempts={gameStatus.attempts} status={gameStatus.status} />
    </div>
  );
}

// TODO
function GameSquares({ attempts, status }: GameStatus) {
  return (
    <>
      {[...Array(5)].map((_, i) => (
        <div
          key={i}
          className={`w-8 h-8 m-2 rounded-sm ${
            i < attempts
              ? i === attempts - 1 && status === "won"
                ? "bg-green-200"
                : "bg-yellow-200"
              : "bg-gray-200"
          }`}
        />
      ))}
    </>
  );
}
