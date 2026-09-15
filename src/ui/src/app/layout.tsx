import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "NEXUS // Digital Twin",
  description: "Industrial digital twin operations console",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
