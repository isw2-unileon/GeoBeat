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
import { GameStatus, STATUS } from "@/types/gameTypes";
import { ModeSelect } from "@/components/mode-select";
import { useHandleInverseGuess } from "@/hooks/handleInverseGuess";
import { getDataFromISO } from "@/lib/getCountryData";

type Props = {
  song: string;
  countryISO: string;
  mode: string;
  setMode: React.Dispatch<React.SetStateAction<string>>;
  inverseStatus: GameStatus;
  setInverseStatus: React.Dispatch<React.SetStateAction<GameStatus>>;
};

export function InverseField({
  song,
  countryISO,
  mode,
  setMode,
  inverseStatus,
  setInverseStatus,
}: Props) {
  const handleInverseGuess = useHandleInverseGuess();
  const country = getDataFromISO(countryISO).name;

  return (
    <>
      <FieldSet className="absolute top-4 right-4 p-4 max-w-xs bg-white/80 rounded-md animate-fade-in-left">
        <FieldLegend className="!text-2xl bg-white rounded-md px-2">
          GeoBeat
        </FieldLegend>
        <FieldDescription>
          <strong className="block">{mode}</strong>
          In guess the country mode you are given a song and have to guess where
          that song is from, in this mode the are no hints!
        </FieldDescription>
        <FieldGroup>
          <FieldSeparator />
          <Field>
            <FieldLabel className="text-1xl">Mode selection</FieldLabel>
            <ModeSelect mode={mode} setMode={setMode} />
          </Field>
          <FieldSeparator />
          <Field>
            <FieldLabel className="text-1xl">
              ¿Where is this song from?
            </FieldLabel>
            <FieldLabel>{song}</FieldLabel>
            <span className="text-sm">Selected country: {country}</span>
          </Field>
          <Field>
            {inverseStatus.status !== STATUS.WON &&
            inverseStatus.status !== STATUS.LOST ? (
              <Button
                type="button"
                onClick={() =>
                  handleInverseGuess(countryISO, country, setInverseStatus)
                }
              >
                Guess
              </Button>
            ) : (
              <Button disabled>Guess</Button>
            )}
          </Field>
        </FieldGroup>
      </FieldSet>
    </>
  );
}
