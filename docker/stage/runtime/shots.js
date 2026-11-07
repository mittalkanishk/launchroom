// The ready-made shots. Each takes the shot's context (stage.js) and adds its animations to the
// timeline at its steps' times: ctx.at(i) is when step i's sentence starts, ctx.length(i) how long
// it is spoken. Layout is in the 1920x1080 frame. Text with *stars* is emphasised.

import { createDrawable, createMotionPath, stagger } from "animejs";
import { chat, terminal, code } from "/screens.js";

// words appends text to parent as word spans; *marked* words (one or several) get the accent
// gradient. Punctuation stays with its word.
export function words(ctx, parent, text) {
  const out = [];
  let marked = false;
  text.split(/\s+/).filter(Boolean).forEach((token, i) => {
    const opens = token.startsWith("*");
    const closes = /\*[^\w]*$/.test(token) && (token.length > 1 || !opens);
    if (opens) marked = true;
    const span = ctx.el("span", marked ? "word grad" : "word", token.replace(/\*/g, ""), parent);
    if (closes) marked = false;
    if (i > 0) parent.insertBefore(document.createTextNode(" "), span);
    out.push(span);
  });
  return out;
}

function title(ctx) {
  const { shot } = ctx;
  const box = ctx.el("div", "");
  box.style.cssText = "text-align:center; display:grid; gap:34px; justify-items:center;";
  if (shot.kicker) {
    const kicker = ctx.el("div", "eyebrow", shot.kicker, box);
    ctx.add(kicker, { opacity: [0, 1], letterSpacing: [".6em", ".32em"], duration: 900, ease: "outExpo" }, ctx.start + 150);
  }
  const heading = ctx.el("h1", "", undefined, box);
  heading.style.cssText = "font-size:132px; font-weight:800; letter-spacing:-.035em; line-height:1.08; max-width:1640px; perspective:900px;";
  const titleText = shot.title.includes("*") ? shot.title : `*${shot.title}*`;
  ctx.add(words(ctx, heading, titleText), { opacity: [0, 1], translateY: [110, 0], rotateX: [-80, 0], duration: 1200, delay: stagger(110), ease: "outExpo" }, ctx.start + 250);
  const bar = ctx.el("div", "", undefined, box);
  bar.style.cssText = "height:6px; border-radius:6px; width:560px; background:linear-gradient(90deg,var(--accent),var(--accent2));";
  ctx.add(bar, { scaleX: [0, 1], opacity: [0, 1], duration: 1000, ease: "inOutExpo" }, ctx.start + 700);
  if (shot.subtitle) {
    const sub = ctx.el("p", "", undefined, box);
    sub.style.cssText = "font-size:44px; color:var(--muted); font-weight:400; max-width:1300px;";
    const w = words(ctx, sub, shot.subtitle);
    ctx.add(w, { opacity: [0, 1], translateY: [24, 0], duration: 700, delay: stagger(45), ease: "outCubic" }, ctx.at(Math.min(1, ctx.steps.length - 1), 200));
  }
}

// kinetic: each step's line (its `show`, else what is said) arrives word by word; earlier lines step
// back and dim, so the newest idea is always the brightest thing on screen.
function kinetic(ctx) {
  const stack = ctx.el("div", "");
  stack.style.cssText = "display:grid; gap:26px; width:1560px; text-align:center;";
  const lines = [];
  ctx.steps.forEach((step, i) => {
    const line = ctx.el("div", "", undefined, stack);
    line.style.cssText = "font-size:82px; font-weight:750; letter-spacing:-.025em; line-height:1.1;";
    const w = words(ctx, line, step.show || step.say);
    line.style.opacity = 0;
    ctx.add(line, { opacity: [0, 1], duration: 10 }, ctx.at(i));
    ctx.add(w, { opacity: [0, 1], translateY: [70, 0], scale: [0.9, 1], duration: 650, delay: stagger(70), ease: "outBack(1.4)" }, ctx.at(i));
    lines.forEach((prev, age) => {
      const dim = Math.max(0.12, 0.42 - 0.15 * (lines.length - 1 - age));
      ctx.add(prev, { opacity: dim, scale: 0.86, duration: 600, ease: "outExpo" }, ctx.at(i));
    });
    lines.push(line);
  });
}

