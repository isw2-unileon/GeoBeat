import { notify } from "@/lib/notifier";
import { Inverse, Status, STATUS } from "@/types/gameTypes";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

export async function getInverse(): Promise<Inverse> {
  const token = localStorage.getItem("token");

  if (!token) {
    notify.error("Couldn't get daily: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/inverse`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    let data;

    const text = await res.text();

    if (text) {
      data = JSON.parse(text);
    }

    if (!res.ok) {
      notify.error("Couldn't get mode data: " + data.error);
      return null;
    }

    if (!data) {
      notify.error("Couldn't get mode data, no response");
      return null;
    }

    if (!Object.values(STATUS).includes(data.status as Status)) {
      notify.error("The structure of the data received was not the expected");
      return null;
    }

    notify.info(data);
    return {
      song: data.song,
      attempts: data.attempts,
      status: data.status,
    };
  } catch {
    notify.error("Failed, couldn't retrieve the mode data");
    return null;
  }
}
