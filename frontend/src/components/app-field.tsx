import {
  FieldLegend,
  FieldSet,
  FieldDescription,
  FieldGroup,
  Field,
  FieldLabel,
  FieldSeparator,
} from "@/components/ui/field";
import {
  Combobox,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxList,
  ComboboxItem,
  ComboboxContent,
} from "@/components/ui/combobox";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { modes, genres } from "@/data/placeholder-data";
import { makeAttempt, getDaily } from "@/services/daily";
import { useState } from "react";
import { notify } from "@/lib/toast";
import { GameStatus } from "@/App";

type Props = {
  country: string;
  setMode: React.Dispatch<React.SetStateAction<string>>;
  setGameStatus: React.Dispatch<React.SetStateAction<GameStatus>>;
};

export function AppField({ country, setMode, setGameStatus }: Props) {
  const [genre, setGenre] = useState<string | null>(null);

  const handleGuess = async () => {
    if (!genre) {
      notify.error("Please enter a valid genre");
      return;
    }

    await makeAttempt(genre);
    const daily = await getDaily();
    if (daily) {
      setGameStatus({
        attempts: daily.attempts,
        status: daily.status,
      });
    }
  };

  return (
    <FieldSet className="absolute top-4 right-4 p-4 max-w-xs bg-white/80 rounded-md animate-fade-in-left">
      <FieldLegend className="!text-2xl bg-white rounded-md px-2">
        {" "}
        GeoBeat{" "}
      </FieldLegend>
      <FieldDescription>
        The not so hit music genre guessing game
      </FieldDescription>
      <FieldGroup>
        <FieldSeparator />
        <Field>
          <FieldLabel className="text-1xl">Mode selection</FieldLabel>
          <Select
            defaultValue={modes[0]}
            onValueChange={(value) => setMode(value)}
          >
            <SelectTrigger className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                {modes.map((mode) => (
                  <SelectItem key={mode} value={mode}>
                    {mode}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </Field>
        <FieldSeparator />
        <Field>
          <FieldLabel className="text-1xl">
            ¿What is the most popular genre of?
          </FieldLabel>
          <FieldLabel>{country}</FieldLabel>
          <Combobox items={genres} value={genre} onValueChange={setGenre}>
            <ComboboxInput placeholder="Select a genre" />
            <ComboboxContent>
              <ComboboxEmpty>No genres available</ComboboxEmpty>
              <ComboboxList>
                {(item: string) => (
                  <ComboboxItem key={item} value={item}>
                    {item}
                  </ComboboxItem>
                )}
              </ComboboxList>
            </ComboboxContent>
          </Combobox>
        </Field>
        <Field>
          <Button type="button" onClick={handleGuess}>
            Guess
          </Button>
        </Field>
      </FieldGroup>
    </FieldSet>
  );
}
