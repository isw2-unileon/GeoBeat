import { notify } from "@/lib/notifier";
import {
  AttemptResult,
  Daily,
  GameStatus,
  MAX_ATTEMPTS,
  Status,
  STATUS,
} from "@/types/gameTypes";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

async function getDaily(): Promise<Daily> {
  const token = localStorage.getItem("token");

  if (!token) {
    notify.error("Couldn't get daily: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/daily`, {
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
      notify.error("Couldn't get daily: " + data.error);
      return null;
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
    notify.error("Network error couldn't retrieve daily");
    return null;
  }
}

async function makeAttempt(guess: string): Promise<AttemptResult> {
  const token = localStorage.getItem("token");

  if (!token) {
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
      gameStatus: {
        attempts: MAX_ATTEMPTS - data.attempts_remaining,
        status: data.status,
      },
      hint: data.hint,
    };
  } catch {
    notify.error("Network error couldn't make attempt");
    return null;
  }
}

export { getDaily, makeAttempt };
