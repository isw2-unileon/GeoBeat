import { useGenres } from "@/context/GuessContext";

export function useNormalizeGenre() {
  const { genres } = useGenres();

  return (name: string | null) => {
    return genres.find((g) => g.name === name)?.normalized_name ?? name;
  };
}
