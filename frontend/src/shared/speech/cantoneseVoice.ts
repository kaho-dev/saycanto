"use client";

import { useEffect, useState } from "react";

/**
 * Picks a Cantonese voice from the browser's speech voices.
 * Accepts Hong Kong Chinese (zh-HK) and Cantonese language tags (yue, yue-HK, yue-Hant-HK),
 * or a voice whose name says Cantonese. Mandarin voices (zh-CN, zh-TW) are never used:
 * they would read the characters in the wrong language.
 * Prefers a local (on-device) voice over a network one.
 */
export function findCantoneseVoice(
  voices: readonly SpeechSynthesisVoice[]
): SpeechSynthesisVoice | null {
  const isCantonese = (v: SpeechSynthesisVoice) => {
    const lang = v.lang.toLowerCase().replace(/_/g, "-");
    return (
      lang === "zh-hk" ||
      lang.startsWith("zh-hk-") ||
      lang === "yue" ||
      lang.startsWith("yue-") ||
      /cantonese|粵|廣東/i.test(v.name)
    );
  };
  const matches = voices.filter(isCantonese);
  return matches.find((v) => v.localService) ?? matches[0] ?? null;
}

export function getSpeechSynthesis(): SpeechSynthesis | null {
  if (typeof window === "undefined" || !("speechSynthesis" in window)) {
    return null;
  }
  return window.speechSynthesis;
}

/**
 * Returns the device's Cantonese voice, or null if there is none (or no speech support).
 * Browsers load voices asynchronously, so this listens for `voiceschanged`.
 */
export function useCantoneseVoice(): SpeechSynthesisVoice | null {
  const [voice, setVoice] = useState<SpeechSynthesisVoice | null>(null);

  useEffect(() => {
    const synth = getSpeechSynthesis();
    if (!synth) {
      return;
    }
    const update = () => setVoice(findCantoneseVoice(synth.getVoices()));
    update();
    synth.addEventListener("voiceschanged", update);
    return () => synth.removeEventListener("voiceschanged", update);
  }, []);

  return voice;
}
