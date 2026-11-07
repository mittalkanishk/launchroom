"""Explainer: the base of every shot. narrate() ties animations to the spoken narration."""

from contextlib import contextmanager
from dataclasses import dataclass

from manim import Scene

from .diagrams import Diagrams
from .style import BACKGROUND
from .text import Texts
from .voice import speak


@dataclass
class Step:
    """One beat of a ready-made shot: what is said, and what it shows or points at.

    narration  the sentence spoken (and subtitled) during this beat
    show       the text this beat puts on screen, where the shot shows text (bullets, terminal, chat)
    lines      what the beat points at, 1-based: code lines in code(), boxes in flow()
    """
    narration: str
    show: str = ""
    lines: tuple = ()


class Tracker:
    """How long a narrate block's speech runs, for timing the animations inside it."""

    def __init__(self, seconds: float):
        self.duration = seconds

    def part(self, fraction: float) -> float:
        """A fraction of the speech, for one animation's run_time."""
        return max(0.1, self.duration * fraction)


class Explainer(Texts, Diagrams, Scene):
    """Every explainer shot is `class Shot(Explainer)` with a construct() that calls the ready-made
    shots (texts.py, diagrams.py) and, where they don't fit, plain Manim inside narrate() blocks."""

    def setup(self):
        self.camera.background_color = BACKGROUND

    @contextmanager
    def narrate(self, text: str):
        """Speaks text over the animations in the block, subtitles it, and makes the block last
        exactly as long as the speech: when the animations finish early the frame holds, so a
        sentence is never cut off. Time animations with the tracker: run_time=t.part(0.3)."""
        path, seconds = speak(text)
        self.add_sound(path)
        self.add_subcaption(text, duration=seconds)
        start = self.renderer.time
        yield Tracker(seconds)
        remaining = seconds - (self.renderer.time - start)
        if remaining > 0.01:
            self.wait(remaining)
