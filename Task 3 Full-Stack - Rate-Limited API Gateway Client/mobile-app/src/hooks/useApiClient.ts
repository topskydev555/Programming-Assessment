import { useCallback, useEffect, useMemo, useState } from "react";
import { ApiClient, ApiRequest, RequestState } from "../api/types";
import { ApiRequestManager, ManagerOptions } from "./apiRequestManager";

type UseApiClientOptions = ManagerOptions;

export type HookState<TData> = {
  data: TData | null;
  state: RequestState;
  error: string | null;
  lastSource: "none" | "cache" | "network";
};

export function useApiClient<TData>(client: ApiClient, options: UseApiClientOptions) {
  const manager = useMemo(() => new ApiRequestManager(client, options), [client, options]);
  const [result, setResult] = useState<HookState<TData>>({
    data: null,
    state: "idle",
    error: null,
    lastSource: "none",
  });

  useEffect(() => {
    return () => {
      manager.cancel();
    };
  }, [manager]);

  const execute = useCallback(
    async (request: ApiRequest) => {
      const optimistic = manager.getCached<TData>(request.key);
      if (optimistic != null) {
        setResult({
          data: optimistic,
          state: "loading",
          error: null,
          lastSource: "cache",
        });
      } else {
        setResult((prev) => ({ ...prev, state: "loading", error: null }));
      }

      try {
        const { data, fromCache } = await manager.run<TData>(request);
        setResult({
          data,
          state: "success",
          error: null,
          lastSource: fromCache ? "cache" : "network",
        });
      } catch (error) {
        const message = error instanceof Error ? error.message : "Unknown request error";
        const cancelled = message.includes("AbortError") || message.toLowerCase().includes("abort");
        setResult((prev) => ({
          ...prev,
          state: cancelled ? "cancelled" : "error",
          error: message,
        }));
      }
    },
    [manager]
  );

  const cancel = useCallback((requestKey?: string) => {
    manager.cancel(requestKey);
  }, [manager]);

  return {
    ...result,
    execute,
    cancel,
  };
}
