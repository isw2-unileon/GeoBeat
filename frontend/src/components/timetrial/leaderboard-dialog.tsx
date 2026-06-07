import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogClose,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Entry, Leaderboard } from "@/types/gameTypes";
import { useHandleLeaderboard } from "@/hooks/handleLeaderboard";
import { useState } from "react";
import { Field, FieldGroup, FieldSeparator } from "@/components/ui/field";

export function LeaderboardDialog() {
  const [leaderboard, setLeaderboard] = useState<Leaderboard>(null);
  const handleLeaderboard = useHandleLeaderboard();

  return (
    <Dialog
      onOpenChange={(open) => {
        if (open) {
          handleLeaderboard(setLeaderboard);
        }
      }}
    >
      <DialogTrigger asChild={true}>
        <Button variant={"outline"}>Show Leaderboard</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Leaderboard</DialogTitle>
          <DialogDescription>Today's top scores</DialogDescription>
        </DialogHeader>
        <div className="max-h-[400px] overflow-y-auto pr-2">
          {leaderboard && <Entries entries={leaderboard.entries} />}
        </div>
        <DialogFooter className="flex flex-wrap">
          {leaderboard?.user_entry && (
            <p className="w-full">
              <strong>Username: </strong>
              {leaderboard.user_entry.username}
              <strong> Duration: </strong>
              {leaderboard.user_entry.duration} segundos
              <strong> Rank: </strong>
              {leaderboard.user_entry.rank}
            </p>
          )}
          <DialogClose asChild>
            <Button className="w-full" variant="outline">
              Close
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

type EntriesProps = {
  entries: Entry[];
};

function Entries({ entries }: EntriesProps) {
  if (!entries) {
    return <span>No entries for today's challenge</span>;
  }

  return (
    <FieldGroup>
      {entries.map((entry, index) => (
        <div key={index}>
          <Field>
            <strong>Username: </strong> {entry.username}
            <strong> Duration: </strong> {entry.duration} segundos
            <strong> Rank: </strong> {entry.rank}
          </Field>
          <FieldSeparator />
        </div>
      ))}
    </FieldGroup>
  );
}
