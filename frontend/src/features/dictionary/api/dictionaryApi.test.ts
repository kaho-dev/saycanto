import { describe, expect, it, vi, beforeEach } from "vitest";
import { getWordById, listWords, searchDictionary } from "./dictionaryApi";
import * as client from "@/shared/api/client";

describe("dictionaryApi", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("calls /api/search with query, limit, and register", async () => {
    const spy = vi.spyOn(client, "apiFetch").mockResolvedValue({
      query: "thank",
      total: 1,
      entries: [],
    });

    await searchDictionary({ query: "thank", register: "spoken", limit: 10 });
    expect(spy).toHaveBeenCalledWith(
      "/api/search?q=thank&limit=10&register=spoken",
      { signal: undefined }
    );
  });

  it("omits register when not provided", async () => {
    const spy = vi.spyOn(client, "apiFetch").mockResolvedValue({
      query: "water",
      total: 0,
      entries: [],
    });

    await searchDictionary({ query: "water" });
    expect(spy).toHaveBeenCalledWith("/api/search?q=water&limit=20", {
      signal: undefined,
    });
  });

  it("calls /api/words/{id} with encoded id", async () => {
    const spy = vi.spyOn(client, "apiFetch").mockResolvedValue({
      id: "m4-goi1",
      english: ["thank you"],
      jyutping: "m4 goi1",
      traditional: "唔該",
      simplified: "唔该",
    });

    const entry = await getWordById("m4-goi1");
    expect(spy).toHaveBeenCalledWith("/api/words/m4-goi1");
    expect(entry.jyutping).toBe("m4 goi1");
  });

  it("calls /api/words with default and custom parameters", async () => {
    const spy = vi.spyOn(client, "apiFetch").mockResolvedValue({
      total: 1,
      limit: 10,
      offset: 0,
      entries: [],
    });

    await listWords();
    expect(spy).toHaveBeenCalledWith("/api/words?limit=10&offset=0", {
      signal: undefined,
    });

    await listWords({ limit: 5, offset: 10, register: "spoken" });
    expect(spy).toHaveBeenCalledWith(
      "/api/words?limit=5&offset=10&register=spoken",
      { signal: undefined }
    );
  });
});
