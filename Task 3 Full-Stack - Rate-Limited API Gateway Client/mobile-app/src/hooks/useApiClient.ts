import { useEffect, useMemo, useRef, useState } from "react";
import { ApiClient, RequestState } from "../api/types";
import { ApiRequestManager } from "./apiRequestManager";

type UseApiClientOptions = {
  cacheTtlMs?: number;
  maxAttempts?: number;
  baseDelayMs?: number;
};

const initialState: RequestState<unknown> = {
  data: null,
  error: null,
  loading: false,
  retrying: false,
  fromCache: false,
  attempts: 0,
};

export function useApiClient<T>(
  apiClient: ApiClient,
  options: UseApiClientOptions = {}
) {
  const [state, setState] = useState<RequestState<T>>(
    initialState as RequestState<T>
  );
  const requestIdRef = useRef(0);

  const manager = useMemo(
    () =>
      new ApiRequestManager<T>(
        apiClient,
        {
          maxAttempts: options.maxAttempts ?? 3,
          baseDelayMs: options.baseDelayMs ?? 200,
        },
        {
          ttlMs: options.cacheTtlMs ?? 5_000,
        }
      ),
    [apiClient, options.baseDelayMs, options.cacheTtlMs, options.maxAttempts]
  );

  const execute = async (key: string) => {
    requestIdRef.current += 1;
    const requestId = requestIdRef.current;

    setState((current) => ({
      ...current,
      loading: true,
      retrying: false,
      error: null,
    }));

    const result = await manager.execute(key, () => {
      if (requestIdRef.current !== requestId) {
        return;
      }
      setState((current) => ({
        ...current,
        loading: true,
        retrying: true,
      }));
    });

    if (requestIdRef.current !== requestId) {
      return result;
    }
    setState(result);
    return result;
  };

  useEffect(() => {
    return () => {
      manager.cancelAll();
    };
  }, [manager]);

  const cancel = () => {
    requestIdRef.current += 1;
    manager.cancelAll();
    setState((current) => ({
      ...current,
      loading: false,
      retrying: false,
      error: { message: "Aborted", code: "CANCELLED" },
    }));
  };

  return {
    state,
    execute,
    cancel,
  };
}
