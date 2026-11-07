"""The narration voice: Microsoft Edge's free text-to-speech, cached by text and voice."""

import asyncio
import hashlib
import os
import time
from pathlib import Path

import av
import edge_tts

VOICE = os.environ.get("LAUNCHROOM_VOICE", "en-US-AndrewNeural")
CACHE = Path(os.environ.get("LAUNCHROOM_VOICE_CACHE", "/tmp/launchroom-voice"))


def speak(text: str, voice: str = VOICE) -> tuple[str, float]:
    """Writes the spoken text to an MP3 and returns its path and length in seconds."""
    CACHE.mkdir(parents=True, exist_ok=True)
    key = hashlib.sha256(f"{voice}\n{text}".encode()).hexdigest()[:24]
    path = CACHE / f"{key}.mp3"
    if not path.exists():
        partial = path.with_suffix(".part.mp3")
        synthesize(text, voice, partial)
        partial.rename(path)
    return str(path), duration(path)


def synthesize(text: str, voice: str, out: Path, attempts: int = 5) -> None:
    """The Edge service now and then answers with no audio; a retry a moment later works."""
    for attempt in range(1, attempts + 1):
        try:
            asyncio.run(edge_tts.Communicate(text, voice).save(str(out)))
            return
        except Exception as error:  # the service's failures come as several exception types
            if attempt == attempts:
                raise RuntimeError(f"text-to-speech failed {attempts} times for {text[:60]!r}: {error}") from error
            time.sleep(1.5 * attempt)


def duration(path: Path) -> float:
    with av.open(str(path)) as media:
        return media.duration / 1_000_000
