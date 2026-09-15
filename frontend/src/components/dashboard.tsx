"use client";

import {
  Activity,
  AlertTriangle,
  ArrowDownRight,
  ArrowUpRight,
  Bell,
  Bot,
  Check,
  ChevronDown,
  ChevronRight,
  Command,
  FileText,
  Gauge,
  HelpCircle,
  Layers3,
  Menu,
  MoreHorizontal,
  Plus,
  Rocket,
  Search,
  Settings,
  Sparkles,
  Target,
  WalletCards,
  X,
  Zap,
} from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";

import { APIRequestError, createAdAccount, getBackendReadiness } from "@/lib/api";
import {
  initialAccounts,
  metrics,
  performanceSeries,
  tasks,
  workflow,
  type DashboardAccount,
} from "@/lib/dashboard-data";
import type { CreateAdAccountInput, Platform } from "@/types/ads";

const navigation = [
  { label: "Tổng quan", icon: Gauge, href: "#overview" },
  { label: "Tài khoản quảng cáo", icon: WalletCards, href: "#accounts", count: 24 },
  { label: "Chiến dịch", icon: Target, href: "#campaigns", count: 37 },
  { label: "Thư viện nội dung", icon: Layers3, href: "#creative" },
  { label: "Phân tích AI", icon: Bot, href: "#ai" },
];

const platformMeta: Record<Platform, { short: string; name: string }> = {
  meta: { short: "M", name: "Meta" },
  tiktok: { short: "T", name: "TikTok" },
  google: { short: "G", name: "Google" },
};

type BackendState = "checking" | "connected" | "offline";

function moneyPoint(value: number, index: number, maxValue: number) {
  const width = 720;
  const height = 210;
  const x = (index / (performanceSeries.length - 1)) * width;
  const y = height - (value / maxValue) * (height - 20) - 6;
  return `${x.toFixed(1)},${y.toFixed(1)}`;
}

function PerformanceChart() {
  const revenuePoints = performanceSeries.map((item, index) => moneyPoint(item.revenue, index, 60)).join(" ");
  const spendPoints = performanceSeries.map((item, index) => moneyPoint(item.spend, index, 60)).join(" ");

  return (
    <div className="chart-wrap" aria-label="Biểu đồ chi tiêu và doanh thu 14 ngày">
      <div className="chart-scale" aria-hidden="true">
        <span>₫60M</span>
        <span>₫40M</span>
        <span>₫20M</span>
        <span>₫0</span>
      </div>
      <div className="chart-stage">
        <svg viewBox="0 0 720 230" role="img" aria-labelledby="chart-title chart-description">
          <title id="chart-title">Hiệu suất quảng cáo trong 14 ngày</title>
          <desc id="chart-description">Doanh thu tăng từ 20,1 lên 55,3 triệu đồng; chi tiêu tăng từ 7,2 lên 16,1 triệu đồng.</desc>
          {[16, 76, 136, 196].map((y) => (
            <line key={y} x1="0" x2="720" y1={y} y2={y} className="chart-grid" />
          ))}
          <polyline points={revenuePoints} className="chart-line chart-line-revenue" />
          <polyline points={spendPoints} className="chart-line chart-line-spend" />
          {performanceSeries.map((item, index) => {
            const [cx, cy] = moneyPoint(item.revenue, index, 60).split(",");
            return <circle key={item.day} cx={cx} cy={cy} r={index === performanceSeries.length - 1 ? 5 : 2.5} className="chart-dot" />;
          })}
        </svg>
        <div className="chart-labels" aria-hidden="true">
          {performanceSeries.filter((_, index) => index % 2 === 0 || index === performanceSeries.length - 1).map((item) => (
            <span key={item.day}>{item.day}</span>
          ))}
        </div>
      </div>
    </div>
  );
}

function PlatformMark({ platform }: { platform: Platform }) {
  return (
    <span className={`platform-mark platform-${platform}`} aria-label={platformMeta[platform].name}>
      {platformMeta[platform].short}
    </span>
  );
}

function BackendIndicator({ state, onRetry }: { state: BackendState; onRetry: () => void }) {
  const label = state === "connected" ? "Backend đã kết nối" : state === "checking" ? "Đang kiểm tra backend" : "Backend chưa kết nối";
  return (
    <button className={`backend-state backend-${state}`} onClick={onRetry} type="button" title="Kiểm tra lại kết nối">
      <span className="state-dot" />
      {label}
    </button>
  );
}

