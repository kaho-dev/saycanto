import Link from "next/link";
import { getWordById } from "@/features/dictionary/api/dictionaryApi";
import { EntryCard } from "@/features/dictionary/components/EntryCard";

export default async function WordDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  let entry = null;
  try {
    entry = await getWordById(id);
  } catch {
    entry = null;
  }

  return (
    <main className="mx-auto max-w-3xl px-4 py-12">
      <Link
        href="/"
        className="mb-6 inline-block text-sm font-medium text-indigo-600 hover:underline"
      >
        ← Back to search
      </Link>
      {entry ? (
        <EntryCard entry={entry} />
      ) : (
        <p className="rounded-xl border border-slate-200 bg-white p-6 text-slate-600">
          Word not found.
        </p>
      )}
    </main>
  );
}
