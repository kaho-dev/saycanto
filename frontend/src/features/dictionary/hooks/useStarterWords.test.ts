import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as api from "../api/dictionaryApi";
import { useStarterWords } from "./useStarterWords";

describe("useStarterWords", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("does not fetch when disabled", () => {
    const spy = vi.spyOn(api, "listWords");
    const { result } = renderHook(() => useStarterWords("", false));

    expect(spy).not.toHaveBeenCalled();
    expect(result.current.entries).toEqual([]);
    expect(result.current.loading).toBe(false);
  });

  it("fetches starter words when enabled", async () => {
    const spy = vi.spyOn(api, "listWords").mockResolvedValue({
      total: 1,
      limit: 8,
      offset: 0,
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

    const { result } = renderHook(() => useStarterWords("spoken", true, 8));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.entries).toHaveLength(1);
    });

    expect(spy).toHaveBeenCalledWith(
      expect.objectContaining({ limit: 8, register: "spoken" })
    );
    expect(result.current.entries[0].jyutping).toBe("m4 goi1");
  });

  it("resets to empty entries on fetch error", async () => {
    vi.spyOn(api, "listWords").mockRejectedValue(new Error("Offline"));

    const { result } = renderHook(() => useStarterWords("", true));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.entries).toEqual([]);
  });
});
