import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { FakeUtterance, installSpeech, uninstallSpeech, voice } from "../speech/speechMock";
import { ListenButton } from "./ListenButton";

describe("ListenButton", () => {
  afterEach(uninstallSpeech);

  it("renders nothing when the device has no Cantonese voice", () => {
    installSpeech([voice("Huihui", "zh-CN")]);
    const { container } = render(<ListenButton text="唔該" />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders nothing without speech support or text", () => {
    const { container } = render(<ListenButton text="唔該" />);
    expect(container).toBeEmptyDOMElement();
    installSpeech([voice("Tracy", "zh-HK")]);
    const { container: empty } = render(<ListenButton text="  " />);
    expect(empty).toBeEmptyDOMElement();
  });

  it("speaks the characters with the Cantonese voice and resets when done", () => {
    const tracy = voice("Tracy", "zh-HK");
    const { synth } = installSpeech([voice("David", "en-US"), tracy]);
    render(<ListenButton text="唔該" label="唔該" />);

    fireEvent.click(screen.getByRole("button", { name: "Listen to 唔該" }));

    expect(synth.cancel).toHaveBeenCalled();
    expect(synth.speak).toHaveBeenCalledTimes(1);
    const utterance = synth.speak.mock.calls[0][0] as FakeUtterance;
    expect(utterance.text).toBe("唔該");
    expect(utterance.voice).toBe(tracy);
    expect(utterance.lang).toBe("zh-HK");
    expect(utterance.rate).toBe(0.85);
    expect(screen.getByText("Playing…")).toBeInTheDocument();

    act(() => utterance.onend?.());
    expect(screen.getByText("Listen")).toBeInTheDocument();
  });

  it("ignores the end of an utterance that was replaced, and resets on error", () => {
    const { synth } = installSpeech([voice("Tracy", "zh-HK")]);
    render(<ListenButton text="多謝" rate={1} />);
    const button = screen.getByRole("button", { name: "Listen to 多謝" });

    fireEvent.click(button);
    fireEvent.click(button);
    const [first, second] = synth.speak.mock.calls.map((c) => c[0] as FakeUtterance);
    expect(second.rate).toBe(1);

    act(() => first.onend?.());
    expect(screen.getByText("Playing…")).toBeInTheDocument();

    act(() => second.onerror?.());
    expect(screen.getByText("Listen")).toBeInTheDocument();
  });

  it("stops speech when unmounted mid-playback", () => {
    const { synth } = installSpeech([voice("Tracy", "zh-HK")]);
    const { unmount } = render(<ListenButton text="水" />);
    fireEvent.click(screen.getByRole("button", { name: "Listen to 水" }));
    synth.cancel.mockClear();
    unmount();
    expect(synth.cancel).toHaveBeenCalledTimes(1);
  });
});
