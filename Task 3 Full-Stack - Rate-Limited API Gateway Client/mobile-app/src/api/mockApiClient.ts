import { ApiClient, ApiRequest } from "./types";

export class MockApiClient implements ApiClient {
  private readonly failuresByKey = new Map<string, number>();
  private readonly waitsByKey = new Map<string, number>();

  setTransientFailures(key: string, failCount: number): void {
    this.failuresByKey.set(key, failCount);
  }

  setDelayMs(key: string, delayMs: number): void {
    this.waitsByKey.set(key, delayMs);
  }

  async request<TData>(request: ApiRequest, signal?: AbortSignal): Promise<TData> {
    const delayMs = this.waitsByKey.get(request.key) ?? 500;
    await this.sleep(delayMs, signal);

    const left = this.failuresByKey.get(request.key) ?? 0;
    if (left > 0) {
      this.failuresByKey.set(request.key, left - 1);
      throw new Error("Transient API error");
    }

    const payload = {
      key: request.key,
      url: request.url,
      at: new Date().toISOString(),
      message: "Gateway response payload",
    };

    return payload as TData;
  }

  private sleep(ms: number, signal?: AbortSignal): Promise<void> {
    return new Promise((resolve, reject) => {
      const timer = setTimeout(resolve, ms);
      signal?.addEventListener("abort", () => {
        clearTimeout(timer);
        reject(new DOMException("Aborted", "AbortError"));
      });
    });
  }
}
