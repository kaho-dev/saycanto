# Frontend notes (Next.js)

Next.js 15 App Router · React 19 · TypeScript strict · Tailwind 3 · Vitest 3 + Testing Library (jsdom). Path alias `@/` → `src/`.

## Structure
- `src/app/page.tsx` — client component. `useSearch` (200 ms debounce, AbortController per keystroke) drives results;
  when the query is empty, `useStarterWords` shows the first 8 entries from `/api/words`.
- `src/app/word/[id]/page.tsx` — **server component**; fetches `getWordById` at request time (so `NEXT_PUBLIC_API_URL`
  must be reachable from the Next server, not just the browser). `params` is a Promise (Next 15).
- `src/features/dictionary/` — `api/dictionaryApi.ts` (URL building), `hooks/`, `components/` (EntryCard, RegisterBadge,
  SearchBar, SuggestionChips), `types.ts` (mirror of the Go `Entry`).
- `src/features/dictionary/components/WordTypeChips.tsx` + `wordTypes.ts` — word-type filter chips with counts,
  shown while searching; only types with matches appear, and the note says how many untyped matches a filter hides.
  Cards label word types after the characters (falls back to the curated `partOfSpeech`).
- `src/app/layout.tsx` — page shell and the data-credits footer (required by the data licences; keep it).
- `src/shared/api/client.ts` — `apiFetch<T>` + `ApiError`; surfaces the backend's `{"error"}` message.
  Production sets `NEXT_PUBLIC_API_URL` in Vercel to the Cloud Run URL.
- `src/shared/ui/CopyButton.tsx` — clipboard copy with a 1.8 s "Copied!" state.
- `src/shared/ui/ListenButton.tsx` + `src/shared/speech/cantoneseVoice.ts` — speaks Chinese characters with the device's
  Cantonese voice (Web Speech API); renders nothing if there is none. Never relax the voice filter to accept Mandarin
  (zh-CN/zh-TW). Tests mock speech with `src/shared/speech/speechMock.ts` (jsdom has no speech engine).

## Conventions
- Feature code stays in `src/features/<feature>/`; anything reused by two features moves to `src/shared/`.
- Components get data through props/hooks; never call `fetch` directly from a component.
- Tests live beside the file (`*.test.ts(x)`); mock `fetch` / the api module, not the network.
- Coverage counts only `src/features/**` and `src/shared/**` and only files a test imports (`all: false`);
  `src/app/**` pages are not measured. Don't read the 99 % as "the app is tested".
- Styling: Tailwind utilities only; indigo is the brand accent, register colours are emerald (spoken),
  amber (formal), sky (both).
