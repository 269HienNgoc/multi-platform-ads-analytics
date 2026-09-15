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

export type WorkflowState =
  | "ACCOUNT_CONNECTED"
  | "ASSETS_SYNCED"
  | "PREFLIGHT_PASSED"
  | "SEED_PENDING"
  | "SEED_CREATING"
  | "SEED_RUNNING"
  | "SEED_COMPLETED"
  | "MAIN_PENDING"
  | "MAIN_VALIDATING"
  | "MAIN_CREATING"
  | "MAIN_RUNNING"
  | "PAUSED"
  | "NEEDS_MANUAL_REVIEW"
  | "FAILED";

export interface CampaignWorkflow {
  id: string;
  organization_id: string;
  ad_account_id: string;
  page_external_id: string;
  pixel_external_id?: string;
  pixel_event?: string;
  existing_post_id?: string;
  seed_campaign_external_id?: string;
  main_campaign_external_id?: string;
  seed_spend_limit_usd: number;
  state: WorkflowState;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateBulkWorkflowInput {
  organization_id: string;
  ad_account_ids: string[];
  page_external_id: string;
  pixel_external_id?: string;
  pixel_event?: string;
  existing_post_id?: string;
  seed_spend_limit_usd?: number;
}

export interface BulkWorkflowResponse {
  data: CampaignWorkflow[];
  count: number;
}
