"use client";

import { useEffect, useRef, useState } from "react";
import { getSpeechSynthesis, useCantoneseVoice } from "../speech/cantoneseVoice";

interface ListenButtonProps {
  /** Chinese text to speak (characters, not Jyutping: speech engines read characters). */
  text: string;
  /** What is being read, for the accessible name, e.g. "唔該". */
  label?: string;
  /** Speaking rate; slightly slow by default for learners. */
  rate?: number;
  className?: string;
}

/**
 * Speaks Cantonese with the device's built-in voice. Renders nothing when the
 * device has no Cantonese voice, so users never hear a Mandarin reading.
 */
export function ListenButton({
  text,
  label = text,
  rate = 0.85,
  className = "",
}: ListenButtonProps) {
  const voice = useCantoneseVoice();
  const [speaking, setSpeaking] = useState(false);
  const utteranceRef = useRef<SpeechSynthesisUtterance | null>(null);

  useEffect(() => {
    return () => {
      if (utteranceRef.current) {
        getSpeechSynthesis()?.cancel();
      }
    };
  }, []);

  if (!voice || !text.trim()) {
    return null;
  }

  function handleListen() {
    const synth = getSpeechSynthesis();
    if (!synth || !voice) {
      return;
    }
    synth.cancel();
    const utterance = new SpeechSynthesisUtterance(text);
    utterance.voice = voice;
    utterance.lang = voice.lang;
    utterance.rate = rate;
    const done = () => {
      if (utteranceRef.current === utterance) {
        utteranceRef.current = null;
        setSpeaking(false);
      }
    };
    utterance.onend = done;
    utterance.onerror = done;
    utteranceRef.current = utterance;
    setSpeaking(true);
    synth.speak(utterance);
  }

  return (
    <button
      type="button"
      onClick={handleListen}
      aria-label={`Listen to ${label}`}
      title={`Computer voice (${voice.name}). May differ from the Jyutping for characters with several readings.`}
      className={`inline-flex items-center gap-1 rounded-lg border px-2 py-1 text-xs font-medium transition focus:outline-none focus:ring-2 focus:ring-indigo-300 ${
        speaking
          ? "border-indigo-300 bg-indigo-50 text-indigo-700"
          : "border-slate-200 bg-slate-50 text-slate-600 hover:border-indigo-300 hover:bg-indigo-50 hover:text-indigo-700"
      } ${className}`}
    >
      <span aria-hidden="true">🔊</span>
      <span>{speaking ? "Playing…" : "Listen"}</span>
    </button>
  );
}
