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
