import { ApiClient } from "./types";

type DemoPayload = {
  id: string;
  value: string;
  fetchedAt: string;
};

let failureCountdown = 1;

export function setMockFailuresBeforeSuccess(count: number) {
  failureCountdown = count;
}

export const mockApiClient: ApiClient = async <T>(
  key: string,
  signal?: AbortSignal
): Promise<T> => {
  await new Promise<void>((resolve, reject) => {
    const timer = setTimeout(() => resolve(), 150);
    signal?.addEventListener("abort", () => {
      clearTimeout(timer);
      const error = new Error("Aborted");
      error.name = "AbortError";
      reject(error);
    });
  });

  if (failureCountdown > 0) {
    failureCountdown -= 1;
    throw new Error("Temporary upstream error");
  }

  const payload: DemoPayload = {
    id: key,
    value: "Demo payload",
    fetchedAt: new Date().toISOString(),
  };

  return payload as T;
};
