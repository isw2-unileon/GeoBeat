export const MAX_ATTEMPTS = 5;

export const STATUS = {
  WON: "won",
  LOST: "lost",
  PLAYING: "playing",
} as const;

export const STATUS_TIME = {
  COMPLETED: "completed",
  PLAYING: "playing",
};

export type Status = (typeof STATUS)[keyof typeof STATUS];

export type StatusTime = (typeof STATUS_TIME)[keyof typeof STATUS_TIME];

export type Daily = {
  country: string;
  attempts: number;
  status: Status;
} | null;

export type GameStatus = {
  attempts: number;
  status: Status;
};

export type AttemptResult = {
  gameStatus: GameStatus;
  hint: string;
} | null;

export type Inverse = {
  song: string;
  attempts: number;
  status: Status;
} | null;

export type Timetrial = {
  status: StatusTime;
  target_country: string;
  start_time: number;
  duration_ms: number;
} | null;

export type TimetrialStatus = {
  attempt_status: GameStatus;
  status: StatusTime;
  next_county: string;
  duration: number;
} | null;

export type Leaderboard = {
  id: number;
  entries: Entry[];
  user_entry: Entry | null;
} | null;

export type Entry = {
  username: string;
  duration: number;
  rank: number;
};

export type Genre = {
  name: string;
  normalized_name: string;
};

export const COUNTRY = {
  UNDEFINED: "Missing country",
  UNDEFINED_ISO: "UNKNOWN",
};
