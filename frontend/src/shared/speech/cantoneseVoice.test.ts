import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { findCantoneseVoice, getSpeechSynthesis, useCantoneseVoice } from "./cantoneseVoice";
import { installSpeech, uninstallSpeech, voice } from "./speechMock";

describe("findCantoneseVoice", () => {
  it("accepts zh-HK and yue language tags", () => {
    expect(findCantoneseVoice([voice("Tracy", "zh-HK")])?.name).toBe("Tracy");
    expect(findCantoneseVoice([voice("Sinji", "zh_HK")])?.name).toBe("Sinji");
    expect(findCantoneseVoice([voice("G", "yue-Hant-HK")])?.name).toBe("G");
    expect(findCantoneseVoice([voice("Y", "yue")])?.name).toBe("Y");
  });

  it("accepts a voice named Cantonese whatever its tag", () => {
    expect(findCantoneseVoice([voice("Google 粵語（香港）", "zh")])?.name).toBe("Google 粵語（香港）");
    expect(findCantoneseVoice([voice("Cantonese Female", "zh")])?.name).toBe("Cantonese Female");
  });

  it("never picks a Mandarin voice", () => {
    expect(
      findCantoneseVoice([voice("Huihui", "zh-CN"), voice("Hanhan", "zh-TW"), voice("David", "en-US")])
    ).toBeNull();
  });

  it("prefers an on-device voice", () => {
    const picked = findCantoneseVoice([voice("Online", "zh-HK", false), voice("Local", "zh-HK", true)]);
    expect(picked?.name).toBe("Local");
    expect(findCantoneseVoice([voice("Online", "zh-HK", false)])?.name).toBe("Online");
  });
});

describe("useCantoneseVoice", () => {
  afterEach(uninstallSpeech);

  it("returns null without speech support", () => {
    uninstallSpeech();
    expect(getSpeechSynthesis()).toBeNull();
    const { result } = renderHook(() => useCantoneseVoice());
    expect(result.current).toBeNull();
  });

  it("picks up voices that load later and unsubscribes on unmount", () => {
    const speech = installSpeech([]);
    const { result, unmount } = renderHook(() => useCantoneseVoice());
    expect(result.current).toBeNull();

    act(() => speech.loadVoices([voice("Tracy", "zh-HK")]));
    expect(result.current?.name).toBe("Tracy");

    unmount();
    expect(speech.listenerCount()).toBe(0);
  });
});
