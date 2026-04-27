import { toApiError } from "../api/errors";
import { ApiClient, RequestState } from "../api/types";

type RetryConfig = {
  maxAttempts: number;
  baseDelayMs: number;
  delayFn?: (ms: number) => Promise<void>;
};

type CacheConfig = {
  ttlMs: number;
};

type CacheEntry<T> = {
  value: T;
  expiresAt: number;
};

type InflightRequest<T> = {
  promise: Promise<RequestState<T>>;
  controller: AbortController;
};

const defaultDelay = (ms: number) =>
  new Promise<void>((resolve) => setTimeout(resolve, ms));

export class ApiRequestManager<T> {
  private readonly client: ApiClient;
  private readonly retry: RetryConfig;
  private readonly cache: CacheConfig;
  private readonly inflight = new Map<string, InflightRequest<T>>();
  private readonly optimisticCache = new Map<string, CacheEntry<T>>();
  private readonly nowFn: () => number;

  constructor(
    client: ApiClient,
    retry: RetryConfig,
    cache: CacheConfig,
    nowFn: () => number = () => Date.now()
  ) {
    this.client = client;
    this.retry = retry;
    this.cache = cache;
    this.nowFn = nowFn;
  }

  async execute(
    key: string,
    onRetry?: (attempt: number, nextDelayMs: number) => void
  ): Promise<RequestState<T>> {
    const now = this.nowFn();
    const cached = this.optimisticCache.get(key);
    if (cached && cached.expiresAt > now) {
      return {
        data: cached.value,
        error: null,
        loading: false,
        retrying: false,
        fromCache: true,
        attempts: 0,
      };
    }

    const existing = this.inflight.get(key);
    if (existing) {
      return existing.promise;
    }

    const controller = new AbortController();
    const promise = this.runWithRetry(key, controller.signal, onRetry).finally(
      () => {
        this.inflight.delete(key);
      }
    );

    this.inflight.set(key, { promise, controller });
    return promise;
  }

  cancelAll() {
    this.inflight.forEach(({ controller }) => controller.abort());
    this.inflight.clear();
  }

  private async runWithRetry(
    key: string,
    signal: AbortSignal,
    onRetry?: (attempt: number, nextDelayMs: number) => void
  ): Promise<RequestState<T>> {
    let attempts = 0;
    let lastError: unknown;

    for (let attempt = 1; attempt <= this.retry.maxAttempts; attempt += 1) {
      attempts = attempt;
      try {
        const data = await this.client<T>(key, signal);
        this.optimisticCache.set(key, {
          value: data,
          expiresAt: this.nowFn() + this.cache.ttlMs,
        });
        return {
          data,
          error: null,
          loading: false,
          retrying: false,
          fromCache: false,
          attempts,
        };
      } catch (error) {
        if ((error as Error).name === "AbortError") {
          return {
            data: null,
            error: { message: "Request cancelled", code: "CANCELLED" },
            loading: false,
            retrying: false,
            fromCache: false,
            attempts,
          };
        }

        lastError = error;
        if (attempt < this.retry.maxAttempts) {
          const delay = this.retry.baseDelayMs * 2 ** (attempt - 1);
          onRetry?.(attempt, delay);
          await (this.retry.delayFn ?? defaultDelay)(delay);
        }
      }
    }

    return {
      data: null,
      error: toApiError(lastError),
      loading: false,
      retrying: false,
      fromCache: false,
      attempts,
    };
  }
}
