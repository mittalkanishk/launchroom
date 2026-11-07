"""launchroom: the scene library the video crew writes explainers with.

Every explainer shot is one file in the video's scenes/ folder defining `class Shot(Explainer)`.
The narration lives inside the scene: each beat speaks its sentence (free Edge voice), subtitles
it, and lasts exactly as long as the sentence. Ready-made shots:

    title(title, subtitle="", narration="")         statement(text, narration="")
    bullets(heading, steps)                          compare(left, right, steps)
    quote(text, source="", narration="")             number(value, caption, narration="")
    code(source, steps, language="go", heading="")   flow(nodes, steps, loop=False)
    terminal(steps, heading="")                      chat(steps)
    self.wait(seconds)                              a silent beat

steps are Step(narration, show="", lines=()). For anything else, write Manim inside a narrate block:

    with self.narrate("The loop calls the model.") as t:
        self.play(Create(arrow), run_time=t.part(0.4))
"""

from manim import *  # noqa: F401,F403  (re-exported: scenes use Manim directly too)

from .explainer import Explainer, Step
from .style import ACCENT, BAD, BACKGROUND, GOOD, INK, MUTED, WARM, box, label
