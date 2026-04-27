import { ApiClient, ApiRequest, CacheEntry } from "../api/types";

export type RetryOptions = {
  maxAttempts: number;
  baseDelayMs: number;
};

export type ManagerOptions = {
  ttlMs: number;
  retry: RetryOptions;
  now?: () => number;
};

type InFlight<TData> = {
  promise: Promise<TData>;
  controller: AbortController;
};

export class ApiRequestManager {
  private readonly client: ApiClient;
  private readonly ttlMs: number;
  private readonly retry: RetryOptions;
  private readonly now: () => number;
  private readonly cache = new Map<string, CacheEntry<unknown>>();
  private readonly inFlight = new Map<string, InFlight<unknown>>();

  constructor(client: ApiClient, options: ManagerOptions) {
    this.client = client;
    this.ttlMs = options.ttlMs;
    this.retry = options.retry;
    this.now = options.now ?? (() => Date.now());
  }

  getCached<TData>(key: string): TData | null {
    const entry = this.cache.get(key);
    if (!entry) return null;
    if (this.now() > entry.expiresAt) {
      this.cache.delete(key);
      return null;
    }
    return entry.value as TData;
  }

  async run<TData>(request: ApiRequest): Promise<{ data: TData; fromCache: boolean }> {
    const cached = this.getCached<TData>(request.key);
    if (cached != null) {
      return { data: cached, fromCache: true };
    }

    const existing = this.inFlight.get(request.key);
    if (existing) {
      const data = (await existing.promise) as TData;
      return { data, fromCache: false };
    }

    const controller = new AbortController();
    const promise = this.executeWithRetry<TData>(request, controller.signal)
      .then((data) => {
        this.cache.set(request.key, {
          value: data,
          expiresAt: this.now() + this.ttlMs,
        });
        return data;
      })
      .finally(() => {
        this.inFlight.delete(request.key);
      });

    this.inFlight.set(request.key, { promise, controller });
    const data = await promise;
    return { data, fromCache: false };
  }

  cancel(key?: string): void {
    if (key) {
      const item = this.inFlight.get(key);
      item?.controller.abort();
      return;
    }

    for (const item of this.inFlight.values()) {
      item.controller.abort();
    }
  }

  private async executeWithRetry<TData>(request: ApiRequest, signal: AbortSignal): Promise<TData> {
    let lastErr: unknown;
    for (let attempt = 1; attempt <= this.retry.maxAttempts; attempt++) {
      try {
        return await this.client.request<TData>(request, signal);
      } catch (error) {
        lastErr = error;
        if (signal.aborted || attempt === this.retry.maxAttempts) {
          throw error;
        }
        const delay = this.retry.baseDelayMs * Math.pow(2, attempt - 1);
        await this.sleep(delay, signal);
      }
    }
    throw lastErr;
  }

  private sleep(ms: number, signal: AbortSignal): Promise<void> {
    return new Promise((resolve, reject) => {
      const timer = setTimeout(resolve, ms);
      signal.addEventListener("abort", () => {
        clearTimeout(timer);
        reject(new DOMException("Aborted", "AbortError"));
      });
    });
  }
}
