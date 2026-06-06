import { toast } from "sonner";

export const notify = {
  error: (msg: string) => {
    console.log("[ERROR]", msg);
    toast.error(msg);
  },

  // Info is not displayed to user
  info: (msg: string) => {
    console.log("[INFO]", msg);
  },

  news: (msg: string) => {
    toast.info(msg);
  },
};
