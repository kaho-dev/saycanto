import Link from "next/link";
import { CopyButton } from "@/shared/ui/CopyButton";
import { ListenButton } from "@/shared/ui/ListenButton";
import type { DictionaryEntry } from "../types";
import { WORD_TYPE_LABELS } from "../wordTypes";
import { RegisterBadge } from "./RegisterBadge";

/** "conjunction, preposition" from wordTypes, or the curated partOfSpeech when there are none. */
function typeLabel(entry: DictionaryEntry): string {
  if (entry.wordTypes && entry.wordTypes.length > 0) {
    return entry.wordTypes.map((t) => WORD_TYPE_LABELS[t].toLowerCase()).join(", ");
  }
  return entry.partOfSpeech ?? "";
}

export function EntryCard({ entry }: { entry: DictionaryEntry }) {
  return (
    <article className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-indigo-300 sm:p-5">
      <div className="flex flex-wrap items-start justify-between gap-2 sm:gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <Link
              href={`/word/${encodeURIComponent(entry.id)}`}
              className="break-words text-xl font-bold tracking-tight text-indigo-700 hover:underline sm:text-2xl"
            >
              {entry.jyutping}
            </Link>
            <CopyButton text={entry.jyutping} label={`Jyutping ${entry.jyutping}`} />
            <ListenButton text={entry.traditional} label={entry.traditional} />
          </div>
          <p className="mt-0.5 text-sm text-slate-500">
            {entry.traditional}
            {entry.simplified && entry.simplified !== entry.traditional
              ? ` · ${entry.simplified}`
              : ""}
            {typeLabel(entry) ? ` · ${typeLabel(entry)}` : ""}
          </p>
        </div>
        <RegisterBadge register={entry.register} />
      </div>

      <p className="mt-3 text-base font-medium text-slate-900">
        {entry.english.join("; ")}
      </p>

      {entry.classifier && (
        <p className="mt-2 text-xs text-slate-600">
          <span className="font-semibold">Classifier:</span> {entry.classifier}
        </p>
      )}

      {entry.notes && (
        <p className="mt-2 text-sm text-slate-600">{entry.notes}</p>
      )}

      {entry.examples && entry.examples.length > 0 && (
        <div className="mt-3 rounded-lg bg-slate-50 p-3 text-sm">
          <div className="flex flex-wrap items-center gap-2">
            <p className="font-medium text-indigo-900">
              {entry.examples[0].jyutping}
            </p>
            <ListenButton
              text={entry.examples[0].cantonese}
              label={`example: ${entry.examples[0].cantonese}`}
            />
          </div>
          <p className="text-slate-600">{entry.examples[0].cantonese}</p>
          <p className="mt-0.5 text-slate-700">{entry.examples[0].english}</p>
        </div>
      )}
    </article>
  );
}
