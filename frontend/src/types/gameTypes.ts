export const MAX_ATTEMPTS = 5;

export const STATUS = {
  WON: "won",
  LOST: "lost",
  PLAYING: "playing",
} as const;

export type Status = (typeof STATUS)[keyof typeof STATUS];

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
