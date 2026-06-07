import {
  Combobox,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxList,
  ComboboxItem,
  ComboboxContent,
} from "@/components/ui/combobox";
import { useGenres } from "@/context/GuessContext";

type Porps = {
  genre: string | null;
  setGenre: React.Dispatch<React.SetStateAction<string | null>>;
};

export function GenreCombobox({ genre, setGenre }: Porps) {
  const { genres } = useGenres();
  const genre_names = genres.map((g) => g.name);

  return (
    <Combobox items={genre_names} value={genre} onValueChange={setGenre}>
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
