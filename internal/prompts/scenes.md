# Explainer scenes

Each scene is one file in `scenes/`, numbered with no gaps: `01-hook.js`, `02-the-loop.js`,
`03-code.py`... One scene per section of the script, 20 to 60 seconds each. The narration lives in
the scene: every step speaks its sentence with a free voice, is subtitled, and lasts exactly as long
as the sentence, so the animation lands on the words.

There are two engines, chosen by the file's extension, and a video can use both:

- **`.js`: stage (anime.js).** Motion graphics: kinetic type, glass cards, light flowing between
  nodes, product screens, style packs. For ideas, products, people, processes and numbers.
- **`.py`: Manim.** Precise drawing in the style of 3Blue1Brown. For code explained line by line,
  formulas, graphs and exact diagrams.

Use the same `style` in every `.js` scene of a video.

## Stage scenes (.js)

```js
import { scene } from "stage";

export default scene({
  style: "aurora",
  shots: [
    { type: "title", kicker: "Part 1", title: "An agent is a *loop*",
      steps: ["Claude Code is a loop.", "Here it is, in a few lines of Go."] },
    { type: "flow", nodes: [{ label: "You", icon: "💬" }, { label: "Model", icon: "✦" }, { label: "Tool", icon: "⚙" }],
      steps: [
        { say: "You ask for something.", path: [1, 2] },
        { say: "The model asks for a tool, and the loop runs it.", path: [2, 3] },
        { say: "The result goes back, and it all repeats.", path: [3, 2] },
      ] },
  ],
});
```

A scene is a list of shots shown one after another, each with `steps`. A step is a sentence
(`"..."`), or `{ say, show, ... }`: `say` is spoken and subtitled; `show` is the text the step puts on
screen where the shot shows text (otherwise `say` is shown). In any shown text, `*words*` get the
accent gradient. Shots: 1 to 6 steps each; a scene: 1 to 5 shots.

**Styles:** `aurora` (deep navy, cyan and violet glow), `ember` (dark, warm orange and gold),
`paper` (light, editorial ink and blue).

| Shot | Fields | Shows |
| --- | --- | --- |
| `title` | `title`, `kicker`, `subtitle` | A big title rising word by word, an eyebrow above, a drawn underline; the subtitle comes in with the second step. |
| `kinetic` | steps' `show` | Big statements arriving word by word; earlier lines step back and dim. Best for the key ideas. |
| `flow` | `nodes` (2 to 4: `"label"` or `{ label, icon }`); a step's `path: [a, b]` or `light: n` | Glass nodes joined by drawn arcs; `path` streams light from node a to node b, `light` pulses a node. |
| `cards` | `heading`; a step's `show: "Title: text"` (or `title`, `text`, `icon`) | One card per step rising into a row, with a light sweep. Up to 4. |
| `stat` | `value` ("6", "$0", "3x"), `caption` | A number counting up inside a drawn ring; later steps add a line each. |
| `compare` | `left`, `right`; a step's `side: "left"` or `"right"` | Two columns; each step adds its `show` to one side (✕ left, ✓ right). |
| `chat` | `room`, `people` (default `["you"]`); a step's `show: "name: message"` | A team chat window. People sit right; anyone else is an agent that types before answering. `"approval: text"` shows an approval prompt whose Allow button gets pressed; `"joined: name"` announces a member. |
| `terminal` | `title`; a step's `show` | A terminal: `"$ command"` is typed, anything else printed as output. |
| `code` | `code`, `file`; a step's `lines: [3, 4]` | An editor with simple highlighting; each step moves a highlight bar to its lines and dims the rest. At most about 15 lines of up to 70 characters. |
| `custom` | `build(ctx)` | Your own anime.js animation (below). |

### Your own animation (`custom`)

`build(ctx)` gets: `ctx.layer` (a 1920x1080 box to add elements to), `ctx.el(tag, className, text,
parent)`, `ctx.svg(tag, attrs, parent)`, `ctx.steps` (each with `say`), `ctx.at(i)` (when step i's
sentence starts, in ms; `ctx.at(i, 300)` is 300 ms later), `ctx.length(i)` (how long it is spoken),
`ctx.start` / `ctx.end` (the shot's span), and `ctx.add(targets, params, atMs)`, which adds an
anime.js animation to the scene's timeline. Import helpers from `"animejs"` (`stagger`,
`createDrawable`, `createMotionPath`, `splitText`). Never call `animate()` or play anything yourself;
add everything with `ctx.add`, so the renderer can seek it frame by frame.

