import { vi } from "vitest";

/** Test helper: installs a fake window.speechSynthesis and SpeechSynthesisUtterance (jsdom has neither). */
export function voice(
  name: string,
  lang: string,
  localService = true
): SpeechSynthesisVoice {
  return { name, lang, localService, default: false, voiceURI: name } as SpeechSynthesisVoice;
}

export class FakeUtterance {
  text: string;
  voice: SpeechSynthesisVoice | null = null;
  lang = "";
  rate = 1;
  onend: (() => void) | null = null;
  onerror: (() => void) | null = null;
  constructor(text: string) {
    this.text = text;
  }
}

export function installSpeech(initialVoices: SpeechSynthesisVoice[]) {
  let voices = initialVoices;
  const listeners = new Set<() => void>();
  const synth = {
    getVoices: vi.fn(() => voices),
    speak: vi.fn(),
    cancel: vi.fn(),
    addEventListener: vi.fn((_: string, fn: () => void) => listeners.add(fn)),
    removeEventListener: vi.fn((_: string, fn: () => void) => listeners.delete(fn)),
  };
  Object.defineProperty(window, "speechSynthesis", { value: synth, configurable: true });
  vi.stubGlobal("SpeechSynthesisUtterance", FakeUtterance);
  return {
    synth,
    /** Simulates the browser finishing its async voice load. */
    loadVoices(next: SpeechSynthesisVoice[]) {
      voices = next;
      listeners.forEach((fn) => fn());
    },
    listenerCount: () => listeners.size,
  };
}

export function uninstallSpeech() {
  // Remove the own property so `"speechSynthesis" in window` is false again.
  delete (window as unknown as Record<string, unknown>).speechSynthesis;
  vi.unstubAllGlobals();
}
