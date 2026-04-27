export type ApiClient = <T>(
  key: string,
  signal?: AbortSignal
) => Promise<T>;

export interface ApiErrorShape {
  message: string;
  code?: string;
}

export interface RequestState<T> {
  data: T | null;
  error: ApiErrorShape | null;
  loading: boolean;
  retrying: boolean;
  fromCache: boolean;
  attempts: number;
}
