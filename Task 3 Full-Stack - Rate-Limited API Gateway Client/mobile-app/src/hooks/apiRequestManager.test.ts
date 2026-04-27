import { describe, expect, it, vi } from "vitest";
import { ApiRequestManager } from "./apiRequestManager";
import { ApiClient } from "../api/types";

type Payload = { value: string };

describe("ApiRequestManager", () => {
  it("deduplicates concurrent requests", async () => {
    let callCount = 0;
    const client: ApiClient = async () => {
      callCount += 1;
      await Promise.resolve();
      return { value: "ok" };
    };

    const manager = new ApiRequestManager<Payload>(
      client,
      { maxAttempts: 2, baseDelayMs: 10, delayFn: vi.fn() },
      { ttlMs: 1_000 },
      () => 1000
    );

    const [a, b] = await Promise.all([
      manager.execute("k1"),
      manager.execute("k1"),
    ]);

    expect(callCount).toBe(1);
    expect(a.data?.value).toBe("ok");
    expect(b.data?.value).toBe("ok");
  });

  it("returns optimistic cache when ttl is valid", async () => {
    let now = 1000;
    const clientMock = vi.fn(async () => ({ value: "cached" }));
    const client: ApiClient = clientMock;
    const manager = new ApiRequestManager<Payload>(
      client,
      { maxAttempts: 2, baseDelayMs: 10, delayFn: vi.fn() },
      { ttlMs: 500 },
      () => now
    );

    const first = await manager.execute("k2");
    now = 1200;
    const second = await manager.execute("k2");

    expect(first.fromCache).toBe(false);
    expect(second.fromCache).toBe(true);
    expect(clientMock).toHaveBeenCalledTimes(1);
  });

  it("retries with exponential delay before failing", async () => {
    const delayFn = vi.fn(async () => {});
    const client: ApiClient = vi.fn(async () => {
      throw new Error("boom");
    });

    const manager = new ApiRequestManager<Payload>(
      client,
      { maxAttempts: 3, baseDelayMs: 100, delayFn },
      { ttlMs: 500 },
      () => 1000
    );

    const result = await manager.execute("k3");

    expect(result.error?.message).toBe("boom");
    expect(result.attempts).toBe(3);
    expect(delayFn).toHaveBeenCalledTimes(2);
    expect(delayFn).toHaveBeenNthCalledWith(1, 100);
    expect(delayFn).toHaveBeenNthCalledWith(2, 200);
  });

  it("cancels inflight requests", async () => {
    const client: ApiClient = async (_key, signal) => {
      return new Promise<Payload>((resolve, reject) => {
        signal?.addEventListener("abort", () => {
          const err = new Error("aborted");
          err.name = "AbortError";
          reject(err);
        });
        setTimeout(() => resolve({ value: "late" }), 1000);
      });
    };

    const manager = new ApiRequestManager<Payload>(
      client,
      { maxAttempts: 1, baseDelayMs: 10, delayFn: vi.fn() },
      { ttlMs: 500 },
      () => 1000
    );

    const promise = manager.execute("k4");
    manager.cancelAll();
    const result = await promise;

    expect(result.error?.code).toBe("CANCELLED");
  });
});
