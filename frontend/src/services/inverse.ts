import { notify } from "@/lib/notifier";
import {
  GameStatus,
  Inverse,
  Status,
  STATUS,
  MAX_ATTEMPTS,
} from "@/types/gameTypes";
import { retrieveToken } from "./auth";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

export async function getInverse(retryNum: number): Promise<Inverse> {
  const token = localStorage.getItem("token");

  if (!token) {
    if (retryNum > 0) {
      await retrieveToken();
      return getInverse(retryNum - 1);
    }
    notify.error("Couldn't get guess the country mode data: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/inverse`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const data = await res.json();

    if (!res.ok) {
      if (res.status === 401 && retryNum > 0) {
        await retrieveToken();
        getInverse(retryNum);
      } else {
        notify.error("Couldn't get guess the country mode data: " + data.error);
      }
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
  retryNum: number,
): Promise<GameStatus | null> {
  const token = localStorage.getItem("token");

  if (!token) {
    if (retryNum > 0) {
      await retrieveToken();
      return makeInverseAttempt(guess, retryNum - 1);
    }
    notify.error("Couldn't make attempt: missing token");
    return null;
  }

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

    const data = await res.json();

    if (!res.ok) {
      if (res.status === 401 && retryNum > 0) {
        await retrieveToken();
        return makeInverseAttempt(guess, retryNum - 1);
      } else {
        notify.error("Couldn't make attempt: " + data.error);
        return null;
      }
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
