export type ApiMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export type ApiRequest<TBody = unknown> = {
  key: string;
  url: string;
  method?: ApiMethod;
  headers?: Record<string, string>;
  body?: TBody;
};

export interface ApiClient {
  request<TData>(request: ApiRequest, signal?: AbortSignal): Promise<TData>;
}

export type CacheEntry<TData> = {
  value: TData;
  expiresAt: number;
};

export type RequestState = "idle" | "loading" | "success" | "error" | "cancelled";
