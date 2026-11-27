package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
	"github.com/kaho-dev/saycanto/backend/internal/importer"
	_ "modernc.org/sqlite"
)

func main() {
	ccCantoPath := flag.String("cccanto", "", "optional path to CC-Canto text file")
	cedictPath := flag.String("cedict", "", "optional path to CC-CEDICT (.txt/.u8 or the .zip from mdbg.net); needs -cedict-readings")
	readingsPath := flag.String("cedict-readings", "", "path to the CC-CEDICT Cantonese readings file from cantonese.org")
	wordfreqPath := flag.String("wordfreq", "", "optional wordfreq Chinese list (large_zh.msgpack.gz) for ranking by frequency")
	hkcancorPath := flag.String("hkcancor", "", "optional HKCanCor corpus zip (hkcancor-utf8.zip) for Cantonese frequencies")
	wordnetPath := flag.String("wordnet", "", "optional Open English WordNet zip (english-wordnet-2025.zip) to type nouns/adjectives/adverbs")
	customPath := flag.String("custom", "../data/custom.json", "path to curated custom JSON overlay")
	outPath := flag.String("out", "./dictionary.db", "output SQLite database path")
	flag.Parse()

	var baseEntries []dictionary.Entry
	if *ccCantoPath != "" {
		f, err := os.Open(*ccCantoPath)
		if err != nil {
			log.Fatalf("open cccanto: %v", err)
		}
		defer f.Close()
		baseEntries, err = importer.ParseCCCanto(f)
		if err != nil {
			log.Fatalf("parse cccanto: %v", err)
		}
		log.Printf("parsed %d CC-Canto entries", len(baseEntries))
	}

	if (*cedictPath == "") != (*readingsPath == "") {
		log.Fatalf("-cedict and -cedict-readings must be given together")
	}
	if *cedictPath != "" {
		rf, err := os.Open(*readingsPath)
		if err != nil {
			log.Fatalf("open cedict readings: %v", err)
		}
		readings, err := importer.ParseCantoReadings(rf)
		_ = rf.Close()
		if err != nil {
			log.Fatalf("parse cedict readings: %v", err)
		}
		log.Printf("parsed Cantonese readings for %d CC-CEDICT headwords", len(readings))

		cr, closeCedict, err := importer.OpenMaybeZip(*cedictPath)
		if err != nil {
			log.Fatalf("open cedict: %v", err)
		}
		cedictEntries, stats, err := importer.ParseCEDICT(cr, readings)
		_ = closeCedict()
		if err != nil {
			log.Fatalf("parse cedict: %v", err)
		}
		log.Printf("CC-CEDICT: %d lines → %d entries; skipped %d with no Cantonese reading, %d cross-references only",
			stats.Lines, stats.Entries, stats.NoReading, stats.OnlyReference)

		before := len(baseEntries)
		baseEntries = importer.CombineBase(baseEntries, cedictEntries)
		log.Printf("CC-CEDICT added %d new entries; %d merged into existing CC-Canto entries",
			len(baseEntries)-before, len(cedictEntries)-(len(baseEntries)-before))
	}

	var overlayEntries []dictionary.Entry
	if *customPath != "" {
		f, err := os.Open(*customPath)
		if err != nil {
			log.Fatalf("open custom json: %v", err)
		}
		defer f.Close()
		overlayEntries, err = importer.ParseCustomJSON(f)
		if err != nil {
			log.Fatalf("parse custom json: %v", err)
		}
		if err := importer.ValidateOverlay(overlayEntries); err != nil {
			log.Fatalf("invalid custom json: %v", err)
		}
		log.Printf("parsed %d custom overlay entries", len(overlayEntries))
	}

	merged := importer.MergeEntries(baseEntries, overlayEntries)

	var wf importer.WordfreqTable
	if *wordfreqPath != "" {
		f, err := os.Open(*wordfreqPath)
		if err != nil {
			log.Fatalf("open wordfreq: %v", err)
		}
		wf, err = importer.LoadWordfreq(f)
		_ = f.Close()
		if err != nil {
			log.Fatalf("load wordfreq: %v", err)
		}
		log.Printf("loaded wordfreq frequencies for %d Chinese words", len(wf))
	}
	var hk importer.CantoneseFreq
	if *hkcancorPath != "" {
		var err error
		hk, err = importer.LoadHKCanCor(*hkcancorPath)
		if err != nil {
			log.Fatalf("load hkcancor: %v", err)
		}
		log.Printf("loaded HKCanCor: %d words, %d word+reading frequencies", hk.Words, len(hk.ByReading))
	}
	if wf != nil || hk.Words > 0 {
		scored := importer.ApplyFrequencies(merged, wf, hk)
		log.Printf("frequency scores for %d of %d entries", scored, len(merged))
	}

	typed := importer.InferWordTypes(merged, hk)
	if *wordnetPath != "" {
		wn, err := importer.LoadWordNet(*wordnetPath)
		if err != nil {
			log.Fatalf("load wordnet: %v", err)
		}
		extra := importer.InferWordTypesFromWordNet(merged, wn)
		log.Printf("WordNet typed %d more entries", extra)
		typed += extra
	}
	perType := map[string]int{}
	for _, e := range merged {
		for _, t := range e.WordTypes {
			perType[t]++
		}
	}
	log.Printf("word types for %d of %d entries (%.0f%%)", typed, len(merged), 100*float64(typed)/float64(max(len(merged), 1)))
	for _, t := range dictionary.WordTypes {
		log.Printf("  %-12s %d", t, perType[t])
	}
	db, err := sql.Open("sqlite", *outPath)
	if err != nil {
		log.Fatalf("open output db: %v", err)
	}
	defer db.Close()

	if err := importer.BuildDatabase(context.Background(), db, merged); err != nil {
		log.Fatalf("build database: %v", err)
	}
	log.Printf("built %s with %d total entries", *outPath, len(merged))
}
