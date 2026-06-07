import {
  FieldLegend,
  FieldSet,
  FieldDescription,
  FieldGroup,
  Field,
  FieldLabel,
  FieldSeparator,
} from "@/components/ui/field";
import { Button } from "@/components/ui/button";
import {
  COUNTRY,
  STATUS_TIME,
  Timetrial,
  TimetrialStatus,
} from "@/types/gameTypes";
import { ModeSelect } from "@/components/mode-select";
import { useState } from "react";
import { GenreCombobox } from "../genre-combobox";
import { useHandleTimetrialGuess } from "@/hooks/handleTimetrialGuess";
import { useHandleStartTimetrial } from "@/hooks/handleStartTimetrial";
import { LeaderboardDialog } from "./leaderboard-dialog";

type Props = {
  mode: string;
  setMode: React.Dispatch<React.SetStateAction<string>>;
  timetrial: Timetrial;
  setTimeTrial: React.Dispatch<React.SetStateAction<Timetrial>>;
  timetrialStatus: TimetrialStatus;
  setTimetrialStatus: React.Dispatch<React.SetStateAction<TimetrialStatus>>;
};

export function TimetrialField({
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

  return (
    <>
      <FieldSet className="absolute top-4 right-4 p-4 max-w-xs bg-white/80 rounded-md animate-fade-in-left">
        <FieldLegend className="!text-2xl bg-white rounded-md px-2">
          GeoBeat
        </FieldLegend>
        <FieldDescription>
          <strong className="block">{mode}</strong>
          In this mode you have infite guesses, start the mode by clicking on
          "Start". After that you will have to guess the top genre for five
          countries. This mode has score depending on the time, so try to be
          fast!
        </FieldDescription>
        <FieldGroup>
          <FieldSeparator />
          <Field>
            <FieldLabel className="text-1xl">Mode selection</FieldLabel>
            <ModeSelect mode={mode} setMode={setMode} />
          </Field>
          <FieldSeparator />
          <Field>
            <LeaderboardDialog />
          </Field>
          <FieldSeparator />
          <Field>
            <FieldLabel className="text-1xl">
              ¿What is the most popular genre of?
            </FieldLabel>
            <FieldLabel>{country}</FieldLabel>
            <GenreCombobox genre={genre} setGenre={setGenre} />
          </Field>
          <Field>
            {hasStarted &&
              (timetrial?.status !== STATUS_TIME.COMPLETED &&
              timetrialStatus?.status !== STATUS_TIME.COMPLETED ? (
                <Button
                  type="button"
                  onClick={() => {
                    handleTimetrial(genre, setTimetrialStatus, setCountry);
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
          </Field>
        </FieldGroup>
      </FieldSet>
    </>
  );
}
