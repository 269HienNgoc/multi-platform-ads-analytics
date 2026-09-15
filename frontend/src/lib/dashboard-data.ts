import type { Platform } from "@/types/ads";

export interface DashboardAccount {
  id: string;
  name: string;
  platform: Platform;
  externalID: string;
  status: "Đang chạy" | "Cần xử lý" | "Tạm dừng";
  campaigns: number;
  spend: string;
  roas: string;
  change: string;
  phase: "Camp mồi" | "Conversion" | "Scale";
}

export const metrics = [
  { label: "Chi tiêu", value: "₫184,6M", change: "+12,4%", tone: "neutral" },
  { label: "Doanh thu", value: "₫612,8M", change: "+18,1%", tone: "good" },
  { label: "ROAS", value: "3,32", change: "+0,16", tone: "good" },
  { label: "Đơn hàng", value: "1.284", change: "+9,7%", tone: "good" },
  { label: "CPA", value: "₫143,8K", change: "−6,2%", tone: "good" },
];

export const performanceSeries = [
  { day: "02/09", spend: 7.2, revenue: 20.1 },
  { day: "03/09", spend: 8.1, revenue: 24.4 },
  { day: "04/09", spend: 7.8, revenue: 22.9 },
  { day: "05/09", spend: 9.4, revenue: 29.8 },
  { day: "06/09", spend: 10.2, revenue: 31.6 },
  { day: "07/09", spend: 9.8, revenue: 34.3 },
  { day: "08/09", spend: 11.1, revenue: 38.5 },
  { day: "09/09", spend: 10.6, revenue: 35.7 },
  { day: "10/09", spend: 12.4, revenue: 41.9 },
  { day: "11/09", spend: 13.2, revenue: 44.1 },
  { day: "12/09", spend: 12.8, revenue: 42.6 },
  { day: "13/09", spend: 14.1, revenue: 48.8 },
  { day: "14/09", spend: 15.4, revenue: 51.6 },
  { day: "15/09", spend: 16.1, revenue: 55.3 },
];

export const initialAccounts: DashboardAccount[] = [
  {
    id: "demo-1",
    name: "Bloom Skincare VN",
    platform: "meta",
    externalID: "act_8902…481",
    status: "Đang chạy",
    campaigns: 8,
    spend: "₫46,8M",
    roas: "4,12",
    change: "+18%",
    phase: "Scale",
  },
  {
    id: "demo-2",
    name: "Luna Home Official",
    platform: "tiktok",
    externalID: "72814…029",
    status: "Cần xử lý",
    campaigns: 5,
    spend: "₫31,2M",
    roas: "1,84",
    change: "−22%",
    phase: "Conversion",
  },
  {
    id: "demo-3",
    name: "Nội thất Mộc",
    platform: "google",
    externalID: "432-901-7762",
    status: "Đang chạy",
    campaigns: 4,
    spend: "₫28,6M",
    roas: "3,48",
    change: "+7%",
    phase: "Conversion",
  },
  {
    id: "demo-4",
    name: "Sách nhỏ mỗi ngày",
    platform: "meta",
    externalID: "act_1740…205",
    status: "Tạm dừng",
    campaigns: 3,
    spend: "₫12,4M",
    roas: "2,06",
    change: "0%",
    phase: "Camp mồi",
  },
];

export const tasks = [
  {
    title: "ROAS giảm 22%",
    detail: "Luna Home · Conversion 09/2026",
    action: "Xem chiến dịch",
    severity: "high",
  },
  {
    title: "3 mẫu quảng cáo sắp mỏi",
    detail: "Tần suất > 3,2 trong 3 ngày",
    action: "Đổi creative",
    severity: "medium",
  },
  {
    title: "Camp mồi đủ ngân sách",
    detail: "Sách nhỏ mỗi ngày · đã tiêu $10",
    action: "Tạo conversion",
    severity: "ready",
  },
];

export const workflow = [
  {
    name: "Nhận tài khoản",
    detail: "Xác minh Page, pixel, quyền và phương thức thanh toán",
    accounts: 3,
    state: "done",
  },
  {
    name: "Chạy Camp mồi",
    detail: "Tương tác bài viết · VN 40 km · dừng khi đạt khoảng $10",
    accounts: 2,
    state: "running",
  },
  {
    name: "Conversion & Scale",
    detail: "2–3 ad set · test creative · tăng 20–30% mỗi ngày",
    accounts: 12,
    state: "next",
  },
];
