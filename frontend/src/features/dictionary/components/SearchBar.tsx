"use client";

import type { RegisterType } from "../types";

interface SearchBarProps {
  query: string;
  onQueryChange: (q: string) => void;
  register: RegisterType;
  onRegisterChange: (r: RegisterType) => void;
}

export function SearchBar({
  query,
  onQueryChange,
  register,
  onRegisterChange,
}: SearchBarProps) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row">
      <input
        type="search"
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        placeholder="Search English or Jyutping (e.g. thank you, where, m4 goi1)..."
        aria-label="Search dictionary"
        className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-base text-slate-900 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-200"
      />
      <select
        value={register}
        onChange={(e) => onRegisterChange(e.target.value as RegisterType)}
        aria-label="Filter by register"
        className="w-full rounded-xl border border-slate-300 bg-white px-3 py-3 text-sm text-slate-700 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-200 sm:w-44"
      >
        <option value="">All registers</option>
        <option value="spoken">Spoken</option>
        <option value="formal">Formal</option>
      </select>
    </div>
  );
}
