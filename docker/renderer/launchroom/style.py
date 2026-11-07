"""The house style every explainer shares: a dark canvas, one accent, generous margins."""

from manim import RoundedRectangle, Text, VGroup

BACKGROUND = "#0e1117"
INK = "#e8e8ee"
MUTED = "#9aa0ae"
ACCENT = "#58c4dd"
WARM = "#f0b35e"
GOOD = "#83c167"
BAD = "#fc6255"
FONT = "Noto Sans"
MONO = "Noto Sans Mono"


def label(text: str, scale: float = 0.55, color: str = INK, mono: bool = False, bold: bool = False) -> Text:
    return Text(text, font=MONO if mono else FONT, color=color, weight="BOLD" if bold else "NORMAL").scale(scale)


def box(text: str, color: str = ACCENT) -> VGroup:
    """A rounded box with a label in it, sized to the label."""
    inner = label(text)
    frame = RoundedRectangle(corner_radius=0.18, width=max(2.2, inner.width + 0.6), height=max(1.1, inner.height + 0.5),
                             color=color, stroke_width=3)
    return VGroup(frame, inner)


def wrap(text: str, width: int) -> list:
    """Splits text into lines of about width characters, on word boundaries."""
    lines, line = [], ""
    for word in text.split():
        if line and len(line) + 1 + len(word) > width:
            lines.append(line)
            line = word
        else:
            line = f"{line} {word}".strip()
    return lines + ([line] if line else [])
