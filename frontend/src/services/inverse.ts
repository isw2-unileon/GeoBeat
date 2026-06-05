import { notify } from "@/lib/notifier";
import {
  GameStatus,
  Inverse,
  Status,
  STATUS,
  MAX_ATTEMPTS,
} from "@/types/gameTypes";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

export async function getInverse(): Promise<Inverse> {
  const token = localStorage.getItem("token");

  if (!token) {
    notify.error("Couldn't get inverse: missing token");
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
      notify.error("Couldn't get guess the country mode data: " + data.error);
      return null;
    }

    if (!data) {
      notify.error("Couldn't get guess the country mode data, no response");
      return null;
    }

    if (!Object.values(STATUS).includes(data.status as Status)) {
      notify.error("The structure of the data received was not the expected");
      return null;
    }

    notify.info(data);
    return {
      song: data.song,
      attempts: data.attempts_used,
      status: data.status,
    };
  } catch {
    notify.error("Failed, couldn't retrieve the guess the country mode data");
    return null;
  }
}

export async function makeInverseAttempt(
  guess: string,
): Promise<GameStatus | null> {
  const token = localStorage.getItem("token");

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/inverse/attempt`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        guess: guess,
      }),
    });

    const text = await res.text();
    let data;

    if (text) {
      data = JSON.parse(text);
    }

    if (!res.ok) {
      notify.error("Couldn't make attempt: " + data.error);
      return null;
    }

    if (!Object.values(STATUS).includes(data.status as Status)) {
      notify.error("The structure of the data received was not the expected");
      return null;
    }

    notify.info(data);
    return {
      attempts: MAX_ATTEMPTS - data.attempts_remaining,
      status: data.status,
    };
  } catch {
    notify.error("Failed, couldn't make attempt");
    return null;
  }
}
