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
import React, { useState } from "react";
import { GameStatus, STATUS } from "@/types/gameTypes";
import { useHandleGuess } from "@/hooks/handleGuess";
import { GenreCombobox } from "../genre-combobox";
import { ModeSelect } from "../mode-select";
import { useNormalizeGenre } from "@/hooks/useNormalizeGenre";

type Props = {
  country: string;
  mode: string;
  setMode: React.Dispatch<React.SetStateAction<string>>;
  gameStatus: GameStatus;
  setGameStatus: React.Dispatch<React.SetStateAction<GameStatus>>;
};

export function DailyField({
  country,
  mode,
  setMode,
  gameStatus,
  setGameStatus,
}: Props) {
  const [genre, setGenre] = useState<string | null>(null);
  const handleGuess = useHandleGuess();
  const normalizeGuess = useNormalizeGenre();

  return (
    <>
      <FieldSet className="absolute top-4 right-4 p-4 max-w-xs bg-white/80 rounded-md animate-fade-in-left">
        <FieldLegend className="!text-2xl bg-white rounded-md px-2">
          GeoBeat
        </FieldLegend>
        <FieldDescription>
          <strong className="block">{mode}</strong>
          In daily mode you are given a country and have to guess the most
          popular song for it. Every mistake gives a hint in form of a song
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
              ¿What is the most popular genre of?
            </FieldLabel>
            <FieldLabel>{country}</FieldLabel>
            <GenreCombobox genre={genre} setGenre={setGenre} />
          </Field>
          <Field>
            {gameStatus.status !== STATUS.WON &&
            gameStatus.status !== STATUS.LOST ? (
              <Button
                type="button"
                onClick={() => {
                  const normalized_genre = normalizeGuess(genre);
                  handleGuess(normalized_genre, genre, setGameStatus);
                }}
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
