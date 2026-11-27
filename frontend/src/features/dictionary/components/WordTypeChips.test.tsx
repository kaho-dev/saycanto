import { act, fireEvent, render, renderHook, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import * as api from "../api/dictionaryApi";
import * as client from "@/shared/api/client";
import { useSearch } from "../hooks/useSearch";
import { EntryCard } from "./EntryCard";
import { WordTypeChips } from "./WordTypeChips";

describe("WordTypeChips", () => {
  it("shows only types with matches, with counts, in display order", () => {
    render(
      <WordTypeChips
        counts={{ verb: 86, conjunction: 10, noun: 4 }}
        untyped={187}
        active=""
        onChange={() => {}}
      />
    );
    const buttons = screen.getAllByRole("button").map((b) => b.textContent);
    expect(buttons).toEqual(["All", "Noun 4", "Verb 86", "Conjunction 10"]);
    expect(screen.getByRole("button", { name: "All" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.queryByText(/no word type yet/)).not.toBeInTheDocument();
  });

  it("selects, toggles off, and resets a type", () => {
    const onChange = vi.fn();
    const { rerender } = render(
      <WordTypeChips counts={{ conjunction: 10 }} untyped={5} active="" onChange={onChange} />
    );
    fireEvent.click(screen.getByRole("button", { name: /Conjunction/ }));
    expect(onChange).toHaveBeenLastCalledWith("conjunction");

    rerender(
      <WordTypeChips counts={{ conjunction: 10 }} untyped={5} active="conjunction" onChange={onChange} />
    );
    expect(screen.getByRole("button", { name: /Conjunction/ })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByText("5 matches have no word type yet and are hidden by this filter.")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Conjunction/ }));
    expect(onChange).toHaveBeenLastCalledWith("");
    fireEvent.click(screen.getByRole("button", { name: "All" }));
    expect(onChange).toHaveBeenLastCalledWith("");
  });

  it("keeps the active type visible even when it has no matches, and uses singular wording", () => {
    render(<WordTypeChips counts={{}} untyped={1} active="particle" onChange={() => {}} />);
    expect(screen.getByRole("button", { name: /Particle 0/ })).toBeInTheDocument();
    expect(screen.getByText("1 match has no word type yet and is hidden by this filter.")).toBeInTheDocument();
  });

  it("renders nothing when no type has matches", () => {
    const { container } = render(
      <WordTypeChips counts={{}} untyped={3} active="" onChange={() => {}} />
    );
    expect(container).toBeEmptyDOMElement();
  });
});

describe("word types in the card, API and search hook", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("labels an entry with its word types, falling back to partOfSpeech", () => {
    const base = { english: ["and"], jyutping: "tung4", traditional: "同", simplified: "同" };
    const { rerender } = render(
      <EntryCard entry={{ ...base, id: "tung4", wordTypes: ["preposition", "conjunction"], partOfSpeech: "conjunction" }} />
    );
    expect(screen.getByText("同 · preposition, conjunction")).toBeInTheDocument();
    rerender(<EntryCard entry={{ ...base, id: "tung4-b", partOfSpeech: "conjunction" }} />);
    expect(screen.getByText("同 · conjunction")).toBeInTheDocument();
  });

  it("sends the word type as ?type=", async () => {
    const spy = vi.spyOn(client, "apiFetch").mockResolvedValue({ query: "so", total: 0, entries: [] });
    await api.searchDictionary({ query: "so", wordType: "conjunction" });
    expect(spy).toHaveBeenCalledWith("/api/search?q=so&limit=20&type=conjunction", { signal: undefined });
  });

  it("re-searches when the word type changes and exposes totals and counts", async () => {
    vi.useFakeTimers();
    const spy = vi.spyOn(api, "searchDictionary").mockResolvedValue({
      query: "so",
      total: 339,
      entries: [{ id: "so2-ji5", english: ["so"], jyutping: "so2 ji5", traditional: "所以", simplified: "所以" }],
      typeCounts: { conjunction: 10 },
      untyped: 187,
    });
    const { result } = renderHook(() => useSearch(10));
    act(() => result.current.setQuery("so"));
    await act(async () => {
      vi.advanceTimersByTime(10);
    });
    expect(result.current.total).toBe(339);
    expect(result.current.typeCounts).toEqual({ conjunction: 10 });
    expect(result.current.untyped).toBe(187);

    act(() => result.current.setWordType("conjunction"));
    await act(async () => {
      vi.advanceTimersByTime(10);
    });
    expect(spy).toHaveBeenLastCalledWith(expect.objectContaining({ query: "so", wordType: "conjunction" }));

    spy.mockResolvedValueOnce({ query: "x", total: 0, entries: [] });
    act(() => result.current.setQuery("x"));
    await act(async () => {
      vi.advanceTimersByTime(10);
    });
    expect(result.current.typeCounts).toEqual({});
    expect(result.current.untyped).toBe(0);

    act(() => result.current.setQuery(""));
    expect(result.current.total).toBe(0);
  });
});
