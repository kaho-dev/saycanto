"use client";

import { EntryCard } from "@/features/dictionary/components/EntryCard";
import { SearchBar } from "@/features/dictionary/components/SearchBar";
import { SuggestionChips } from "@/features/dictionary/components/SuggestionChips";
import { WordTypeChips } from "@/features/dictionary/components/WordTypeChips";
import { useSearch } from "@/features/dictionary/hooks/useSearch";
import { useStarterWords } from "@/features/dictionary/hooks/useStarterWords";

export default function HomePage() {
  const {
    query,
    setQuery,
    register,
    setRegister,
    wordType,
    setWordType,
    entries,
    total,
    typeCounts,
    untyped,
    loading,
    error,
  } = useSearch();
  const isSearching = query.trim() !== "";
  const starter = useStarterWords(register, !isSearching);
  const displayedEntries = isSearching ? entries : starter.entries;

  return (
    <main className="mx-auto max-w-3xl px-4 py-8 sm:py-12">
      <header className="mb-6 text-center sm:mb-8">
        <h1 className="text-3xl font-extrabold tracking-tight text-slate-900 sm:text-4xl">
          Say<span className="text-indigo-600">Canto</span>
        </h1>
        <p className="mt-2 text-sm text-slate-600">
          Open-source English → Cantonese dictionary with Jyutping first.
        </p>
      </header>

      <SearchBar
        query={query}
        onQueryChange={setQuery}
        register={register}
        onRegisterChange={setRegister}
      />

      <SuggestionChips activeQuery={query} onSelect={setQuery} />

      {isSearching && (
        <WordTypeChips
          counts={typeCounts}
          untyped={untyped}
          active={wordType}
          onChange={setWordType}
        />
      )}

      <section className="mt-6 space-y-4">
        {isSearching && loading && (
          <p className="text-center text-sm text-slate-500">Searching...</p>
        )}
        {isSearching && error && (
          <p className="rounded-xl bg-red-50 p-4 text-sm text-red-700">
            {error}
          </p>
        )}
        {isSearching && !loading && !error && entries.length === 0 && (
          <p className="text-center text-sm text-slate-500">
            No entries found for &ldquo;{query}&rdquo;
            {wordType ? " with this word type" : ""}.
          </p>
        )}
        {isSearching && !error && entries.length > 0 && total > entries.length && (
          <p className="text-xs text-slate-500">
            Showing the top {entries.length} of {total} matches.
          </p>
        )}
        {!isSearching && starter.loading && starter.entries.length === 0 && (
          <p className="text-center text-sm text-slate-500">
            Loading starter words...
          </p>
        )}
        {!isSearching && starter.entries.length > 0 && (
          <div className="flex items-center justify-between">
            <h2 className="text-xs font-semibold uppercase tracking-wider text-slate-500">
              Starter words
            </h2>
          </div>
        )}
        {displayedEntries.map((entry) => (
          <EntryCard key={entry.id} entry={entry} />
        ))}
      </section>
    </main>
  );
}