// flow: nodes in a row joined by drawn arcs. A step with `path: [a, b]` (1-based) sends a stream of
// light from node a to node b; `light: n` makes node n pulse.
function flow(ctx) {
  const nodes = ctx.shot.nodes || [];
  const n = nodes.length;
  const xs = nodes.map((_, i) => 960 + (i - (n - 1) / 2) * Math.min(520, 1500 / Math.max(1, n - 1)));
  const y = 520;
  const svg = ctx.svg("svg", { width: 1920, height: 1080, viewBox: "0 0 1920 1080", style: "position:absolute;inset:0" }, ctx.layer);
  const paths = {};
  for (let i = 0; i < n; i++) {
    for (let j = 0; j < n; j++) {
      if (i === j) continue;
      // Forward paths arc over the cards (from their tops), backward ones under them.
      const dir = j > i ? -1 : 1;
      const edge = y + dir * 110;
      const bend = y + dir * (190 + Math.abs(xs[j] - xs[i]) * 0.18);
      const d = `M ${xs[i]} ${edge} C ${xs[i]} ${bend}, ${xs[j]} ${bend}, ${xs[j]} ${edge}`;
      const shown = j === i + 1;
      paths[`${i}-${j}`] = ctx.svg("path", { d, fill: "none", stroke: shown ? "var(--glass-edge)" : "none", "stroke-width": 3, "stroke-dasharray": "2 12", "stroke-linecap": "round" }, svg);
    }
  }
  ctx.add(createDrawable(Object.values(paths)), { draw: ["0 0", "0 1"], duration: 1200, ease: "inOutQuad" }, ctx.start + 500);

  const cards = nodes.map((node, i) => {
    const card = ctx.el("div", "glass");
    const label = typeof node === "string" ? node : node.label;
    const icon = typeof node === "string" ? "" : node.icon || "";
    card.style.cssText = `position:absolute; left:${xs[i] - 150}px; top:${y - 105}px; width:300px; height:210px; display:grid; place-items:center; align-content:center; gap:10px; text-align:center;`;
    if (icon) ctx.el("div", "", icon, card).style.cssText = "font-size:64px; line-height:1;";
    ctx.el("div", "", label, card).style.cssText = "font-size:40px; font-weight:700;";
    return card;
  });
  ctx.add(cards, { opacity: [0, 1], translateY: [60, 0], scale: [0.8, 1], duration: 900, delay: stagger(120), ease: "outBack(1.6)" }, ctx.start + 150);

  ctx.steps.forEach((step, s) => {
    if (step.path) {
      const [a, b] = step.path.map((v) => v - 1);
      const path = paths[`${a}-${b}`];
      if (!path) return;
      path.setAttribute("stroke", "var(--glass-edge)");
      const dots = Array.from({ length: 7 }, (_, k) => {
        const dot = ctx.svg("circle", { r: k === 0 ? 11 : 7 - k * 0.6, fill: k % 2 ? "var(--accent2)" : "var(--accent)", opacity: 0 }, svg);
        dot.style.filter = "drop-shadow(0 0 10px var(--accent))";
        return dot;
      });
      const travel = Math.min(1700, ctx.length(s) * 0.7);
      dots.forEach((dot, k) => {
        ctx.add(dot, { ...createMotionPath(path), opacity: [{ to: 1, duration: 120 }, { to: 1, duration: travel - 300 }, { to: 0, duration: 180 }], duration: travel, ease: "inOutSine" }, ctx.at(s, k * 70));
      });
      ctx.add(cards[b], { scale: [1, 1.12, 1], borderColor: ["var(--glass-edge)", "var(--accent)"], duration: 700, ease: "outElastic(1, .6)" }, ctx.at(s, travel - 100));
    }
    if (step.light) {
      ctx.add(cards[step.light - 1], { scale: [1, 1.12, 1], borderColor: ["var(--glass-edge)", "var(--accent)"], duration: 800, ease: "outElastic(1, .6)" }, ctx.at(s));
    }
  });
}

// cards: a heading, then one card per step (its `show` as "Title: text", or `title` and `text`),
// rising into a row as each is said.
function cards(ctx) {
  const { shot } = ctx;
  const box = ctx.el("div", "");
  box.style.cssText = "display:grid; gap:56px; justify-items:center; width:1700px;";
  if (shot.heading) {
    const h = ctx.el("h2", "", undefined, box);
    h.style.cssText = "font-size:68px; font-weight:780; letter-spacing:-.02em; text-align:center;";
    ctx.add(words(ctx, h, shot.heading), { opacity: [0, 1], translateY: [40, 0], duration: 700, delay: stagger(60), ease: "outExpo" }, ctx.start + 200);
  }
  const row = ctx.el("div", "", undefined, box);
  const cols = Math.min(4, ctx.steps.length);
  row.style.cssText = `display:grid; grid-template-columns:repeat(${cols}, 1fr); gap:32px; width:${Math.min(1700, cols * 420)}px;`;
  ctx.steps.forEach((step, i) => {
    const [t, ...rest] = (step.show || "").split(": ");
    const card = ctx.el("div", "glass", undefined, row);
    card.style.cssText = "padding:52px 48px; display:grid; gap:20px; align-content:start; min-height:340px; position:relative; overflow:hidden;";
    const badge = ctx.el("div", "", step.icon || String(i + 1).padStart(2, "0"), card);
    badge.style.cssText = "font-size:50px; font-weight:800; color:var(--accent);";
    ctx.el("div", "", step.title || t || step.say, card).style.cssText = "font-size:48px; font-weight:750; line-height:1.15;";
    const text = step.text || rest.join(": ");
    if (text) ctx.el("div", "", text, card).style.cssText = "font-size:36px; color:var(--muted); line-height:1.35;";
    const shine = ctx.el("div", "", undefined, card);
    shine.style.cssText = "position:absolute; inset:-50% auto -50% -60%; width:40%; transform:rotate(18deg); background:linear-gradient(90deg,transparent,rgba(255,255,255,.18),transparent);";
    card.style.opacity = 0;
    ctx.add(card, { opacity: [0, 1], translateY: [80, 0], rotateX: [35, 0], duration: 800, ease: "outExpo" }, ctx.at(i));
    ctx.add(shine, { left: ["-60%", "130%"], duration: 1100, ease: "inOutSine" }, ctx.at(i, 250));
  });
}

