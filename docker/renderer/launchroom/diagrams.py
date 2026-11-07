"""Ready-made shots made of structure: code walkthrough, flow diagram, terminal, chat."""

from manim import (
    DOWN, LEFT, ORIGIN, RIGHT, UP,
    AddTextLetterByLetter, Arrow, Code, Create, CurvedArrow, FadeIn, FadeOut, Indicate, LaggedStart,
    Paragraph, RoundedRectangle, SurroundingRectangle, VGroup,
)

from .style import ACCENT, BACKGROUND, GOOD, INK, MONO, MUTED, WARM, box, label, wrap


class Diagrams:
    def code(self, source: str, steps: list, language: str = "go", heading: str = ""):
        """Exact, syntax-highlighted code; each Step boxes its lines (1-based) while it is said."""
        source = source.strip("\n")
        block = Code(code_string=source, language=language, formatter_style="monokai", add_line_numbers=True,
                     background="rectangle", paragraph_config={"font": MONO})
        block.scale_to_fit_height(min(5.4, 0.4 * (source.count("\n") + 2)))
        block.scale_to_fit_width(min(block.width, 12.5))
        group = VGroup(block)
        if heading:
            group = VGroup(label(heading, scale=0.6, color=ACCENT), block).arrange(DOWN, buff=0.35)
        group.move_to(UP * 0.45)  # clear of the subtitles at the bottom
        self.play(FadeIn(group), run_time=0.6)
        lines = block.code_lines
        marker = None
        for step in steps:
            with self.narrate(step.narration) as t:
                picked = [lines[n - 1] for n in step.lines if 0 < n <= len(lines)]
                if not picked:
                    continue
                target = SurroundingRectangle(VGroup(*picked), color=WARM, buff=0.08, corner_radius=0.06)
                if marker is None:
                    marker = target
                    self.play(Create(marker), run_time=t.part(0.2))
                else:
                    self.play(marker.animate.become(target), run_time=t.part(0.2))
        self.play(FadeOut(group), *([FadeOut(marker)] if marker else []), run_time=0.5)

    def flow(self, nodes: list, steps: list, loop: bool = False):
        """Boxes joined left to right by arrows (and back to the start when loop=True). Each Step
        lights up the boxes numbered in its lines, 1-based, as it is said."""
        boxes = VGroup(*[box(n) for n in nodes]).arrange(RIGHT, buff=0.9)
        boxes.scale_to_fit_width(min(boxes.width, 12.5)).move_to(ORIGIN + UP * 0.3)
        arrows = VGroup(*[Arrow(a.get_right(), b.get_left(), buff=0.1, color=MUTED) for a, b in zip(boxes, boxes[1:])])
        back = VGroup()
        if loop and len(boxes) > 1:
            back = CurvedArrow(boxes[-1].get_bottom() + DOWN * 0.1, boxes[0].get_bottom() + DOWN * 0.1, angle=-1.2, color=ACCENT)
        self.play(LaggedStart(*[FadeIn(b) for b in boxes], lag_ratio=0.25), run_time=1.0)
        self.play(*[Create(a) for a in arrows], *([Create(back)] if loop else []), run_time=0.8)
        for step in steps:
            with self.narrate(step.narration) as t:
                for n in step.lines:
                    if 0 < n <= len(boxes):
                        self.play(Indicate(boxes[n - 1], color=WARM, scale_factor=1.12), run_time=t.part(0.3))
        self.play(FadeOut(VGroup(boxes, arrows, back)), run_time=0.5)

    def terminal(self, steps: list, heading: str = ""):
        """A terminal window. Each Step types its `show` as a command (a line starting "$ ") or
        prints it as output, while it is said."""
        window = RoundedRectangle(corner_radius=0.2, width=12, height=6.2, color=MUTED, stroke_width=2,
                                  fill_color="#151a22", fill_opacity=1)
        group = VGroup(window)
        if heading:
            group.add(label(heading, scale=0.45, color=MUTED).next_to(window.get_top(), DOWN, buff=0.2))
        self.play(FadeIn(group), run_time=0.5)
        shown = VGroup()
        for step in steps:
            text = step.show or step.narration
            command = text.startswith("$ ")
            line = Paragraph(*(wrap(text, 56) or [" "]), font=MONO, color=GOOD if command else INK, line_spacing=0.7).scale(0.47)
            if len(shown):
                line.next_to(shown, DOWN, buff=0.2, aligned_edge=LEFT)
            else:
                line.move_to(window.get_corner(UP + LEFT) + RIGHT * 0.4 + DOWN * 0.8, aligned_edge=UP + LEFT)
            with self.narrate(step.narration) as t:
                anim = AddTextLetterByLetter(line) if command else FadeIn(line)
                self.play(anim, run_time=t.part(0.35 if command else 0.15))
            shown.add(line)
            if shown.get_bottom()[1] < window.get_bottom()[1] + 0.4:
                self.play(shown.animate.shift(UP * (window.get_bottom()[1] + 0.4 - shown.get_bottom()[1])), run_time=0.3)
        self.play(FadeOut(VGroup(group, shown)), run_time=0.5)

    def chat(self, steps: list):
        """A conversation as message bubbles. Each Step's `show` is "who: message" (who is user,
        model or tool); user messages sit right, the others left."""
        shown = VGroup()
        colors = {"user": ACCENT, "model": INK, "tool": WARM}
        for step in steps:
            who, _, text = (step.show or step.narration).partition(": ")
            who = who.strip().lower()
            words = Paragraph(*wrap(text, 34), font=MONO if who == "tool" else "Noto Sans", color=BACKGROUND, line_spacing=0.75).scale(0.58)
            bubble = RoundedRectangle(corner_radius=0.2, width=words.width + 0.7, height=words.height + 0.5,
                                      fill_color=colors.get(who, INK), fill_opacity=1, stroke_width=0)
            message = VGroup(bubble, words.move_to(bubble))
            tag = label(who, scale=0.42, color=MUTED).next_to(message, UP, buff=0.08, aligned_edge=LEFT)
            item = VGroup(tag, message)
            side = RIGHT if who == "user" else LEFT
            if len(shown):
                item.next_to(shown, DOWN, buff=0.45)
            else:
                item.to_edge(UP, buff=0.8)
            item.to_edge(side, buff=1.6)
            with self.narrate(step.narration) as t:
                self.play(FadeIn(item, shift=UP * 0.2), run_time=t.part(0.2))
            shown.add(item)
            if shown.get_bottom()[1] < -3.4:
                self.play(shown.animate.shift(UP * (-3.4 - shown.get_bottom()[1])), run_time=0.3)
        self.play(FadeOut(shown), run_time=0.5)
