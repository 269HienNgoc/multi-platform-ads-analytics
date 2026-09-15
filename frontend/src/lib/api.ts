import type { AdAccount, APIErrorResponse, CreateAdAccountInput, HealthResponse } from "@/types/ads";

export class APIRequestError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
  ) {
    super(message);
    this.name = "APIRequestError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/backend${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      ...(init?.body ? { "Content-Type": "application/json" } : {}),
      ...init?.headers,
    },
  });

  if (!response.ok) {
    let details: APIErrorResponse | undefined;
    try {
      details = (await response.json()) as APIErrorResponse;
    } catch {
      details = undefined;
    }
    throw new APIRequestError(
      details?.error.message ?? `Yêu cầu thất bại (${response.status})`,
      response.status,
      details?.error.code,
    );
  }

  return (await response.json()) as T;
}

export function getBackendReadiness(signal?: AbortSignal) {
  return request<HealthResponse>("/health/ready", { signal });
}

export function createAdAccount(input: CreateAdAccountInput) {
  return request<AdAccount>("/api/v1/ad-accounts", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