// stat: one big number counting up, a ring drawing round it, and the caption; later steps add a
// line each under it.
function stat(ctx) {
  const { shot } = ctx;
  const box = ctx.el("div", "");
  box.style.cssText = "display:grid; justify-items:center; gap:30px; position:relative;";
  const ring = ctx.svg("svg", { width: 560, height: 560, viewBox: "0 0 560 560" }, box);
  ring.style.cssText = "position:absolute; top:-90px;";
  const circle = ctx.svg("circle", { cx: 280, cy: 280, r: 255, fill: "none", stroke: "url(#g)", "stroke-width": 10, "stroke-linecap": "round", transform: "rotate(-90 280 280)" }, ring);
  const defs = ctx.svg("defs", {}, ring);
  const grad = ctx.svg("linearGradient", { id: "g" }, defs);
  ctx.svg("stop", { offset: "0", "stop-color": "var(--accent)" }, grad);
  ctx.svg("stop", { offset: "1", "stop-color": "var(--accent2)" }, grad);
  ctx.add(createDrawable(circle), { draw: ["0 0", "0 1"], duration: 1600, ease: "inOutCubic" }, ctx.at(0));

  const value = String(shot.value);
  const target = parseFloat(value.replace(/[^\d.]/g, "")) || 0;
  const [prefix, suffix] = value.split(/[\d.,]+/);
  const big = ctx.el("div", "grad", value, box);
  big.style.cssText = "font-size:230px; font-weight:850; letter-spacing:-.05em; line-height:380px; height:380px; font-variant-numeric:tabular-nums;";
  const counter = { v: 0 };
  ctx.add(counter, { v: [0, target], duration: 1500, ease: "outExpo",
    onRender: () => { big.textContent = `${prefix || ""}${Math.round(counter.v).toLocaleString("en-US")}${suffix || ""}`; } }, ctx.at(0));
  const caption = ctx.el("div", "", shot.caption || "", box);
  caption.style.cssText = "font-size:46px; color:var(--muted); font-weight:500; margin-top:110px;";
  ctx.add(caption, { opacity: [0, 1], translateY: [20, 0], duration: 700, ease: "outCubic" }, ctx.at(0, 500));
  ctx.steps.slice(1).forEach((step, k) => {
    const line = ctx.el("div", "", step.show || step.say, box);
    line.style.cssText = "font-size:36px; font-weight:600;";
    ctx.add(line, { opacity: [0, 1], translateY: [20, 0], duration: 600, ease: "outCubic" }, ctx.at(k + 1));
  });
}

// compare: two columns; each step adds its `show` to the side named by `side` ("left" or "right").
function compare(ctx) {
  const { shot } = ctx;
  const grid = ctx.el("div", "");
  grid.style.cssText = "display:grid; grid-template-columns:1fr 1fr; gap:70px; width:1700px; align-items:stretch;";
  const cols = [shot.left || "Before", shot.right || "After"].map((name, i) => {
    const col = ctx.el("div", "glass", undefined, grid);
    col.style.cssText = `padding:56px; display:grid; gap:30px; min-height:420px; align-content:start; ${i ? "border-color:var(--accent)" : "opacity:.9"}`;
    const h = ctx.el("div", "", name, col);
    h.style.cssText = `font-size:54px; font-weight:800; letter-spacing:-.01em; color:${i ? "var(--accent)" : "var(--muted)"};`;
    ctx.add(col, { opacity: [0, 1], translateX: [i ? 120 : -120, 0], duration: 900, ease: "outExpo" }, ctx.start + 150 + i * 150);
    return col;
  });
  ctx.steps.forEach((step, s) => {
    const side = step.side === "right" ? 1 : 0;
    const item = ctx.el("div", "", (side ? "✓  " : "✕  ") + (step.show || step.say), cols[side]);
    item.style.cssText = `font-size:48px; font-weight:650; line-height:1.25; ${side ? "" : "color:var(--muted);"}`;
    ctx.add(item, { opacity: [0, 1], translateX: [side ? 60 : -60, 0], duration: 650, ease: "outExpo" }, ctx.at(s));
  });
}

// custom: the scene's own build(ctx), for anything the other shots don't cover.
function custom(ctx) {
  if (typeof ctx.shot.build !== "function") throw new Error("a custom shot needs build(ctx)");
  ctx.shot.build(ctx);
}

export const SHOTS = { title, kinetic, flow, cards, stat, compare, chat, terminal, code, custom };
