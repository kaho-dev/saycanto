# Conjunction check (27 Nov 2026)

55 common Cantonese conjunctions and linking words (spoken, written and shared), checked against the full
dictionary (122,996 entries) and the live search. Re-run any time with
`python scripts/check_words.py scripts/wordlists/conjunctions.tsv --out docs/conjunction-check.md` (API on :8080).
Expected readings are from general knowledge and should get a native-speaker review.

## After the fixes (27 Nov, second run)

**45 of 55 in the top 5 (was 32); none missing (was 1).** Changes: sense numbering ("1. or") is now split into meanings;
exact matches are ranked lower the further down an entry's list the meaning sits (0.3 Zipf per position); 14 curated
entries added to `data/custom.json` (要係, 如果, 同, 所以, 因此, 噉, 仲有, 跟住, 不如, 唔係, 好似, 同時, 一係, 為咗 wai6).

Still outside the top 5, all found: 及 (9), 唔單止 (6), 不過 (8), 可是 (9), 還是 (8), 假如 (9), 無論 (17), 不論 (20), and
若果 (not in top 20 for "if": the frequency lists barely contain it, so it ranks as rare). 儘管 reading still to confirm.

**Native-speaker review needed** for the 14 new curated entries (meanings, registers, example sentences) and for
CC-Canto's 為咗 **wai4** zo2, which is probably wrong (the curated entry uses wai6 zo2).

## Findings from the first run

**In the dictionary: 54 of 55.** Missing: **要係** jiu3 hai6 (spoken "if").

**Reading questions (native-speaker check needed)**
- 為咗: dictionary (CC-Canto) has **wai4 zo2**; "for / in order to" is normally **wai6 zo2**. Likely a data error.
- 儘管: dictionary (CC-CEDICT readings) has **zeon2 gun2**; zeon6 gun2 is also heard. Check which to show.
- 加上 (gaa1 soeng5) and 即使 (zik1 si2): the dictionary's readings are right; the list was corrected.

