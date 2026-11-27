import { describe, expect, it, vi, beforeEach } from "vitest";
import { ApiError, apiFetch, getApiBaseUrl } from "./client";

describe("api client", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    delete process.env.NEXT_PUBLIC_API_URL;
  });

  it("uses default API base URL when env is unset", () => {
    expect(getApiBaseUrl()).toBe("http://localhost:8080");
  });

  it("uses NEXT_PUBLIC_API_URL when set", () => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.saycanto.org";
    expect(getApiBaseUrl()).toBe("https://api.saycanto.org");
  });

  it("returns parsed JSON on success", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ status: "ok" }),
      })
    );

    const data = await apiFetch<{ status: string }>("/api/health", {
      headers: { "X-Custom": "1" },
    });
    expect(data.status).toBe("ok");
  });

  it("throws ApiError with server message on error JSON", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({ error: "not found" }),
      })
    );

    await expect(apiFetch("/api/words/missing")).rejects.toEqual(
      new ApiError(404, "not found")
    );
  });

  it("throws ApiError with fallback message on non-JSON error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => {
          throw new Error("invalid json");
        },
      })
    );

    await expect(apiFetch("/api/fail")).rejects.toEqual(
      new ApiError(500, "Request failed with status 500")
    );
  });
});
