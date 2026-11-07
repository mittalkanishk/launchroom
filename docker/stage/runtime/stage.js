// stage: explainer scenes animated with anime.js, rendered frame by frame by render.mjs.
//
// A scene is `export default scene({ style, shots })`. Each shot is a ready-made type (shots.js)
// with steps; each step is one spoken sentence. The narration is voiced first, so every step's start
// and length are known, and every shot animates to them on one paused anime.js timeline that the
// host seeks to each frame's time. Nothing here runs on the clock, so every render is identical.
//
// The host calls window.__stage: load() → the sentences to voice; build(seconds) → the timing plan;
// seek(ms) before capturing each frame.

import { createTimeline, utils } from "animejs";
import { SHOTS } from "/shots.js";

export function scene(spec) {
  return spec;
}

// Timing, in milliseconds.
const ENTER = 650;   // a shot's entrance; its first sentence starts partway in
const LEAD = 350;    // from a shot's start to its first sentence
const GAP = 280;     // silence between sentences
const HOLD = 450;    // after a shot's last sentence, before it leaves
const EXIT = 450;    // a shot's exit
const HANDOVER = 150; // the next shot starts entering this long before the exit ends
export const timing = { ENTER, LEAD, GAP, HOLD, EXIT };

let spec;
let shots = [];
let tl;

// steps turns a shot's steps into objects: a plain string is a sentence with nothing else.
function steps(shot) {
  return (shot.steps || []).map((s) => (typeof s === "string" ? { say: s } : { ...s }));
}

async function load() {
  spec = (await import("/scene.js")).default;
  if (!spec || !Array.isArray(spec.shots) || spec.shots.length === 0) {
    throw new Error("scene.js must `export default scene({ shots: [...] })` with at least one shot");
  }
  shots = spec.shots.map((shot, i) => {
    if (!SHOTS[shot.type]) {
      throw new Error(`shot ${i + 1}: unknown type "${shot.type}"; the types are ${Object.keys(SHOTS).join(", ")}`);
    }
    const s = steps(shot);
    if (s.length === 0 || s.some((st) => !st.say)) {
      throw new Error(`shot ${i + 1} (${shot.type}): every shot needs steps, and every step something to say`);
    }
    return { ...shot, steps: s };
  });
  document.getElementById("frame").dataset.style = spec.style || "aurora";
  await document.fonts.ready;
  return shots.flatMap((shot) => shot.steps.map((st) => st.say));
}

// build lays the shots out in time from each sentence's spoken length, builds every animation on one
// paused timeline, and returns the plan the host mixes the sound and subtitles from.
function build(seconds) {
  tl = createTimeline({ autoplay: false });
  const container = document.getElementById("shots");
  const plan = { steps: [], cuts: [], total: 0 };
  let k = 0;
  let start = 0;
  shots.forEach((shot, index) => {
    let cursor = start + LEAD;
    for (const st of shot.steps) {
      st.start = cursor;
      st.end = cursor + seconds[k++] * 1000;
      plan.steps.push({ start: st.start, end: st.end, say: st.say });
      cursor = st.end + GAP;
    }
    shot.start = start;
    shot.end = cursor - GAP + HOLD + EXIT;
    plan.cuts.push(start);

    const layer = document.createElement("div");
    layer.className = `shot shot-${shot.type}`;
    container.appendChild(layer);
    const ctx = context(layer, shot, index);
    SHOTS[shot.type](ctx);
    enterAndExit(layer, shot, index);
    moveBackdrop(index, start);

    start = shot.end - HANDOVER;
  });
  plan.total = shots[shots.length - 1].end;
  driftBackdrop(plan.total);
  tl.seek(0);
  return plan;
}

// context is what a shot builds with: its layer (a 1920x1080 box), its steps with their times, the
// timeline, and small helpers. Shots add animations at absolute times, e.g. ctx.at(0).
function context(layer, shot, index) {
  return {
    layer, shot, index, tl,
    steps: shot.steps,
    at: (i, offset = 0) => shot.steps[i].start + offset,
    length: (i) => shot.steps[i].end - shot.steps[i].start,
    start: shot.start,
    end: shot.end - EXIT,
    add: (targets, params, at) => {
      applyFrom(targets, params);
      return tl.add(targets, params, Math.max(0, at));
    },
    el: (tag, className, text, parent = layer) => {
      const node = document.createElement(tag);
      if (className) node.className = className;
      if (text !== undefined) node.textContent = text;
      parent.appendChild(node);
      return node;
    },
    svg: (tag, attrs, parent) => {
      const node = document.createElementNS("http://www.w3.org/2000/svg", tag);
      for (const [k, v] of Object.entries(attrs || {})) node.setAttribute(k, v);
      if (parent) parent.appendChild(node);
      return node;
    },
  };
}

// anime.js applies a tween's "from" value only when that tween starts, so anything entering later
// than its shot (a staggered card, a word, a line of code) would first show at rest, then vanish
// and enter: a visible double load. applyFrom sets every [from, to] value at build time, once per
// element and property, so an element's first tween decides how it starts.
const initialised = new WeakMap();
function applyFrom(targets, params) {
  const list = typeof targets === "string" ? document.querySelectorAll(targets) : targets?.[Symbol.iterator] ? targets : [targets];
  for (const target of list) {
    if (!target || typeof target !== "object") continue;
    let done = initialised.get(target);
    if (!done) initialised.set(target, (done = new Set()));
    const from = {};
    for (const [key, value] of Object.entries(params)) {
      if (!Array.isArray(value) || value.length !== 2 || typeof value[0] === "object" || done.has(key)) continue;
      from[key] = value[0];
      done.add(key);
    }
    if (Object.keys(from).length) utils.set(target, from);
  }
}

// Each shot comes in rising out of a slight blur of depth and leaves upward, alternating direction
// so consecutive shots don't move the same way.
function enterAndExit(layer, shot, index) {
  const from = index % 2 ? -1 : 1;
  tl.add(layer, { opacity: [0, 1], translateY: [40 * from, 0], scale: [0.965, 1], duration: ENTER, ease: "outExpo" }, shot.start);
  tl.add(layer, { opacity: [1, 0], translateY: [0, -50 * from], scale: [1, 1.03], duration: EXIT, ease: "inQuad" }, shot.end - EXIT);
}

// The backdrop's orbs glide to a new arrangement at every shot, like a camera moving between sets.
const SETS = [
  [[0, 0], [0, 0], [0, 0]],
  [[520, -120], [-480, 200], [-620, 140]],
  [[-180, 380], [-900, -150], [420, -60]],
  [[900, 150], [-200, 320], [-900, 80]],
];
function moveBackdrop(index, at) {
  const set = SETS[index % SETS.length];
  ["a", "b", "c"].forEach((name, i) => {
    const [x, y] = set[i];
    tl.add(`.orb-${name}`, { translateX: x, translateY: y, duration: 1800, ease: "inOutSine" }, at);
  });
}

// The grid floor scrolls slowly toward the viewer for the whole video.
function driftBackdrop(total) {
  tl.add("#grid", { backgroundPositionY: ["0px", `${Math.round(total / 1000) * 45}px`], duration: total, ease: "linear" }, 0);
}

function fit() {
  const frame = document.getElementById("frame");
  frame.style.transform = `scale(${window.innerWidth / 1920})`;
}

window.__stage = {
  load,
  build,
  // seek returns nothing: the host would otherwise copy the whole timeline back on every frame.
  seek: (ms) => {
    tl.seek(ms);
  },
};
fit();
