import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import {
  COUNTRY,
  STATUS_TIME,
  Timetrial,
  TimetrialStatus,
} from "@/types/gameTypes";
import { GenreCombobox } from "@/components/genre-combobox";
import { ModeSelect } from "@/components/mode-select";
import { LeaderboardDialog } from "./leaderboard-dialog";
import { useHandleTimetrialGuess } from "@/hooks/handleTimetrialGuess";
import { useHandleStartTimetrial } from "@/hooks/handleStartTimetrial";
import { useNormalizeGenre } from "@/hooks/useNormalizeGenre";

type Props = {
  mode: string;
  setMode: React.Dispatch<React.SetStateAction<string>>;
  timetrial: Timetrial;
  setTimeTrial: React.Dispatch<React.SetStateAction<Timetrial>>;
  timetrialStatus: TimetrialStatus;
  setTimetrialStatus: React.Dispatch<React.SetStateAction<TimetrialStatus>>;
};

export function TimetrialDrawer({
  mode,
  setMode,
  timetrial,
  setTimeTrial,
  timetrialStatus,
  setTimetrialStatus,
}: Props) {
  const [country, setCountry] = useState<string>(
    timetrial?.target_country ?? COUNTRY.UNDEFINED,
  );
  const [genre, setGenre] = useState<string | null>(null);
  const [hasStarted, setHasStarted] = useState<boolean>(false);
  const handleTimetrial = useHandleTimetrialGuess();
  const handleStartTimetrial = useHandleStartTimetrial();
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const normalizeGuess = useNormalizeGenre();

  return (
    <Drawer open={isOpen} onOpenChange={setIsOpen} modal={false}>
      <DrawerTrigger className="absolute top-10 right-4" asChild>
        <Button className="bg-white/80 text-black">Menu</Button>
      </DrawerTrigger>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle className="text-xl">GeoBeat</DrawerTitle>
          <DrawerDescription className="max-w-xs mx-auto">
            <strong className="block">{mode}</strong>
            In this mode you have infite guesses, start the mode by clicking on
            "Start". After that you will have to guess the top genre for five
            countries. This mode has score depending on the time, so try to be
            fast!
          </DrawerDescription>
        </DrawerHeader>
        <div className="text-center mb-4">
          <h1 className="text-base">Mode selection</h1>
          <div className="max-w-xs mx-auto">
            <ModeSelect mode={mode} setMode={setMode} setIsOpen={setIsOpen} />
          </div>
        </div>
        <div className="max-w-xs mx-auto pb-3">
          <LeaderboardDialog />
        </div>
        <div className="text-center mb-4">
          <h1 className="text-base">What is the most popular genre of?</h1>
          <label>{country}</label>
          <div className="max-w-xs mx-auto">
            <GenreCombobox genre={genre} setGenre={setGenre} />
          </div>
        </div>
        <div className="w-full flex justify-center">
          {hasStarted &&
            (timetrial?.status !== STATUS_TIME.COMPLETED &&
            timetrialStatus?.status !== STATUS_TIME.COMPLETED ? (
              <Button
                type="button"
                onClick={() => {
                  const normalized_genre = normalizeGuess(genre);
                  handleTimetrial(
                    normalized_genre,
                    setTimetrialStatus,
                    setCountry,
                  );
                }}
              >
                Guess
              </Button>
            ) : (
              <Button disabled>Guess</Button>
            ))}
          {!hasStarted && (
            <Button
              type="button"
              onClick={() => {
                handleStartTimetrial(setTimeTrial, setHasStarted, setCountry);
              }}
            >
              Start
            </Button>
          )}
        </div>
        <DrawerFooter>
          <DrawerClose asChild>
            <Button type="button" variant="outline" className="mx-auto">
              Close
            </Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
