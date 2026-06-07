import { getLeaderboard } from "@/services/timetrial";
import { Leaderboard } from "@/types/gameTypes";

export function useHandleLeaderboard() {
  return async function handleLeaderboard(
    setLeaderboard: React.Dispatch<React.SetStateAction<Leaderboard>>,
  ) {
    const leaderboard = await getLeaderboard(1);

    if (leaderboard) {
      setLeaderboard(leaderboard);
    }
  };
}
