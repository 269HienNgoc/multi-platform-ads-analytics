export type Platform = "meta" | "tiktok" | "google";
export type EntityStatus = "active" | "paused" | "archived";

export interface AdAccount {
  id: string;
  platform: Platform;
  external_id: string;
  name: string;
  currency: string;
  timezone: string;
  status: EntityStatus;
  provider_data: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface CreateAdAccountInput {
  platform: Platform;
  external_id: string;
  name: string;
  currency: string;
  timezone: string;
  status: EntityStatus;
  provider_data?: Record<string, unknown>;
}

export interface HealthResponse {
  status: "up" | "down";
  checks: Record<string, "up" | "down">;
}

export interface APIErrorResponse {
  error: {
    code: string;
    message: string;
  };
}
