export type Workflow = {
  id: string;
  organizationId: string;
  adAccountId: string;
  pageId: string;
  pixelId?: string;
  pixelEvent?: string;
  existingPostId?: string;
  seedSpendLimitUsd: number;
  state: string;
  lastError?: string;
  createdAt: string;
  updatedAt: string;
};

export type BulkWorkflowInput = {
  organizationId: string;
  adAccountIds: string[];
  pageId: string;
  pixelId?: string;
  pixelEvent?: string;
  existingPostId?: string;
  seedSpendLimitUsd: number;
};

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function getWorkflows(): Promise<Workflow[]> {
  const response = await fetch(`${API_URL}/api/v1/workflows`, { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`Cannot load workflows (${response.status})`);
  }
  const body = (await response.json()) as { data: Workflow[] };
  return body.data;
}

export async function createBulkWorkflows(input: BulkWorkflowInput): Promise<Workflow[]> {
  const response = await fetch(`${API_URL}/api/v1/workflows/bulk`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `Cannot create workflows (${response.status})`);
  }
  const body = (await response.json()) as { data: Workflow[] };
  return body.data;
}
