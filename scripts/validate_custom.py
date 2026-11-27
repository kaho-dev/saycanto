#!/usr/bin/env python3
"""Check data/custom.json against the SayCanto entry rules (run in CI on every pull request).

Usage: python scripts/validate_custom.py [path/to/custom.json]

Uses only the standard library. Exits 1 and lists every problem if any entry breaks a rule.
The rules mirror CLAUDE.md "Data model" and "Cantonese content rules"; keep them in sync.
"""

import json
import re
import sys
from pathlib import Path

ALLOWED_KEYS = {
    "id", "english", "jyutping", "traditional", "simplified",
    "partOfSpeech", "register", "classifier", "notes", "examples",
}
REQUIRED_KEYS = {"id", "english", "jyutping", "traditional", "simplified"}
# Set by the importer, never written by hand.
IMPORTER_KEYS = {"source", "wordTypes", "frequency"}
REGISTERS = {"spoken", "formal", "both"}
EXAMPLE_KEYS = {"jyutping", "cantonese", "english"}

# One Jyutping syllable: lowercase letters followed by a tone 1-6 (m4, ng5, goi1).
SYLLABLE = re.compile(r"^[a-z]+[1-6]$")
# The id is the syllables joined by "-", optionally with a -2, -3 ... suffix for a second sense.
ID_SUFFIX = re.compile(r"-([2-9]|[1-9][0-9]+)$")
HAN = re.compile(r"[㐀-鿿豈-﫿\U00020000-\U0003134f]")
LATIN = re.compile(r"[A-Za-z]")


def check_jyutping(value, where, errors):
    if not isinstance(value, str) or not value.strip():
        errors.append(f"{where}: jyutping must be a non-empty string")
        return []
    if value != value.strip() or "  " in value:
        errors.append(f"{where}: jyutping has extra spaces: {value!r}")
    syllables = value.split()
    bad = [s for s in syllables if not SYLLABLE.match(s)]
    if bad:
        errors.append(
            f"{where}: jyutping syllables must be lowercase with a tone 1-6 on each "
            f"(e.g. 'm4 goi1'); bad: {', '.join(bad)}"
        )
    return syllables


def check_chinese(value, field, where, errors):
    if not isinstance(value, str) or not value.strip():
        errors.append(f"{where}: {field} must be a non-empty string")
    elif not HAN.search(value) or LATIN.search(value):
        errors.append(f"{where}: {field} must be Chinese characters, got {value!r}")


def check_entry(i, entry, errors):
    if not isinstance(entry, dict):
        errors.append(f"entry #{i}: must be an object")
        return None
    where = f"entry #{i} ({entry.get('id', '?')})"

    for key in sorted(set(entry) & IMPORTER_KEYS):
        errors.append(f"{where}: '{key}' is set by the importer; remove it")
    for key in sorted(set(entry) - ALLOWED_KEYS - IMPORTER_KEYS):
        errors.append(f"{where}: unknown field '{key}'")
    for key in sorted(REQUIRED_KEYS - set(entry)):
        errors.append(f"{where}: missing required field '{key}'")

    english = entry.get("english")
    if not isinstance(english, list) or not english or not all(
        isinstance(e, str) and e.strip() for e in english
    ):
        errors.append(f"{where}: english must be a non-empty list of non-empty strings")

    syllables = check_jyutping(entry.get("jyutping"), where, errors)
    check_chinese(entry.get("traditional"), "traditional", where, errors)
    check_chinese(entry.get("simplified"), "simplified", where, errors)

    entry_id = entry.get("id")
    if not isinstance(entry_id, str) or not entry_id:
        errors.append(f"{where}: id must be a non-empty string")
    elif syllables:
        expected = "-".join(syllables)
        if ID_SUFFIX.sub("", entry_id) != expected:
            errors.append(f"{where}: id should be '{expected}' (Jyutping joined by '-', "
                          f"plus -2, -3 for another sense)")

    register = entry.get("register")
    if register is not None and register not in REGISTERS:
        errors.append(f"{where}: register must be one of {sorted(REGISTERS)}, got {register!r}")

    for key in ("partOfSpeech", "classifier", "notes"):
        if key in entry and (not isinstance(entry[key], str) or not entry[key].strip()):
            errors.append(f"{where}: {key} must be a non-empty string (or left out)")

    examples = entry.get("examples", [])
    if not isinstance(examples, list):
        errors.append(f"{where}: examples must be a list")
        examples = []
    for j, ex in enumerate(examples, 1):
        ex_where = f"{where} example {j}"
        if not isinstance(ex, dict):
            errors.append(f"{ex_where}: must be an object")
            continue
        if set(ex) != EXAMPLE_KEYS:
            errors.append(f"{ex_where}: needs exactly {sorted(EXAMPLE_KEYS)}, got {sorted(ex)}")
        for key in EXAMPLE_KEYS & set(ex):
            if not isinstance(ex[key], str) or not ex[key].strip():
                errors.append(f"{ex_where}: {key} must be a non-empty string")
        if isinstance(ex.get("cantonese"), str) and not HAN.search(ex["cantonese"]):
            errors.append(f"{ex_where}: cantonese must be written in Chinese characters")

    return entry_id, entry.get("traditional"), entry.get("jyutping")


def validate(path):
    try:
        data = json.loads(Path(path).read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        return [f"{path}: cannot read JSON: {exc}"]
    if not isinstance(data, list):
        return [f"{path}: top level must be a JSON array of entries"]

    errors = []
    seen_ids, seen_words = {}, {}
    for i, entry in enumerate(data, 1):
        keys = check_entry(i, entry, errors)
        if not keys:
            continue
        entry_id, trad, jyut = keys
        if isinstance(entry_id, str):
            if entry_id in seen_ids:
                errors.append(f"entry #{i}: duplicate id '{entry_id}' (first at #{seen_ids[entry_id]})")
            seen_ids.setdefault(entry_id, i)
        word = (trad, jyut)
        if all(isinstance(w, str) for w in word):
            if word in seen_words:
                errors.append(f"entry #{i}: {trad} {jyut} already exists at #{seen_words[word]}; "
                              f"add the meaning there instead")
            seen_words.setdefault(word, i)
    return errors


def main():
    sys.stdout.reconfigure(encoding="utf-8")
    path = sys.argv[1] if len(sys.argv) > 1 else str(Path(__file__).resolve().parent.parent / "data" / "custom.json")
    errors = validate(path)
    if errors:
        print(f"{len(errors)} problem(s) in {path}:")
        for e in errors:
            print(f"  - {e}")
        sys.exit(1)
    print(f"OK: {path} passes all checks")


if __name__ == "__main__":
    main()
