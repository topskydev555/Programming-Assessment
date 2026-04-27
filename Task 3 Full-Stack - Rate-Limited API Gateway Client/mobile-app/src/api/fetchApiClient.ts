import { ApiClient } from "./types";

export const fetchApiClient: ApiClient = async <T>(
  key: string,
  signal?: AbortSignal
): Promise<T> => {
  const response = await fetch(key, { signal });
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }
  return (await response.json()) as T;
};
