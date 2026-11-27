"""Check that a list of words is in the dictionary and that an English search finds them.

Usage (API running on :8080):
    python scripts/check_words.py scripts/wordlists/conjunctions.tsv [--db backend/dictionary.db] [--out report.md]

The word list is tab-separated: traditional, expected Jyutping, register (spoken/formal/both),
English search query, note. Lines starting with # are comments.
"""
import argparse
import json
import sqlite3
import sys
import urllib.parse
import urllib.request

API = "http://localhost:8080/api/search"


def search(query, limit=20):
    url = API + "?" + urllib.parse.urlencode({"q": query, "limit": limit})
    with urllib.request.urlopen(url, timeout=10) as resp:
        return json.load(resp)["entries"]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("wordlist")
    ap.add_argument("--db", default="backend/dictionary.db")
    ap.add_argument("--out")
    args = ap.parse_args()
    sys.stdout.reconfigure(encoding="utf-8")

    db = sqlite3.connect(args.db)
    rows = []
    with open(args.wordlist, encoding="utf-8") as f:
        for line in f:
            if not line.strip() or line.startswith("#"):
                continue
            trad, jp, register, query, note = (line.rstrip("\n").split("\t") + [""] * 5)[:5]
            found = db.execute(
                "SELECT jyutping, source, english_json, freq FROM entries WHERE traditional = ?", (trad,)
            ).fetchall()
            readings = [r[0] for r in found]
            in_dict = bool(found)
            reading_ok = jp in readings
            match = next((r for r in found if r[0] == jp), found[0] if found else None)
            meaning = "; ".join(json.loads(match[2]))[:70] if match else ""
            results = [e["traditional"] for e in search(query)] if query else []
            rank = results.index(trad) + 1 if trad in results else None
            rows.append({
                "trad": trad, "jp": jp, "register": register, "query": query, "note": note,
                "in_dict": in_dict, "reading_ok": reading_ok, "readings": readings,
                "source": match[1] if match else "", "freq": match[3] if match else 0,
                "meaning": meaning, "rank": rank, "top3": results[:3],
            })

    def status(r):
        if not r["in_dict"]:
            return "❌ missing"
        if not r["reading_ok"]:
            return "⚠️ other reading"
        if r["rank"] is None:
            return "⚠️ not in top 20"
        if r["rank"] > 5:
            return f"⚠️ rank {r['rank']}"
        return f"✅ rank {r['rank']}"

    lines = [
        "| Word | Jyutping | Register | Search | Result | In dictionary as | Top 3 for search |",
        "|---|---|---|---|---|---|---|",
    ]
    for r in rows:
        got = f"{', '.join(r['readings'])} ({r['source']}) {r['meaning']}" if r["in_dict"] else "—"
        lines.append(
            f"| {r['trad']} | {r['jp']} | {r['register']} | {r['query']} | {status(r)} | {got} | {'、'.join(r['top3'])} |"
        )
    ok = sum(1 for r in rows if status(r).startswith("✅"))
    missing = sum(1 for r in rows if not r["in_dict"])
    summary = f"{len(rows)} words: {ok} found in the top 5 of their search, {missing} missing from the dictionary."
    report = summary + "\n\n" + "\n".join(lines) + "\n"
    print(report)
    if args.out:
        with open(args.out, "w", encoding="utf-8") as f:
            f.write(report)


if __name__ == "__main__":
    main()