function AddAccountDialog({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: (account: DashboardAccount) => void;
}) {
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState<{ type: "error" | "success"; text: string }>();

  useEffect(() => {
    if (!open) return;
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [open, onClose]);

  if (!open) return null;

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formElement = event.currentTarget;
    setSubmitting(true);
    setMessage(undefined);
    const form = new FormData(formElement);
    const input: CreateAdAccountInput = {
      platform: form.get("platform") as Platform,
      external_id: String(form.get("external_id")),
      name: String(form.get("name")),
      currency: String(form.get("currency")).toUpperCase(),
      timezone: String(form.get("timezone")),
      status: "active",
      provider_data: {},
    };

    try {
      const created = await createAdAccount(input);
      onCreated({
        id: created.id,
        name: created.name,
        platform: created.platform,
        externalID: created.external_id,
        status: "Đang chạy",
        campaigns: 0,
        spend: "₫0",
        roas: "—",
        change: "Mới",
        phase: "Camp mồi",
      });
      setMessage({ type: "success", text: `Đã thêm “${created.name}” vào backend.` });
      formElement.reset();
    } catch (error) {
      const text = error instanceof APIRequestError && error.status === 409
        ? "Tài khoản này đã tồn tại trên cùng nền tảng."
        : "Không thể thêm tài khoản. Kiểm tra backend và dữ liệu kết nối.";
      setMessage({ type: "error", text });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="dialog-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className="dialog" role="dialog" aria-modal="true" aria-labelledby="add-account-title">
        <div className="dialog-head">
          <div>
            <p className="dialog-kicker">Kết nối dữ liệu</p>
            <h2 id="add-account-title">Thêm tài khoản quảng cáo</h2>
          </div>
          <button className="icon-button" onClick={onClose} type="button" aria-label="Đóng">
            <X size={20} />
          </button>
        </div>
        <p className="dialog-intro">Tạo bản ghi chuẩn hóa trong backend. Access token sẽ được cấu hình ở bước connector riêng.</p>
        <form onSubmit={submit}>
          <div className="form-grid">
            <label>
              Nền tảng
              <select name="platform" defaultValue="meta">
                <option value="meta">Meta Ads</option>
                <option value="tiktok">TikTok Ads</option>
                <option value="google">Google Ads</option>
              </select>
            </label>
            <label>
              Mã tài khoản trên nền tảng
              <input name="external_id" placeholder="Ví dụ: act_123456789" required />
            </label>
            <label className="form-wide">
              Tên dễ nhận biết
              <input name="name" placeholder="Ví dụ: Bloom Skincare VN" required />
            </label>
            <label>
              Tiền tệ
              <input name="currency" defaultValue="VND" minLength={3} maxLength={3} required />
            </label>
            <label>
              Múi giờ
              <input name="timezone" defaultValue="Asia/Ho_Chi_Minh" required />
            </label>
          </div>
          {message && <p className={`form-message message-${message.type}`} role="status">{message.text}</p>}
          <div className="dialog-actions">
            <button className="button button-quiet" type="button" onClick={onClose}>Hủy</button>
            <button className="button button-primary" type="submit" disabled={submitting}>
              {submitting ? "Đang lưu…" : "Thêm tài khoản"}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}

function WorkflowDrawer({ open, onClose }: { open: boolean; onClose: () => void }) {
  if (!open) return null;
  return (
    <div className="drawer-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <aside className="drawer" role="dialog" aria-modal="true" aria-labelledby="workflow-title">
        <div className="dialog-head">
          <div>
            <p className="dialog-kicker">Thiết lập có hướng dẫn</p>
            <h2 id="workflow-title">Tạo luồng chiến dịch</h2>
          </div>
          <button className="icon-button" onClick={onClose} type="button" aria-label="Đóng"><X size={20} /></button>
        </div>
        <p className="dialog-intro">Chọn tài khoản rồi hoàn thành từng bước. Backend workflow hàng loạt chưa mở nên nút chạy thật đang được khóa.</p>
        <div className="drawer-account">
          <PlatformMark platform="meta" />
          <div><strong>Bloom Skincare VN</strong><span>Meta · act_8902…481</span></div>
          <ChevronDown size={18} />
        </div>
        <ol className="setup-steps">
          <li className="step-complete">
            <span className="step-number"><Check size={16} /></span>
            <div><strong>Kiểm tra tài khoản</strong><p>Page, pixel Purchase và phương thức thanh toán đã sẵn sàng.</p></div>
          </li>
          <li className="step-active">
            <span className="step-number">2</span>
            <div><strong>Cấu hình Camp mồi</strong><p>Chọn bài viết có sẵn, vị trí Việt Nam + 40 km và ngân sách $10.</p></div>
          </li>
          <li>
            <span className="step-number">3</span>
            <div><strong>Chuẩn bị Conversion</strong><p>Pixel, event, 2–3 ad set và 2–3 creative cho mỗi ad set.</p></div>
          </li>
        </ol>
        <div className="drawer-fields">
          <label>Bài viết dùng cho Camp mồi<select defaultValue="post-1"><option value="post-1">Video UGC — Giá dùng thử</option><option value="post-2">Carousel — Routine 3 bước</option></select></label>
          <div className="split-fields">
            <label>Ngân sách<input defaultValue="$10" /></label>
            <label>Bán kính<input defaultValue="40 km" /></label>
          </div>
        </div>
        <div className="drawer-footer">
          <p><AlertTriangle size={16} /> Cần bổ sung API workflow trước khi chạy thật.</p>
          <button className="button button-primary" disabled type="button">Lưu và chạy Camp mồi</button>
        </div>
      </aside>
    </div>
  );
}

export function Dashboard() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [addAccountOpen, setAddAccountOpen] = useState(false);
  const [workflowOpen, setWorkflowOpen] = useState(false);
  const [backendState, setBackendState] = useState<BackendState>("checking");
  const [period, setPeriod] = useState("14 ngày");
  const [platform, setPlatform] = useState<"all" | Platform>("all");
  const [query, setQuery] = useState("");
  const [accounts, setAccounts] = useState(initialAccounts);

  async function checkBackend() {
    setBackendState("checking");
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 4500);
    try {
      const result = await getBackendReadiness(controller.signal);
      setBackendState(result.status === "up" ? "connected" : "offline");
    } catch {
      setBackendState("offline");
    } finally {
      window.clearTimeout(timeout);
    }
  }

  useEffect(() => {
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 4500);
    void getBackendReadiness(controller.signal)
      .then((result) => setBackendState(result.status === "up" ? "connected" : "offline"))
      .catch(() => setBackendState("offline"))
      .finally(() => window.clearTimeout(timeout));

    return () => {
      controller.abort();
      window.clearTimeout(timeout);
    };
  }, []);

  const filteredAccounts = useMemo(() => accounts.filter((account) => {
    const matchesPlatform = platform === "all" || account.platform === platform;
    const normalized = query.trim().toLocaleLowerCase("vi");
    const matchesQuery = !normalized || `${account.name} ${account.externalID}`.toLocaleLowerCase("vi").includes(normalized);
    return matchesPlatform && matchesQuery;
  }), [accounts, platform, query]);

  return (
    <div className="app-shell">
      <aside className={`sidebar ${menuOpen ? "sidebar-open" : ""}`}>
        <a className="brand" href="#overview" aria-label="AdPilot — về tổng quan">
          <span className="brand-symbol"><Activity size={21} strokeWidth={2.5} /></span>
          <span><strong>AdPilot</strong><small>Campaign OS</small></span>
        </a>
        <nav className="primary-nav" aria-label="Điều hướng chính">
          <p>Không gian làm việc</p>
          {navigation.map((item, index) => {
            const Icon = item.icon;
            return (
              <a className={index === 0 ? "active" : ""} href={item.href} key={item.label} onClick={() => setMenuOpen(false)}>
                <Icon size={19} />
                <span>{item.label}</span>
                {item.count && <b>{item.count}</b>}
              </a>
            );
          })}
        </nav>
        <div className="sidebar-lab">
          <span><Sparkles size={17} /> Gợi ý từ AI</span>
          <p>2 chiến dịch có thể giảm CPA nếu chuyển 15% ngân sách.</p>
          <a href="#ai">Xem phân tích <ChevronRight size={15} /></a>
        </div>
        <nav className="secondary-nav" aria-label="Trợ giúp và cài đặt">
          <a href="#reports"><FileText size={18} /> Báo cáo</a>
          <a href="#settings"><Settings size={18} /> Cài đặt</a>
          <a href="#help"><HelpCircle size={18} /> Trợ giúp</a>
        </nav>
        <div className="profile">
          <span className="avatar">HN</span>
          <span><strong>Hiển Ngọc</strong><small>Workspace owner</small></span>
          <MoreHorizontal size={18} />
        </div>
      </aside>

      <main>
        <header className="topbar">
          <button className="mobile-menu icon-button" type="button" onClick={() => setMenuOpen((value) => !value)} aria-label="Mở menu"><Menu size={21} /></button>
          <BackendIndicator state={backendState} onRetry={() => void checkBackend()} />
          <div className="top-actions">
            <button className="command-search" type="button"><Search size={17} /><span>Tìm nhanh</span><kbd><Command size={12} /> K</kbd></button>
            <div className="notification-wrap">
              <button className="icon-button notification-button" type="button" onClick={() => setNotificationsOpen((value) => !value)} aria-expanded={notificationsOpen} aria-label="Thông báo">
                <Bell size={19} /><span />
              </button>
              {notificationsOpen && (
                <div className="notification-popover">
                  <strong>Cần bạn xử lý</strong>
                  <p>ROAS của Luna Home đã giảm dưới ngưỡng 2,0.</p>
                  <button type="button" onClick={() => { setNotificationsOpen(false); document.querySelector("#attention")?.scrollIntoView({ behavior: "smooth" }); }}>Xem cảnh báo</button>
                </div>
              )}
            </div>
            <button className="button button-quiet top-quiet" type="button" onClick={() => setAddAccountOpen(true)}><Plus size={17} /> Thêm tài khoản</button>
            <button className="button button-primary" type="button" onClick={() => setWorkflowOpen(true)}><Rocket size={17} /> Tạo chiến dịch</button>
          </div>
        </header>

        <div className="page" id="overview">
          <section className="page-heading">
            <div>
              <p className="today">Thứ Ba, 15 tháng 9</p>
              <h1>Chào buổi sáng, Hiển.</h1>
              <p>Bạn đang có <strong>17 chiến dịch hoạt động</strong> trên 24 tài khoản.</p>
            </div>
            <div className="period-control" aria-label="Khoảng thời gian">
              {["7 ngày", "14 ngày", "30 ngày"].map((item) => (
                <button key={item} className={period === item ? "active" : ""} type="button" onClick={() => setPeriod(item)}>{item}</button>
              ))}
              <button className="date-button" type="button">02/09 – 15/09 <ChevronDown size={15} /></button>
            </div>
          </section>

          <section className="data-note" aria-label="Trạng thái dữ liệu">
            <span>Dữ liệu mô phỏng</span>
            <p>Backend chưa có endpoint đọc metrics và danh sách tài khoản. Kết nối thật đã sẵn sàng cho Health và Thêm tài khoản.</p>
          </section>

          <section className="metric-ribbon" aria-label="Chỉ số tổng quan">
            {metrics.map((metric) => (
              <article key={metric.label}>
                <span>{metric.label}</span>
                <strong>{metric.value}</strong>
                <small className={`metric-${metric.tone}`}><ArrowUpRight size={13} /> {metric.change}</small>
              </article>
            ))}
          </section>

          <div className="dashboard-grid">
            <section className="panel performance-panel">
              <div className="panel-head">
                <div><h2>Nhịp hiệu suất</h2><p>So sánh dòng tiền quảng cáo trong {period.toLowerCase()} qua.</p></div>
                <div className="chart-legend"><span className="legend-revenue">Doanh thu</span><span className="legend-spend">Chi tiêu</span></div>
              </div>
              <div className="performance-summary">
                <div><span>Doanh thu hôm nay</span><strong>₫55,3M</strong><small><ArrowUpRight size={13} /> 7,2% so với hôm qua</small></div>
                <div><span>Chi tiêu hôm nay</span><strong>₫16,1M</strong><small>29,1% doanh thu</small></div>
              </div>
              <PerformanceChart />
            </section>

            <section className="panel attention-panel" id="attention">
              <div className="panel-head"><div><h2>Cần xử lý</h2><p>Ưu tiên theo tác động ước tính.</p></div><span className="task-count">3</span></div>
              <div className="task-list">
                {tasks.map((task) => (
                  <article key={task.title} className={`task task-${task.severity}`}>
                    <span className="task-signal">{task.severity === "ready" ? <Check size={16} /> : <AlertTriangle size={16} />}</span>
                    <div><strong>{task.title}</strong><p>{task.detail}</p><button type="button">{task.action} <ChevronRight size={14} /></button></div>
                  </article>
                ))}
              </div>
              <button className="text-button" type="button">Xem toàn bộ đề xuất <ChevronRight size={15} /></button>
            </section>
          </div>

          <section className="panel account-panel" id="accounts">
            <div className="panel-head account-head">
              <div><h2>Tài khoản quảng cáo</h2><p>Theo dõi chi tiêu và giai đoạn vận hành.</p></div>
              <div className="account-tools">
                <div className="platform-filter">
                  {(["all", "meta", "tiktok", "google"] as const).map((item) => (
                    <button key={item} className={platform === item ? "active" : ""} onClick={() => setPlatform(item)} type="button">
                      {item === "all" ? "Tất cả" : platformMeta[item].name}
                    </button>
                  ))}
                </div>
                <label className="table-search"><Search size={16} /><span className="sr-only">Tìm tài khoản</span><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Tìm tài khoản" /></label>
              </div>
            </div>
            <div className="table-wrap">
              <table>
                <thead><tr><th>Tài khoản</th><th>Trạng thái</th><th>Chiến dịch</th><th>Chi tiêu</th><th>ROAS</th><th>Giai đoạn</th><th><span className="sr-only">Thao tác</span></th></tr></thead>
                <tbody>
                  {filteredAccounts.map((account) => (
                    <tr key={account.id}>
                      <td><div className="account-name"><PlatformMark platform={account.platform} /><span><strong>{account.name}</strong><small>{platformMeta[account.platform].name} · {account.externalID}</small></span></div></td>
                      <td><span className={`status status-${account.status === "Đang chạy" ? "active" : account.status === "Cần xử lý" ? "warning" : "paused"}`}><i />{account.status}</span></td>
                      <td>{account.campaigns}</td>
                      <td><strong>{account.spend}</strong></td>
                      <td><span className="roas-value">{account.roas}</span><small className={account.change.startsWith("−") ? "down" : "up"}>{account.change.startsWith("−") ? <ArrowDownRight size={13} /> : <ArrowUpRight size={13} />}{account.change}</small></td>
                      <td><span className="phase">{account.phase}</span></td>
                      <td><button className="icon-button table-menu" type="button" aria-label={`Mở thao tác cho ${account.name}`}><MoreHorizontal size={18} /></button></td>
                    </tr>
                  ))}
                  {filteredAccounts.length === 0 && <tr><td colSpan={7} className="empty-row">Không tìm thấy tài khoản phù hợp.</td></tr>}
                </tbody>
              </table>
            </div>
          </section>

          <section className="workflow-panel" id="campaigns">
            <div className="workflow-copy">
              <span className="workflow-icon"><Zap size={19} /></span>
              <h2>Luồng thiết lập chiến dịch</h2>
              <p>Mỗi tài khoản đi qua cùng một quy trình kiểm soát, từ nhận tài khoản đến mở rộng ngân sách.</p>
              <button className="button button-ink" type="button" onClick={() => setWorkflowOpen(true)}>Mở trình thiết lập <ChevronRight size={16} /></button>
            </div>
            <div className="workflow-track">
              {workflow.map((step, index) => (
                <article key={step.name} className={`workflow-step workflow-${step.state}`}>
                  <div className="step-top"><span>{step.state === "done" ? <Check size={16} /> : index + 1}</span><b>{step.accounts} tài khoản</b></div>
                  <h3>{step.name}</h3><p>{step.detail}</p>
                </article>
              ))}
            </div>
          </section>
        </div>
      </main>

      {menuOpen && <button className="mobile-overlay" type="button" onClick={() => setMenuOpen(false)} aria-label="Đóng menu" />}
      <AddAccountDialog open={addAccountOpen} onClose={() => setAddAccountOpen(false)} onCreated={(account) => setAccounts((current) => [account, ...current])} />
      <WorkflowDrawer open={workflowOpen} onClose={() => setWorkflowOpen(false)} />
    </div>
  );
}
