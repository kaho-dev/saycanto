import type { RegisterType } from "../types";

const STYLES: Record<string, string> = {
  spoken: "bg-emerald-100 text-emerald-800 border-emerald-200",
  formal: "bg-amber-100 text-amber-800 border-amber-200",
  both: "bg-sky-100 text-sky-800 border-sky-200",
};

const LABELS: Record<string, string> = {
  spoken: "Spoken",
  formal: "Formal",
  both: "Spoken & Formal",
};

export function RegisterBadge({ register }: { register?: RegisterType }) {
  if (!register || !LABELS[register]) {
    return null;
  }
  return (
    <span
      className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${STYLES[register]}`}
    >
      {LABELS[register]}
    </span>
  );
}
