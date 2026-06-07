import { retrieveToken } from "@/services/auth";
import { notify } from "./notifier";

export async function handleRespose(res: Response, error_msg: string) {
  try {
    const data = await res.json();

    if (!res.ok) {
      if (res.status === 401) {
        await retrieveToken();
        return { retry: true };
      } else {
        notify.error(error_msg + data.error);
        return null;
      }
    }

    return data;
  } catch {
    return null;
  }
}
