import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { CopyButton } from "./CopyButton";

describe("CopyButton", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("copies text to clipboard and shows temporary Copied! feedback", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText },
      configurable: true,
    });

    render(<CopyButton text="m4 goi1" label="Jyutping m4 goi1" />);
    const button = screen.getByRole("button", { name: "Copy Jyutping m4 goi1" });

    await act(async () => {
      fireEvent.click(button);
    });

    expect(writeText).toHaveBeenCalledWith("m4 goi1");
    expect(screen.getByText("Copied!")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Copied Jyutping m4 goi1" })
    ).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(1800);
    });

    expect(screen.getByText("Copy")).toBeInTheDocument();
  });

  it("handles missing or failing clipboard gracefully", async () => {
    Object.defineProperty(navigator, "clipboard", {
      value: undefined,
      configurable: true,
    });

    render(<CopyButton text="seoi2" />);
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Copy text" }));
    });
    expect(screen.getByText("Copy")).toBeInTheDocument();

    const writeText = vi.fn().mockRejectedValue(new Error("Denied"));
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText },
      configurable: true,
    });

    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Copy text" }));
    });
    expect(screen.getByText("Copy")).toBeInTheDocument();
  });
});
