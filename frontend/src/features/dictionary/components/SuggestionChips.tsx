"use client";

export const DEFAULT_SUGGESTIONS = [
  "thank you",
  "where",
  "water",
  "good morning",
] as const;

interface SuggestionChipsProps {
  suggestions?: readonly string[];
  activeQuery?: string;
  onSelect: (term: string) => void;
}

export function SuggestionChips({
  suggestions = DEFAULT_SUGGESTIONS,
  activeQuery = "",
  onSelect,
}: SuggestionChipsProps) {
  const normalizedActive = activeQuery.trim().toLowerCase();

  return (
    <div
      className="mt-3 flex flex-wrap items-center gap-2"
      aria-label="Quick search suggestions"
    >
      <span className="text-xs font-medium text-slate-500">Try:</span>
      {suggestions.map((term) => {
        const isActive = normalizedActive === term.toLowerCase();
        return (
          <button
            key={term}
            type="button"
            onClick={() => onSelect(term)}
            aria-pressed={isActive}
            className={`rounded-full border px-3 py-1 text-xs font-medium transition focus:outline-none focus:ring-2 focus:ring-indigo-300 ${
              isActive
                ? "border-indigo-500 bg-indigo-50 text-indigo-700"
                : "border-slate-200 bg-white text-slate-700 hover:border-indigo-300 hover:bg-indigo-50/60 hover:text-indigo-700"
            }`}
          >
            {term}
          </button>
        );
      })}
    </div>
  );
}
