import { apiFetch } from "@/shared/api/client";
import type {
  DictionaryEntry,
  ListWordsResult,
  RegisterType,
  SearchResult,
  WordType,
} from "../types";

export interface SearchParams {
  query: string;
  register?: RegisterType;
  wordType?: WordType | "";
  limit?: number;
  signal?: AbortSignal;
}

export interface ListWordsParams {
  limit?: number;
  offset?: number;
  register?: RegisterType;
  signal?: AbortSignal;
}

export async function searchDictionary({
  query,
  register,
  wordType,
  limit = 20,
  signal,
}: SearchParams): Promise<SearchResult> {
  const params = new URLSearchParams({
    q: query,
    limit: String(limit),
  });
  if (register) {
    params.set("register", register);
  }
  if (wordType) {
    params.set("type", wordType);
  }
  return apiFetch<SearchResult>(`/api/search?${params.toString()}`, { signal });
}

export async function listWords({
  limit = 10,
  offset = 0,
  register,
  signal,
}: ListWordsParams = {}): Promise<ListWordsResult> {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  if (register) {
    params.set("register", register);
  }
  return apiFetch<ListWordsResult>(`/api/words?${params.toString()}`, {
    signal,
  });
}

export async function getWordById(id: string): Promise<DictionaryEntry> {
  return apiFetch<DictionaryEntry>(`/api/words/${encodeURIComponent(id)}`);
}
