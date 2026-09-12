"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createBulkWorkflows, getWorkflows, Workflow } from "@/lib/api";

const stateStyle: Record<string, string> = {
  SEED_PENDING: "bg-amber-500/10 text-amber-300 ring-amber-500/20",
  SEED_RUNNING: "bg-blue-500/10 text-blue-300 ring-blue-500/20",
  MAIN_PENDING: "bg-violet-500/10 text-violet-300 ring-violet-500/20",
  MAIN_RUNNING: "bg-emerald-500/10 text-emerald-300 ring-emerald-500/20",
  PAUSED: "bg-zinc-500/10 text-zinc-300 ring-zinc-500/20",
  FAILED: "bg-red-500/10 text-red-300 ring-red-500/20",
  NEEDS_MANUAL_REVIEW: "bg-orange-500/10 text-orange-300 ring-orange-500/20",
};

export function Dashboard() {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const [organizationId, setOrganizationId] = useState("org-demo");
  const [pageId, setPageId] = useState("page-demo");
  const [pixelId, setPixelId] = useState("pixel-demo");
  const [pixelEvent, setPixelEvent] = useState("CompleteRegistration");
  const [existingPostId, setExistingPostId] = useState("post-demo");
  const [seedLimit, setSeedLimit] = useState(10);
  const [accounts, setAccounts] = useState("act_001\nact_002\nact_003");

  const refresh = useCallback(async () => {
    try {
      setError("");
      setWorkflows(await getWorkflows());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Cannot load workflows");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const stats = useMemo(() => {
    const running = workflows.filter((item) => item.state.endsWith("RUNNING")).length;
    const review = workflows.filter((item) => item.state === "NEEDS_MANUAL_REVIEW" || item.state === "FAILED").length;
    return { total: workflows.length, running, review };
  }, [workflows]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    const adAccountIds = accounts
      .split(/[\n,]/)
      .map((value) => value.trim())
      .filter(Boolean);

    if (!adAccountIds.length) {
      setError("Nhập ít nhất một ad account.");
      return;
    }

    try {
      setSubmitting(true);
      setError("");
      await createBulkWorkflows({
        organizationId,
        adAccountIds,
        pageId,
        pixelId,
        pixelEvent,
        existingPostId,
        seedSpendLimitUsd: seedLimit,
      });
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="min-h-screen px-5 py-8 md:px-10 lg:px-16">
      <div className="mx-auto max-w-7xl space-y-8">
        <header className="flex flex-col gap-3 border-b border-zinc-800 pb-7 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="text-sm font-medium uppercase tracking-[0.2em] text-zinc-500">Multi-platform ads</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">Campaign Automation Control Center</h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-zinc-400">
              Bulk-create independent workflows for multiple advertising accounts and monitor each account state separately.
            </p>
          </div>
          <button onClick={() => void refresh()} className="rounded-lg border border-zinc-700 px-4 py-2 text-sm hover:bg-zinc-900">
            Refresh
          </button>
        </header>

        <section className="grid gap-4 sm:grid-cols-3">
          <Stat label="Total workflows" value={stats.total} />
          <Stat label="Running" value={stats.running} />
          <Stat label="Need attention" value={stats.review} />
        </section>

        <div className="grid gap-6 xl:grid-cols-[380px_1fr]">
          <form onSubmit={submit} className="space-y-4 rounded-xl border border-zinc-800 bg-zinc-950 p-5">
            <div>
              <h2 className="font-semibold">Bulk campaign workflow</h2>
              <p className="mt-1 text-xs leading-5 text-zinc-500">One request can create workflows for many ad accounts; each keeps its own state and errors.</p>
            </div>
            <Field label="Organization" value={organizationId} onChange={setOrganizationId} />
            <Field label="Page" value={pageId} onChange={setPageId} />
            <Field label="Pixel" value={pixelId} onChange={setPixelId} />
            <Field label="Pixel event" value={pixelEvent} onChange={setPixelEvent} />
            <Field label="Existing post" value={existingPostId} onChange={setExistingPostId} />
            <label className="block text-sm">
              <span className="mb-1.5 block text-zinc-400">Seed spend limit (USD)</span>
              <input type="number" min="1" step="0.5" value={seedLimit} onChange={(e) => setSeedLimit(Number(e.target.value))} className="w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 outline-none focus:border-zinc-600" />
            </label>
            <label className="block text-sm">
              <span className="mb-1.5 block text-zinc-400">Ad accounts — one per line</span>
              <textarea value={accounts} onChange={(e) => setAccounts(e.target.value)} rows={5} className="w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 font-mono text-xs outline-none focus:border-zinc-600" />
            </label>
            <button disabled={submitting} className="w-full rounded-lg bg-white px-4 py-2.5 text-sm font-semibold text-black disabled:opacity-50">
              {submitting ? "Creating…" : "Create workflows"}
            </button>
            {error ? <p className="rounded-lg bg-red-500/10 p-3 text-xs text-red-300">{error}</p> : null}
          </form>

          <section className="overflow-hidden rounded-xl border border-zinc-800 bg-zinc-950">
            <div className="border-b border-zinc-800 px-5 py-4">
              <h2 className="font-semibold">Account workflows</h2>
              <p className="mt-1 text-xs text-zinc-500">Seed and main campaign states are intentionally independent per ad account.</p>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[760px] text-left text-sm">
                <thead className="border-b border-zinc-800 text-xs uppercase tracking-wide text-zinc-500">
                  <tr>
                    <th className="px-5 py-3">Ad account</th>
                    <th className="px-5 py-3">Page</th>
                    <th className="px-5 py-3">State</th>
                    <th className="px-5 py-3">Pixel event</th>
                    <th className="px-5 py-3 text-right">Seed limit</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-900">
                  {loading ? (
                    <tr><td colSpan={5} className="px-5 py-8 text-center text-zinc-500">Loading…</td></tr>
                  ) : workflows.length === 0 ? (
                    <tr><td colSpan={5} className="px-5 py-8 text-center text-zinc-500">No workflows yet. Create a bulk workflow on the left.</td></tr>
                  ) : workflows.map((item) => (
                    <tr key={item.id} className="hover:bg-zinc-900/50">
                      <td className="px-5 py-4 font-mono text-xs">{item.adAccountId}</td>
                      <td className="px-5 py-4 font-mono text-xs text-zinc-400">{item.pageId}</td>
                      <td className="px-5 py-4"><StateBadge state={item.state} /></td>
                      <td className="px-5 py-4 text-zinc-400">{item.pixelEvent || "—"}</td>
                      <td className="px-5 py-4 text-right">${item.seedSpendLimitUsd.toFixed(2)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        </div>
      </div>
    </main>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
      <div className="text-xs uppercase tracking-wide text-zinc-500">{label}</div>
      <div className="mt-2 text-3xl font-semibold">{value}</div>
    </div>
  );
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="block text-sm">
      <span className="mb-1.5 block text-zinc-400">{label}</span>
      <input value={value} onChange={(e) => onChange(e.target.value)} className="w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 outline-none focus:border-zinc-600" />
    </label>
  );
}

function StateBadge({ state }: { state: string }) {
  const style = stateStyle[state] ?? "bg-zinc-500/10 text-zinc-300 ring-zinc-500/20";
  return <span className={`rounded-full px-2.5 py-1 text-[11px] font-medium ring-1 ring-inset ${style}`}>{state}</span>;
}
