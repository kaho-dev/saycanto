import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { EntryCard } from "./EntryCard";
import { RegisterBadge } from "./RegisterBadge";
import { SearchBar } from "./SearchBar";
import { SuggestionChips } from "./SuggestionChips";

describe("dictionary components", () => {
  it("renders RegisterBadge for valid registers and null otherwise", () => {
    const { rerender, container } = render(<RegisterBadge register="spoken" />);
    expect(screen.getByText("Spoken")).toBeInTheDocument();

    rerender(<RegisterBadge register="formal" />);
    expect(screen.getByText("Formal")).toBeInTheDocument();

    rerender(<RegisterBadge register="both" />);
    expect(screen.getByText("Spoken & Formal")).toBeInTheDocument();

    rerender(<RegisterBadge register="" />);
    expect(container.firstChild).toBeNull();
  });

  it("handles SearchBar input and register changes", () => {
    const onQueryChange = vi.fn();
    const onRegisterChange = vi.fn();
    render(
      <SearchBar
        query=""
        onQueryChange={onQueryChange}
        register=""
        onRegisterChange={onRegisterChange}
      />
    );

    fireEvent.change(screen.getByLabelText("Search dictionary"), {
      target: { value: "water" },
    });
    expect(onQueryChange).toHaveBeenCalledWith("water");

    fireEvent.change(screen.getByLabelText("Filter by register"), {
      target: { value: "spoken" },
    });
    expect(onRegisterChange).toHaveBeenCalledWith("spoken");
  });

  it("renders EntryCard with Jyutping-first layout, copy button, simplified variant, classifier, notes, and example", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText },
      configurable: true,
    });

    render(
      <EntryCard
        entry={{
          id: "m4-goi1",
          english: ["thank you"],
          jyutping: "m4 goi1",
          traditional: "唔該",
          simplified: "唔该",
          partOfSpeech: "phrase",
          register: "spoken",
          classifier: "ge3 (個)",
          notes: "Polite phrase.",
          examples: [
            {
              jyutping: "m4 goi1 saai3",
              cantonese: "唔該晒",
              english: "thank you very much",
            },
          ],
        }}
      />
    );

    expect(screen.getByText("m4 goi1")).toBeInTheDocument();
    expect(screen.getByText("thank you")).toBeInTheDocument();
    expect(screen.getByText(/唔該 · 唔该/)).toBeInTheDocument();
    expect(screen.getByText(/ge3 \(個\)/)).toBeInTheDocument();
    expect(screen.getByText("Polite phrase.")).toBeInTheDocument();
    expect(screen.getByText("thank you very much")).toBeInTheDocument();

    await act(async () => {
      fireEvent.click(
        screen.getByRole("button", { name: "Copy Jyutping m4 goi1" })
      );
    });
    expect(writeText).toHaveBeenCalledWith("m4 goi1");
    expect(screen.getByText("Copied!")).toBeInTheDocument();
  });

  it("renders SuggestionChips and triggers onSelect when clicked", () => {
    const onSelect = vi.fn();
    render(<SuggestionChips activeQuery="water" onSelect={onSelect} />);

    const waterChip = screen.getByRole("button", { name: "water" });
    expect(waterChip).toHaveAttribute("aria-pressed", "true");

    const thankChip = screen.getByRole("button", { name: "thank you" });
    expect(thankChip).toHaveAttribute("aria-pressed", "false");

    fireEvent.click(thankChip);
    expect(onSelect).toHaveBeenCalledWith("thank you");
  });
});
