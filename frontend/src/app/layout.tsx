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
      <body className="flex min-h-screen flex-col antialiased">
        <div className="flex-1">{children}</div>
        <footer className="mx-auto max-w-3xl px-4 pb-8 pt-4 text-center text-xs text-slate-500">
          Dictionary data from CC-Canto, CC-CEDICT, HKCanCor, wordfreq and Open English WordNet,
          shared under{" "}
          <a
            href="https://creativecommons.org/licenses/by-sa/4.0/"
            className="underline hover:text-indigo-600"
          >
            CC BY-SA 4.0
          </a>
          .{" "}
          <a
            href="https://github.com/kaho-dev/saycanto/blob/main/data/CREDITS.md"
            className="underline hover:text-indigo-600"
          >
            Full credits
          </a>{" "}
          ·{" "}
          <a
            href="https://github.com/kaho-dev/saycanto"
            className="underline hover:text-indigo-600"
          >
            Source on GitHub
          </a>
        </footer>
      </body>
    </html>
  );
}
