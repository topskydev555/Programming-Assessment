import { ApiErrorShape } from "./types";

export function toApiError(error: unknown): ApiErrorShape {
  if (error instanceof Error) {
    return { message: error.message };
  }
  return { message: "Unknown error" };
}
export function createAbortError(): Error {
  const error = new Error("Request aborted");
  error.name = "AbortError";
  return error;
}

export function isAbortError(error: unknown): boolean {
  if (!(error instanceof Error)) {
    return false;
  }
  return error.name === "AbortError" || error.message.toLowerCase().includes("abort");
}
