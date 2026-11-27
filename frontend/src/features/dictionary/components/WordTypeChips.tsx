"use client";

import type { WordType } from "../types";
import { WORD_TYPE_LABELS, WORD_TYPES } from "../wordTypes";

interface WordTypeChipsProps {
  /** Matches per word type for the current search (ignoring the type filter). */
  counts: Partial<Record<WordType, number>>;
  /** Matches with no known word type. */
  untyped: number;
  active: WordType | "";
  onChange: (type: WordType | "") => void;
}

/**
 * Filter chips for word types. Only types with matches are shown, each with its count.
 * Many entries have no word type yet, so when a filter is active the chips say how many
 * matches it hides.
 */
export function WordTypeChips({
  counts,
  untyped,
  active,
  onChange,
}: WordTypeChipsProps) {
  const available = WORD_TYPES.filter((t) => (counts[t] ?? 0) > 0 || t === active);
  if (available.length === 0) {
    return null;
  }

  const chip = (selected: boolean) =>
    `rounded-full border px-3 py-1 text-xs font-medium transition focus:outline-none focus:ring-2 focus:ring-indigo-300 ${
      selected
        ? "border-indigo-500 bg-indigo-600 text-white"
        : "border-slate-200 bg-white text-slate-700 hover:border-indigo-300 hover:bg-indigo-50/60 hover:text-indigo-700"
    }`;

  return (
    <div className="mt-3" role="group" aria-label="Filter by word type">
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-xs font-medium text-slate-500">Type:</span>
        <button
          type="button"
          aria-pressed={active === ""}
          onClick={() => onChange("")}
          className={chip(active === "")}
        >
          All
        </button>
        {available.map((t) => (
          <button
            key={t}
            type="button"
            aria-pressed={active === t}
            onClick={() => onChange(active === t ? "" : t)}
            className={chip(active === t)}
          >
            {WORD_TYPE_LABELS[t]}{" "}
            <span className={active === t ? "text-indigo-100" : "text-slate-400"}>
              {counts[t] ?? 0}
            </span>
          </button>
        ))}
      </div>
      {active !== "" && untyped > 0 && (
        <p className="mt-2 text-xs text-slate-500">
          {untyped} {untyped === 1 ? "match has" : "matches have"} no word type yet
          and {untyped === 1 ? "is" : "are"} hidden by this filter.
        </p>
      )}
    </div>
  );
}
