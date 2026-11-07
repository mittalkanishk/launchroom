"""Ready-made shots made of words: title, statement, bullets, comparison, quote, number."""

from manim import DOWN, LEFT, RIGHT, UP, FadeIn, FadeOut, Line, Paragraph, VGroup, Write

from .style import ACCENT, BAD, FONT, GOOD, INK, MUTED, WARM, label, wrap


class Texts:
    def title(self, title: str, subtitle: str = "", narration: str = ""):
        """A title card: the title written out, the subtitle fading in under it."""
        heading = label(title, scale=1.0, bold=True)
        heading.scale_to_fit_width(min(heading.width, 12))
        group = VGroup(heading)
        if subtitle:
            group.add(label(subtitle, color=ACCENT))
        group.arrange(DOWN, buff=0.45)
        with self.narrate(narration or title) as t:
            self.play(Write(heading), run_time=t.part(0.5))
            if subtitle:
                self.play(FadeIn(group[1], shift=UP * 0.2), run_time=t.part(0.2))
        self.play(FadeOut(group), run_time=0.4)

    def statement(self, text: str, narration: str = ""):
        """One big sentence on screen: the idea the shot is about."""
        words = Paragraph(*wrap(text, 30), font=FONT, color=INK, alignment="center", line_spacing=0.9).scale(0.8)
        with self.narrate(narration or text) as t:
            self.play(Write(words), run_time=t.part(0.6))
        self.play(FadeOut(words), run_time=0.4)

    def bullets(self, heading: str, steps: list):
        """A heading, then one point per Step (its `show`, else its narration), each as it is said."""
        head = label(heading, scale=0.8, color=ACCENT, bold=True).to_edge(UP, buff=0.8)
        self.play(FadeIn(head), run_time=0.5)
        rows = VGroup()
        for step in steps:
            row = Paragraph(*wrap("•  " + (step.show or step.narration), 48), font=FONT, color=INK, line_spacing=0.8).scale(0.5)
            if len(rows):
                row.next_to(rows, DOWN, buff=0.4, aligned_edge=LEFT)
            else:
                row.next_to(head, DOWN, buff=0.7).to_edge(LEFT, buff=1.4)
            with self.narrate(step.narration) as t:
                self.play(FadeIn(row, shift=RIGHT * 0.3), run_time=t.part(0.25))
            rows.add(row)
        self.play(FadeOut(VGroup(head, rows)), run_time=0.5)

    def compare(self, left: str, right: str, steps: list):
        """Two columns, left (red) against right (green). Each Step adds its `show` to the column
        named by lines=(1,) for left or lines=(2,) for right, as it is said."""
        heads = VGroup(label(left, scale=0.7, color=BAD, bold=True), label(right, scale=0.7, color=GOOD, bold=True))
        heads[0].move_to(LEFT * 3.4 + UP * 2.7)
        heads[1].move_to(RIGHT * 3.4 + UP * 2.7)
        divider = Line(UP * 3, DOWN * 3, color=MUTED, stroke_width=2)
        self.play(FadeIn(heads), FadeIn(divider), run_time=0.6)
        columns = [VGroup(), VGroup()]
        for step in steps:
            side = 1 if 2 in step.lines else 0
            row = Paragraph(*wrap(step.show or step.narration, 24), font=FONT, color=INK, line_spacing=0.8).scale(0.56)
            anchor = columns[side] if len(columns[side]) else heads[side]
            row.next_to(anchor, DOWN, buff=0.45)
            with self.narrate(step.narration) as t:
                self.play(FadeIn(row, shift=DOWN * 0.15), run_time=t.part(0.25))
            columns[side].add(row)
        self.play(FadeOut(VGroup(heads, divider, *columns)), run_time=0.5)

    def quote(self, text: str, source: str = "", narration: str = ""):
        """A quotation, with who said it."""
        words = Paragraph(*wrap("“" + text + "”", 34), font=FONT, color=INK, alignment="center", line_spacing=0.9).scale(0.7)
        group = VGroup(words)
        if source:
            group.add(label("— " + source, scale=0.45, color=MUTED))
        group.arrange(DOWN, buff=0.4)
        with self.narrate(narration or text) as t:
            self.play(FadeIn(words), run_time=t.part(0.3))
            if source:
                self.play(FadeIn(group[1]), run_time=t.part(0.1))
        self.play(FadeOut(group), run_time=0.4)

    def number(self, value: str, caption: str, narration: str = ""):
        """One big number and what it measures."""
        big = label(value, scale=2.2, color=WARM, bold=True)
        under = label(caption, scale=0.6, color=MUTED).next_to(big, DOWN, buff=0.4)
        with self.narrate(narration or f"{value} {caption}") as t:
            self.play(Write(big), run_time=t.part(0.35))
            self.play(FadeIn(under), run_time=t.part(0.15))
        self.play(FadeOut(VGroup(big, under)), run_time=0.4)
