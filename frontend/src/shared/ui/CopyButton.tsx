"use client";

import { useEffect, useRef, useState } from "react";

interface CopyButtonProps {
  text: string;
  label?: string;
  className?: string;
}

export function CopyButton({
  text,
  label = "text",
  className = "",
}: CopyButtonProps) {
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  async function handleCopy() {
    if (!navigator?.clipboard?.writeText) {
      return;
    }
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
      timerRef.current = setTimeout(() => setCopied(false), 1800);
    } catch {
      setCopied(false);
    }
  }

  return (
    <button
      type="button"
      onClick={handleCopy}
      aria-label={copied ? `Copied ${label}` : `Copy ${label}`}
      className={`inline-flex items-center gap-1 rounded-lg border px-2 py-1 text-xs font-medium transition focus:outline-none focus:ring-2 focus:ring-indigo-300 ${
        copied
          ? "border-emerald-300 bg-emerald-50 text-emerald-700"
          : "border-slate-200 bg-slate-50 text-slate-600 hover:border-indigo-300 hover:bg-indigo-50 hover:text-indigo-700"
      } ${className}`}
    >
      <span aria-hidden="true">{copied ? "✓" : "⧉"}</span>
      <span>{copied ? "Copied!" : "Copy"}</span>
    </button>
  );
}
