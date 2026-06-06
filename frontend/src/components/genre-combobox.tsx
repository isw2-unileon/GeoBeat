import {
  Combobox,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxList,
  ComboboxItem,
  ComboboxContent,
} from "@/components/ui/combobox";

import { genres } from "@/data/constats";

type Porps = {
  genre: string | null;
  setGenre: React.Dispatch<React.SetStateAction<string | null>>;
};

export function GenreCombobox({ genre, setGenre }: Porps) {
  return (
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
  );
}
