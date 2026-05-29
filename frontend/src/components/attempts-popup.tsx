import { GameStatus, STATUS } from "@/types/game";

type PopUpProps = {
  status: GameStatus["status"];
};

// TODO add shadcn hover component and memory of last guesses
export function PopUp({ status }: PopUpProps) {
  return status === STATUS.WON ? (
    <label className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 animate-pop-fade text-6xl">
      ✅
    </label>
  ) : (
    <label className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 animate-pop-fade text-6xl">
      ❌
    </label>
  );
}
