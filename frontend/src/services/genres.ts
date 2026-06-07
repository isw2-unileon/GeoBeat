import { notify } from "@/lib/notifier";
import { Genre } from "@/types/gameTypes";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

export async function getGenres(): Promise<Genre[] | null> {
  try {
    const res = await fetch(`${BACKEND_URL}/api/genres`);

    const data: Genre[] = await res.json();

    if (!data) {
      notify.info("No genre data obtained");
      return null;
    }

    if (!res.ok) {
      notify.info("Failed to retrive genres " + res.status);
      return null;
    }

    const names = data.map((g) => ({
      name: g.name,
      normalized_name: g.normalized_name,
    }));

    console.log(names);
    notify.info("Genres retrieved succesfully");
    return names;
  } catch {
    notify.info("Failed to retrive genres");
    return null;
  }
}
