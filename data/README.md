# SayCanto Dictionary Data

## Sources
1. **`custom.json`** (included in this repo): Curated English-first Cantonese entries with spoken/formal register tags, classifiers, usage notes, and natural example sentences.
2. **CC-Canto** (optional bulk import, ~34k entries): download [cccanto-170202.zip](https://cantonese.org/cccanto-170202.zip)
   from [cantonese.org/download.html](https://cantonese.org/download.html) (© Pleco Inc., **CC BY-SA 3.0**) and unzip
   `cccanto-webdist.txt` into this folder. It is gitignored; don't commit it.

3. **CC-CEDICT** (optional, ~125k Chinese–English entries, MDBG and contributors, **CC BY-SA 4.0**): download
   `cedict_1_0_ts_utf-8_mdbg.zip` **by hand in your browser** from
   [mdbg.net/chinese/dictionary?page=cc-cedict](https://www.mdbg.net/chinese/dictionary?page=cc-cedict)
   (MDBG prohibits scripted downloads) and save it here; the importer reads the zip directly.
4. **CC-CEDICT Cantonese readings** (needed with 3; © Pleco Software, **CC BY-SA 3.0**): download
   [cccedict-canto-readings-150923.zip](https://cantonese.org/cccedict-canto-readings-150923.zip) and unzip
   `cccedict-canto-readings-150923.txt` here. It is from 2015, so newer CC-CEDICT headwords have no Jyutping
   and are skipped (readings are never guessed).

5. **Word frequencies** (optional, ranking only; only a score per entry is stored, the lists are not redistributed):
   - [wordfreq](https://github.com/rspeer/wordfreq) `large_zh.msgpack.gz` by Robyn Speer, data **CC BY-SA 4.0**
     (snapshot to ~2021). Its sources must be credited, including SUBTLEX and the Google Books Ngram Viewer
     (https://books.google.com/ngrams); see the wordfreq README.
   - [HKCanCor](https://github.com/fcbond/hkcancor) `data/hkcancor-utf8.zip`, the Hong Kong Cantonese Corpus by
     K. K. Luke and May L. Y. Wong, **CC BY 4.0**. Cite: Luke & Wong (2015), *The Hong Kong Cantonese Corpus: Design and Uses*.

6. **English word types** (optional, used only to tag nouns/adjectives/adverbs; not redistributed):
   [Open English WordNet](https://github.com/globalwordnet/english-wordnet) 2025 edition,
   `english-wordnet-2025.zip`, **CC BY 4.0**. Cite: McCrae et al., *English WordNet 2019 – An Open-Source WordNet
   for English* (GWC 2019).

All downloaded files are gitignored. The combined dictionary data is distributed under **CC BY-SA 4.0**
(CC BY-SA 3.0 allows adaptations under later versions). Credit all three sources wherever the data is shown.

## Importing Data
```bash
cd backend
go run ./cmd/importer -custom ../data/custom.json -out ./dictionary.db
```
With CC-Canto:
```bash
go run ./cmd/importer -cccanto ../data/cccanto-webdist.txt -custom ../data/custom.json -out ./dictionary.db
```

Full dictionary with frequency ranking (CC-Canto + CC-CEDICT + curated overlay + wordfreq + HKCanCor; add
`-wordfreq ../data/large_zh.msgpack.gz -hkcancor ../data/hkcancor-utf8.zip -wordnet ../data/english-wordnet-2025.zip`
to the command below):
```bash
go run ./cmd/importer -cccanto ../data/cccanto-webdist.txt   -cedict ../data/cedict_1_0_ts_utf-8_mdbg.zip -cedict-readings ../data/cccedict-canto-readings-150923.txt   -custom ../data/custom.json -out ./dictionary.db
```