```js
import { scene } from "stage";
import { stagger } from "animejs";

export default scene({ style: "aurora", shots: [{ type: "custom", steps: ["Three agents, one room."],
  build(ctx) {
    const dots = [0, 1, 2].map((i) => {
      const d = ctx.el("div", "glass");
      d.style.cssText = `position:absolute; left:${660 + i * 220}px; top:460px; width:160px; height:160px; border-radius:50%;`;
      return d;
    });
    ctx.add(dots, { scale: [0, 1], opacity: [0, 1], delay: stagger(150), duration: 700, ease: "outBack" }, ctx.at(0));
  } }] });
```

CSS variables for colour: `var(--ink)`, `var(--muted)`, `var(--accent)`, `var(--accent2)`,
`var(--accent3)`, `var(--glass)`; classes `glass`, `grad` (gradient text), `eyebrow`, `mono`.

## Manim scenes (.py)

Each defines `class Shot(Explainer)` with `from launchroom import *`:

```python
from launchroom import *

class Shot(Explainer):
    def construct(self):
        self.title("An agent is just a loop", "Part 1", narration="Claude Code is a loop.")
        self.code(LOOP, heading="engine.Run", steps=[
            Step("It calls the model.", lines=(2,)),
        ])
```

### Manim ready-made shots

Each beat is a `Step(narration, show="", lines=())`: `narration` is spoken and subtitled; `show` is
the text the beat puts on screen where the shot shows text; `lines` is what the beat points at,
counting from 1.

| Call | Shows |
| --- | --- |
| `self.title(title, subtitle="", narration="")` | A title card. |
| `self.statement(text, narration="")` | One big sentence: the idea of the moment. |
| `self.bullets(heading, steps)` | A heading, then one point per step (its `show`, else its narration). |
| `self.compare(left, right, steps)` | Two columns, red against green; `lines=(1,)` adds to the left, `(2,)` to the right. |
| `self.quote(text, source="", narration="")` | A quotation and who said it. |
| `self.number(value, caption, narration="")` | One big number and what it measures. |
| `self.code(source, steps, language="go", heading="")` | Exact, highlighted code; each step boxes its `lines`. At most about 22 lines; cut longer code to the part you talk about. |
| `self.flow(nodes, steps, loop=False)` | Boxes joined by arrows, left to right (and back to the start with `loop=True`); each step lights up the boxes in its `lines`. 2 to 5 boxes. |
| `self.terminal(steps, heading="")` | A terminal; a step's `show` starting "$ " is typed as a command, any other is printed as output. |
| `self.chat(steps)` | Message bubbles; each `show` is "user: ...", "model: ..." or "tool: ...". |
| `self.wait(seconds)` | A silent beat. |

### Your own Manim animation

Anything Manim Community 0.22 can draw works. Put it inside a narrate block so it is timed to the
sentence; the block holds the frame until the sentence ends.

```python
with self.narrate("The result goes back to the model.") as t:
    self.play(GrowArrow(arrow), run_time=t.part(0.4))   # 40% of the sentence
```

Colours: `INK` (text), `MUTED`, `ACCENT`, `WARM` (highlight), `GOOD`, `BAD`, `BACKGROUND`. Helpers:
`label(text, scale=0.55, color=INK, mono=False, bold=False)` and `box(text)`. Clear the frame at the
end of your own animation (`self.play(FadeOut(group))`), as the ready-made shots do.

## Rules for every scene

- One scene per section of the script, 20 to 60 seconds each. Sentences of 6 to 20 words read best.
- Vary the shots: don't use the same shot type twice in a row.
- Show what is said: every sentence changes something on screen.
- Code on screen must be real: copy it from the source, trimmed, never invented.
- Render with `video render <name>`; fix what fails (the error names the line) and render that
  scene again with `--shots N`.
