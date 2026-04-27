import { ApiClient, ApiRequest } from "./types";

export class FetchApiClient implements ApiClient {
  async request<TData>(request: ApiRequest, signal?: AbortSignal): Promise<TData> {
    const response = await fetch(request.url, {
      method: request.method ?? "GET",
      headers: {
        "Content-Type": "application/json",
        ...(request.headers ?? {}),
      },
      body: request.body != null ? JSON.stringify(request.body) : undefined,
      signal,
    });

    if (!response.ok) {
      throw new Error(`Request failed with status ${response.status}`);
    }

    return (await response.json()) as TData;
  }
}
