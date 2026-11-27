import type { WordType } from "./types";

/** Word types in display order, mirroring dictionary.WordTypes in the Go backend. */
export const WORD_TYPES: readonly WordType[] = [
  "noun",
  "verb",
  "adjective",
  "adverb",
  "pronoun",
  "preposition",
  "conjunction",
  "classifier",
  "particle",
  "interjection",
  "number",
  "expression",
] as const;

export const WORD_TYPE_LABELS: Record<WordType, string> = {
  noun: "Noun",
  verb: "Verb",
  adjective: "Adjective",
  adverb: "Adverb",
  pronoun: "Pronoun",
  preposition: "Preposition",
  conjunction: "Conjunction",
  classifier: "Classifier",
  particle: "Particle",
  interjection: "Interjection",
  number: "Number",
  expression: "Expression",
};