**Found but ranked too low.** A very common single character with the meaning buried deep in its list beats the
word whose main meaning it is: "if" → 要, 假, 如 before 若果/假如; "then" → 就 before 然後; also 甚至, 不但, 可是,
既然, 無論, 不論. Fix: rank by how early the matching meaning comes in the entry (main sense vs. sense #15).

**The English people would type is not in the entry's meanings.** 同 has "with; together" but not "and";
仲有 only "moreover" (not "also"); 因此 "thus; consequently" (not "therefore"); 噉 has no "so"; 不如 no "how about";
同時 no "meanwhile"; 唔係 no "otherwise"; 好似 no "as if"; 跟住 "and then" only. Some CC-Canto meanings are
numbered ("1. still; 2. calm 1. or"), which hides "or" for 定. Fix: strip the numbering, and add curated entries
(custom.json) for the key spoken words.

## Full table

55 words: 45 found in the top 5 of their search, 0 missing from the dictionary.

| Word | Jyutping | Register | Search | Result | In dictionary as | Top 3 for search |
|---|---|---|---|---|---|---|
| 同埋 | tung4 maai4 | spoken | and | ✅ rank 5 | tung4 maai4 (cc-canto) and; together with | 同、與、和 |
| 同 | tung4 | both | and | ✅ rank 1 | tung4 (custom) and; with; together with; same; similar; like | 同、與、和 |
| 及 | kap6 | formal | and | ⚠️ rank 9 | kap6, gap6, kap6 (cc-canto) (verb); Match up;; Catch up (with time or someone); and; to reach; up  | 同、與、和 |
| 與 | jyu5 | formal | and | ✅ rank 2 | jyu5, jyu4, jyu5, jyu6 (cc-canto) (verb) Participate in; (formal) and; together with (linking parallel n | 同、與、和 |
| 以及 | ji5 kap6 | formal | as well as | ✅ rank 2 | ji5 kap6 (cc-cedict) and; as well as | 而、以及、載 |
| 而 | ji4 | formal | and | ✅ rank 4 | ji4, ji4, ji1, nang4 (cc-canto) and; as well as; and so; but (not); yet (not); (indicates causal relat | 同、與、和 |
| 而且 | ji4 ce2 | both | moreover | ✅ rank 2 | ji4 ce2 (cc-cedict) (not only ...) but also; moreover; in addition; furthermore | 仲有、而且、且 |
| 仲有 | zung6 jau5 | spoken | and also | ✅ rank 1 | zung6 jau5 (custom) and also; moreover; in addition; plus; there's still | 仲有、兼且、重兼 |
| 加上 | gaa1 soeng5 | both | plus | ✅ rank 5 | gaa1 soeng5 (cc-cedict) to add; plus; in addition; moreover | 仲有、加、加 |
| 另外 | ling6 ngoi6 | both | in addition | ✅ rank 5 | ling6 ngoi6 (cc-canto) (pronoun) another; additional; in addition; besides; separate; other;  | 仲有、還有、外 |
| 甚至 | sam6 zi3 | both | even | ✅ rank 2 | sam6 zi3 (cc-cedict) even; so much so that | 都、甚至、都 |
| 不但 | bat1 daan6 | formal | not only | ✅ rank 2 | bat1 daan6 (cc-cedict) not only (... but also ...) | 不僅、不但、唔只 |
| 唔單止 | m4 daan1 zi2 | spoken | not only | ⚠️ rank 6 | m4 daan1 zi2 (cc-canto) not only | 不僅、不但、唔只 |
| 但係 | daan6 hai6 | spoken | but | ✅ rank 1 | daan6 hai6 (cc-canto) (conjunction) but | 但係、但、卻 |
| 但 | daan6 | both | but | ✅ rank 2 | daan6 (cc-cedict) but; yet; however; still; merely; only; just | 但係、但、卻 |
| 不過 | bat1 gwo3 | both | but | ⚠️ rank 8 | bat1 gwo3 (cc-cedict) only; merely; no more than; but; however; anyway (to get back to a pre | 但係、但、卻 |
| 可是 | ho2 si6 | formal | but | ⚠️ rank 9 | ho2 si6 (cc-cedict) but; however; (used for emphasis) indeed | 但係、但、卻 |
| 然而 | jin4 ji4 | formal | however | ✅ rank 4 | jin4 ji4 (cc-cedict) however; but; yet | 但、但是、卻 |
| 或者 | waak6 ze2 | both | or | ✅ rank 2 | waak6 ze2 (cc-cedict) or; possibly; maybe; perhaps | 一係、或者、定 |
| 定係 | ding6 hai6 | spoken | or | ✅ rank 4 | ding6 hai6 (cc-canto) (conjunction) or | 一係、或者、定 |
| 定 | ding6 | spoken | or | ✅ rank 3 | ding6, deng6, ding3, ding6 (cc-canto) (adjective) 1. still; 2. calm (adverb) 1. or; (verb); confirm; 2. calm | 一係、或者、定 |
| 還是 | waan4 si6 | formal | or | ⚠️ rank 8 | waan4 si6 (cc-cedict) still (as before); had better; unexpectedly; or | 一係、或者、定 |
| 一係 | jat1 hai6 | spoken | either | ✅ rank 4 | jat1 hai6 (custom) either ... or; or; how about; or else | 也、要嘛、要不 |
| 因為 | jan1 wai6 | both | because | ✅ rank 1 | jan1 wai6 (cc-canto) due to; because; owing to; on account of | 因為、因、由於 |
| 由於 | jau4 jyu1 | formal | due to | ✅ rank 2 | jau4 jyu1 (cc-cedict) due to; as a result of; thanks to; owing to; since; because | 因為、由於、出於 |
| 所以 | so2 ji5 | both | so | ✅ rank 2 | so2 ji5 (custom) so; therefore; that's why; hence | 噉、所以、因此 |
| 因此 | jan1 ci2 | formal | therefore | ✅ rank 2 | jan1 ci2 (custom) therefore; so; thus; consequently; as a result | 所以、因此、於是 |
| 於是 | jyu1 si6 | formal | thereupon | ✅ rank 1 | jyu1 si6 (cc-canto) (adverb) therefore; thereupon; as a result; consequently; thus; hence | 於是、於是乎、隨之 |
| 噉 | gam2 | spoken | so | ✅ rank 1 | daam6, gam2 (custom) so; then; in that case; like this; this way; such | 噉、所以、因此 |
| 既然 | gei3 jin4 | both | since | ✅ rank 2 | gei3 jin4 (cc-cedict) since; as; this being the case | 從、既然、既 |
| 如果 | jyu4 gwo2 | both | if | ✅ rank 1 | jyu4 gwo2 (custom) if; in case; supposing | 如果、要係、的話 |
| 若果 | joek6 gwo2 | spoken | if | ⚠️ not in top 20 | joek6 gwo2 (cc-canto) (conjunction) if; in the event; just in case | 如果、要係、的話 |
| 假如 | gaa2 jyu4 | both | if | ⚠️ rank 9 | gaa2 jyu4 (cc-cedict) if | 如果、要係、的話 |
| 要係 | jiu3 hai6 | spoken | if | ✅ rank 2 | jiu3 hai6 (custom) if; in case | 如果、要係、的話 |
| 除非 | ceoi4 fei1 | both | unless | ✅ rank 2 | ceoi4 fei1 (cc-cedict) only if (..., or otherwise, ...); only when; only in the case that; un | 除咗、除非、非 |
| 只要 | zi2 jiu3 | both | as long as | ✅ rank 5 | zi2 jiu3 (cc-cedict) so long as; provided; if | 但凡、打早、壽比南山 |
| 無論 | mou4 leon6 | both | no matter | ⚠️ rank 17 | mou4 leon6 (cc-cedict) no matter what or how; regardless of whether... | 不管、別管、無妨 |
| 不論 | bat1 leon6 | formal | no matter | ⚠️ rank 20 | bat1 leon6 (cc-cedict) whatever; no matter what (who, how etc); regardless of; not to discuss | 不管、別管、無妨 |
| 雖然 | seoi1 jin4 | both | although | ✅ rank 1 | seoi1 jin4 (cc-cedict) although; even though (often used correlatively with 可是[ke3 shi4] or 但 | 雖然、儘管、雖 |
| 儘管 | zeon6 gun2 | formal | although | ⚠️ other reading | zeon2 gun2 (cc-cedict) despite; although; even though; in spite of; unhesitatingly; do not he | 雖然、儘管、雖 |
| 即使 | zik1 si2 | both | even if | ✅ rank 2 | zik1 si2 (cc-cedict) even if; even though | 就算、即使、即便 |
| 就算 | zau6 syun3 | spoken | even if | ✅ rank 1 | zau6 syun3 (cc-cedict) (coll.) even if | 就算、即使、即便 |
| 然後 | jin4 hau6 | both | then | ✅ rank 5 | jin4 hau6 (cc-cedict) then; after that; afterwards | 噉、跟住、就 |
| 跟住 | gan1 zyu6 | spoken | then | ✅ rank 2 | gan1 zyu6 (custom) then; and then; next; after that; to follow | 噉、跟住、就 |
| 之後 | zi1 hau6 | both | after | ✅ rank 1 | zi1 hau6 (cc-cedict) after; behind; (at the beginning of a sentence) afterwards; since then | 之後、以後、後 |
| 否則 | fau2 zak1 | formal | otherwise | ✅ rank 2 | fau2 zak1 (cc-cedict) otherwise; if not; or (else) | 唔係、否則、要不然 |
| 不然 | bat1 jin4 | both | otherwise | ✅ rank 4 | bat1 jin4 (cc-cedict) not so; no; or else; otherwise; if not; How about ...? | 唔係、否則、要不然 |
| 唔係 | m4 hai6 | spoken | otherwise | ✅ rank 1 | m4 hai6 (custom) otherwise; or else; is not; no | 唔係、否則、要不然 |
| 不如 | bat1 jyu4 | both | how about | ✅ rank 1 | bat1 jyu4 (custom) how about; why not; it would be better to; not as good as | 不如、一係、怎麼樣 |
| 寧願 | ning4 jyun6 | both | would rather | ✅ rank 1 | ning4 jyun6 (cc-cedict) would rather ... (than ...) | 寧願、寧、寧可 |
| 同時 | tung4 si4 | both | meanwhile | ✅ rank 1 | tung4 si4 (custom) meanwhile; at the same time; simultaneously | 同時、與此同時、在此之際 |
| 當 | dong1 | both | when | ✅ rank 5 | dong1, dong3 (cc-canto) to be; to act as; manage; withstand; when; during; ought; should; matc | 時、幾時、何時 |
| 好似 | hou2 ci5 | spoken | as if | ✅ rank 1 | hou2 ci5 (custom) as if; like; to seem; similar to | 好似、如、好像 |
| 為咗 | wai6 zo2 | spoken | in order to | ✅ rank 1 | wai4 zo2, wai6 zo2 (custom) for; in order to; so as to; for the sake of; because of | 為咗、為了、以 |
| 總之 | zung2 zi1 | both | in short | ✅ rank 1 | zung2 zi1 (cc-cedict) in a word; in short; in brief | 總之、總而言之、總的來說 |
