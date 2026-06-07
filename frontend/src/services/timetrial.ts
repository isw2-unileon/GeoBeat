import {
  COUNTRY,
  Leaderboard,
  STATUS,
  Timetrial,
  TimetrialStatus,
  StatusTime,
  STATUS_TIME,
  Entry,
} from "@/types/gameTypes";
import { retrieveToken } from "./auth";
import { notify } from "@/lib/notifier";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

async function getTimetrialStatus(retryNum: number): Promise<Timetrial> {
  const token = localStorage.getItem("token");

  if (!token) {
    if (retryNum > 0) {
      await retrieveToken();
      return getTimetrialStatus(retryNum - 1);
    }
    notify.error("Couldn't retrieve time trial data: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/timetrial/status`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const data = await res.json();

    if (!res.ok) {
      if (res.status === 401 && retryNum > 0) {
        await retrieveToken();
        return getTimetrialStatus(retryNum - 1);
      } else {
        notify.error("Couldn't retrieve time trial data: " + data.error);
        return null;
      }
    }

    if (!data) {
      notify.error("Couldn't get retrieve time trial data no response");
      return null;
    }

    if (!Object.values(STATUS_TIME).includes(data.status as StatusTime)) {
      notify.error("The structure of the data received was not the expected");
      return null;
    }

    notify.info(data);
    return {
      status: data.status,
      target_country: data.target_country ?? COUNTRY.UNDEFINED,
      start_time: data.start_time,
      duration_ms: data.duration_ms ?? 0.0,
    };
  } catch {
    notify.error("Failed, couldn't retrieve time trial data");
    return null;
  }
}

async function startGame(retryNum: number): Promise<Timetrial> {
  const token = localStorage.getItem("token");

  if (!token) {
    if (retryNum > 0) {
      await retrieveToken();
      return startGame(retryNum - 1);
    }
    notify.error("Couldn't start time trial: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/timetrial/start`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const data = await res.json();

    if (!res.ok) {
      if (res.status === 401 && retryNum > 0) {
        await retrieveToken();
        return startGame(retryNum - 1);
      } else {
        notify.error("Couldn't start time trial: " + data.error);
        return null;
      }
    }

    if (!data) {
      notify.error("Couldn't get retrieve time trial response");
      return null;
    }

    if (!Object.values(STATUS_TIME).includes(data.status as StatusTime)) {
      notify.error("The structure of the data received was not the expected");
      return null;
    }

    notify.info(data);
    return {
      status: data.status,
      target_country: data.target_country ?? COUNTRY.UNDEFINED,
      start_time: data.start_time,
      duration_ms: data.duration_ms ?? 0.0,
    };
  } catch {
    notify.error("Failed, couldn't start time trial data");
    return null;
  }
}

async function makeTimetrialAttempt(
  retryNum: number,
  guess: string,
): Promise<TimetrialStatus> {
  const token = localStorage.getItem("token");

  if (!token) {
    if (retryNum > 0) {
      await retrieveToken();
      return makeTimetrialAttempt(retryNum - 1, guess);
    }
    notify.error("Couldn't make attempt: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/timetrial/attempt`, {
      method: "POST",
      headers: {
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
        return makeTimetrialAttempt(retryNum - 1, guess);
      } else {
        notify.error("Couldn't make attempt: " + data.error);
        return null;
      }
    }

    if (!data) {
      notify.error("Couldn't get attempt response");
      return null;
    }

    if (!Object.values(STATUS_TIME).includes(data.status as StatusTime)) {
      notify.error("The structure of the data received was not the expected");
      return null;
    }

    notify.info(data);
    return {
      attempt_status: data.correct
        ? { status: STATUS.WON, attempts: null }
        : { status: STATUS.PLAYING, attempts: null },
      status: data.status,
      next_county: data.next_country ?? COUNTRY.UNDEFINED,
      duration: data.duration ?? 0.0,
    };
  } catch {
    notify.error("Failed, couldn't make time trial attempt");
    return null;
  }
}

async function getLeaderboard(retryNum: number): Promise<Leaderboard> {
  const token = localStorage.getItem("token");

  if (!token) {
    if (retryNum > 0) {
      await retrieveToken();
      return getLeaderboard(retryNum - 1);
    }
    notify.error("Couldn't make attempt: missing token");
    return null;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/game/timetrial/leaderboard`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const data = await res.json();

    if (!res.ok) {
      if (res.status === 401 && retryNum > 0) {
        await retrieveToken();
        return getLeaderboard(retryNum - 1);
      } else {
        notify.error("Couldn't get ladderboard: " + data.error);
        return null;
      }
    }

    if (!data) {
      notify.error("Couldn't get attempt response");
      return null;
    }

    const entries = data.entries.map((entry: Entry) => ({
      ...entry,
      duration: entry.duration / 1000,
    }));

    let user_entry = null;
    if (data.user_entry) {
      user_entry = {
        ...data.user_entry,
        duration: data.user_entry.duration / 1000,
      };
    }

    notify.info(data);
    return {
      id: data.challenge_id,
      entries: entries,
      user_entry: user_entry,
    };
  } catch {
    notify.error("Failed, couldn't retrieve leaderboard");
    return null;
  }
}

export { getTimetrialStatus, startGame, makeTimetrialAttempt, getLeaderboard };
