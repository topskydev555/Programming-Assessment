import { describe, expect, it } from "vitest";
import { ApiClient, ApiRequest } from "../api/types";
import { ApiRequestManager } from "./apiRequestManager";

class StubApiClient implements ApiClient {
  public calls = 0;
  private failures = 0;

  setFailures(count: number) {
    this.failures = count;
  }

  async request<TData>(request: ApiRequest, signal?: AbortSignal): Promise<TData> {
    this.calls++;
    if (signal?.aborted) {
      throw new DOMException("Aborted", "AbortError");
    }
    if (this.failures > 0) {
      this.failures--;
      throw new Error("Transient");
    }
    return { key: request.key, ok: true } as TData;
  }
}

describe("ApiRequestManager", () => {
  it("deduplicates concurrent requests by key", async () => {
    const client = new StubApiClient();
    const manager = new ApiRequestManager(client, {
      ttlMs: 1_000,
      retry: { maxAttempts: 1, baseDelayMs: 1 },
    });

    const req = { key: "a", url: "/a", method: "GET" as const };
    const [r1, r2] = await Promise.all([manager.run(req), manager.run(req)]);

    expect(r1.data).toEqual(r2.data);
    expect(client.calls).toBe(1);
  });

  it("retries transient failures", async () => {
    const client = new StubApiClient();
    client.setFailures(2);
    const manager = new ApiRequestManager(client, {
      ttlMs: 1_000,
      retry: { maxAttempts: 3, baseDelayMs: 1 },
    });

    const req = { key: "b", url: "/b", method: "GET" as const };
    const result = await manager.run(req);

    expect(result.data).toEqual({ key: "b", ok: true });
    expect(client.calls).toBe(3);
  });

  it("serves cached response before ttl expiration", async () => {
    const client = new StubApiClient();
    let now = 1000;
    const manager = new ApiRequestManager(client, {
      ttlMs: 50,
      retry: { maxAttempts: 1, baseDelayMs: 1 },
      now: () => now,
    });

    const req = { key: "c", url: "/c", method: "GET" as const };
    await manager.run(req);
    now += 10;
    const second = await manager.run(req);

    expect(second.fromCache).toBe(true);
    expect(client.calls).toBe(1);
  });
});
