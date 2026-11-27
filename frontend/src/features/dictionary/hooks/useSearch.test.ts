import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useSearch } from "./useSearch";
import * as api from "../api/dictionaryApi";

describe("useSearch", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.restoreAllMocks();
  });

  it("returns empty entries when query is blank", () => {
    const { result } = renderHook(() => useSearch(100));
    expect(result.current.entries).toEqual([]);
    expect(result.current.loading).toBe(false);
  });

  it("debounces search and populates entries", async () => {
    const spy = vi.spyOn(api, "searchDictionary").mockResolvedValue({
      query: "thank",
      total: 1,
      entries: [
        {
          id: "m4-goi1",
          english: ["thank you"],
          jyutping: "m4 goi1",
          traditional: "唔該",
          simplified: "唔该",
        },
      ],
    });

    const { result } = renderHook(() => useSearch(150));
    act(() => {
      result.current.setQuery("thank");
    });

    expect(spy).not.toHaveBeenCalled();

    await act(async () => {
      vi.advanceTimersByTime(150);
    });

    expect(spy).toHaveBeenCalledTimes(1);
    expect(result.current.entries).toHaveLength(1);
    expect(result.current.entries[0].jyutping).toBe("m4 goi1");
  });

  it("sets error message on API failure", async () => {
    vi.spyOn(api, "searchDictionary").mockRejectedValue(
      new Error("Network error")
    );

    const { result } = renderHook(() => useSearch(50));
    act(() => {
      result.current.setQuery("water");
    });

    await act(async () => {
      vi.advanceTimersByTime(50);
    });

    expect(result.current.error).toBe("Network error");
  });
});
