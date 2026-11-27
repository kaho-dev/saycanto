"use client";

import { useEffect, useState } from "react";
import { listWords } from "../api/dictionaryApi";
import type { DictionaryEntry, RegisterType } from "../types";

export interface UseStarterWordsState {
  entries: DictionaryEntry[];
  loading: boolean;
}

export function useStarterWords(
  register: RegisterType,
  enabled: boolean,
  limit = 8
): UseStarterWordsState {
  const [entries, setEntries] = useState<DictionaryEntry[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const controller = new AbortController();
    setLoading(true);

    listWords({ limit, register, signal: controller.signal })
      .then((res) => {
        setEntries(res.entries);
      })
      .catch(() => {
        if (!controller.signal.aborted) {
          setEntries([]);
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      });

    return () => {
      controller.abort();
    };
  }, [register, enabled, limit]);

  return { entries, loading };
}
