import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Ads Automation",
  description: "Multi-account advertising campaign automation dashboard",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="vi">
      <body>{children}</body>
    </html>
  );
}
