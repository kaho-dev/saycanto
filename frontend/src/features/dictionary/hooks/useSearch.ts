"use client";

import { useEffect, useState } from "react";
import { searchDictionary } from "../api/dictionaryApi";
import type { DictionaryEntry, RegisterType, WordType } from "../types";

export interface UseSearchState {
  query: string;
  setQuery: (q: string) => void;
  register: RegisterType;
  setRegister: (r: RegisterType) => void;
  wordType: WordType | "";
  setWordType: (t: WordType | "") => void;
  entries: DictionaryEntry[];
  total: number;
  typeCounts: Partial<Record<WordType, number>>;
  untyped: number;
  loading: boolean;
  error: string | null;
}

export function useSearch(debounceMs = 200): UseSearchState {
  const [query, setQuery] = useState("");
  const [register, setRegister] = useState<RegisterType>("");
  const [wordType, setWordType] = useState<WordType | "">("");
  const [entries, setEntries] = useState<DictionaryEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [typeCounts, setTypeCounts] = useState<Partial<Record<WordType, number>>>({});
  const [untyped, setUntyped] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const trimmed = query.trim();
    if (!trimmed) {
      setEntries([]);
      setTotal(0);
      setTypeCounts({});
      setUntyped(0);
      setError(null);
      setLoading(false);
      return;
    }

    const controller = new AbortController();
    const timer = setTimeout(async () => {
      setLoading(true);
      setError(null);
      try {
        const res = await searchDictionary({
          query: trimmed,
          register,
          wordType,
          signal: controller.signal,
        });
        setEntries(res.entries);
        setTotal(res.total);
        setTypeCounts(res.typeCounts ?? {});
        setUntyped(res.untyped ?? 0);
      } catch (err) {
        if ((err as Error).name !== "AbortError") {
          setError(err instanceof Error ? err.message : "Search failed");
        }
      } finally {
        setLoading(false);
      }
    }, debounceMs);

    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [query, register, wordType, debounceMs]);

  return {
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
  };
}
