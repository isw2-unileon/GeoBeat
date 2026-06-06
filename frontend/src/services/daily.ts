import { notify } from "@/lib/notifier";
import {
  AttemptResult,
  Daily,
  MAX_ATTEMPTS,
  Status,
  STATUS,
} from "@/types/gameTypes";
import { retrieveToken } from "./auth";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

export async function getDaily(retryNum: number): Promise<Daily> {
  const token = localStorage.getItem("token");

  if (!token) {
    if (retryNum > 0) {
      await retrieveToken();
      return getDaily(retryNum - 1);
    }
    notify.error("Couldn't get daily: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/daily`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const data = await res.json();

    if (!res.ok) {
      if (res.status === 401 && retryNum > 0) {
        await retrieveToken();
        return getDaily(retryNum - 1);
      } else {
        notify.error("Couldn't get daily: " + data.error);
        return null;
      }
    }

    if (!data) {
      notify.error("Couldn't get daily no response data");
      return null;
    }

    if (!Object.values(STATUS).includes(data.status as Status)) {
      notify.error("The structure of the data received was not the expected");
      return null;
    }

    notify.info(data);
    return {
      country: data.country,
      attempts: data.attempts_used,
      status: data.status,
    };
  } catch {
    notify.error("Failed, couldn't retrieve daily");
    return null;
  }
}

export async function makeAttempt(
  guess: string,
  retryNum: number,
): Promise<AttemptResult> {
  const token = localStorage.getItem("token");

  if (!token) {
    if (retryNum > 0) {
      await retrieveToken();
      return makeAttempt(guess, retryNum - 1);
    }
    notify.error("Couldn't make attempt: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/daily/attempt`, {
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
        return makeAttempt(guess, retryNum - 1);
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
      gameStatus: {
        attempts: MAX_ATTEMPTS - data.attempts_remaining,
        status: data.status,
      },
      hint: data.hint,
    };
  } catch {
    notify.error("Failed, couldn't make attempt");
    return null;
  }
}
