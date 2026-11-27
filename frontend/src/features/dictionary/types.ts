export type RegisterType = "spoken" | "formal" | "both" | "";

export type WordType =
  | "noun"
  | "verb"
  | "adjective"
  | "adverb"
  | "pronoun"
  | "preposition"
  | "conjunction"
  | "classifier"
  | "particle"
  | "interjection"
  | "number"
  | "expression";

export interface ExampleSentence {
  jyutping: string;
  cantonese: string;
  english: string;
}

export interface DictionaryEntry {
  id: string;
  english: string[];
  jyutping: string;
  traditional: string;
  simplified: string;
  partOfSpeech?: string;
  register?: RegisterType;
  classifier?: string;
  notes?: string;
  examples?: ExampleSentence[];
  /** Where the entry came from: hand-curated (custom.json) or bulk CC-Canto. Set by the importer. */
  source?: "custom" | "cc-canto" | "cc-cedict";
  /** Parts of speech worked out by the importer; missing when unknown. */
  wordTypes?: WordType[];
}

export interface SearchResult {
  query: string;
  /** All matches (for the active word type, if any), not just this page. */
  total: number;
  entries: DictionaryEntry[];
  /** Matches per word type, ignoring the word-type filter. */
  typeCounts?: Partial<Record<WordType, number>>;
  /** Matches with no known word type. */
  untyped?: number;
}

export interface ListWordsResult {
  total: number;
  limit: number;
  offset: number;
  entries: DictionaryEntry[];
}
