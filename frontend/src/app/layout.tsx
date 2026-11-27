import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "SayCanto — English to Cantonese Dictionary",
  description:
    "Open-source English to Cantonese dictionary with Jyutping front-and-center, spoken vs. formal register tags, and natural examples.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="min-h-screen antialiased">{children}</body>
    </html>
  );
}
