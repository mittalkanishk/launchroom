// Ready-made shots that look like a product's screens: a chat room, a terminal, a code editor.

import { stagger, steps } from "animejs";

const AGENT_COLORS = ["var(--accent)", "var(--accent2)", "var(--accent3)"];

// win is a glass window with a title bar, returning its body.
function win(ctx, width, height, title) {
  const frame = ctx.el("div", "glass");
  frame.style.cssText = `width:${width}px; height:${height}px; display:grid; grid-template-rows:84px 1fr; overflow:hidden;`;
  const bar = ctx.el("div", "", undefined, frame);
  bar.style.cssText = "display:flex; align-items:center; gap:14px; padding:0 30px; border-bottom:1px solid var(--glass-edge);";
  ["#ff5f57", "#febc2e", "#28c840"].forEach((c) => (ctx.el("span", "", undefined, bar).style.cssText = `width:16px;height:16px;border-radius:50%;background:${c};opacity:.85`));
  ctx.el("span", "", title, bar).style.cssText = "margin-left:18px; font-size:32px; font-weight:650; color:var(--muted);";
  const body = ctx.el("div", "", undefined, frame);
  body.style.cssText = "position:relative; overflow:hidden;";
  ctx.add(frame, { opacity: [0, 1], translateY: [70, 0], rotateX: [18, 0], scale: [0.94, 1], duration: 1000, ease: "outExpo" }, ctx.start + 100);
  return body;
}

// chat: a team chat room. Each step's `show` is "name: message". People (shot.people, default you)
// sit on the right; anyone else is an agent on the left, typing before it answers. "approval: text"
// is an approval prompt whose Allow button gets pressed; "joined: name" announces a member.
export function chat(ctx) {
  const { shot } = ctx;
  const people = (shot.people || ["you"]).map((p) => p.toLowerCase());
  const body = win(ctx, 1500, 920, `# ${shot.room || "general"}`);
  const list = ctx.el("div", "", undefined, body);
  list.style.cssText = "position:absolute; left:0; right:0; top:0; padding:40px 48px; display:grid; gap:32px;";
  const agents = [];
  const items = ctx.steps.map((step) => {
    const [who, ...rest] = (step.show || step.say).split(": ");
    const text = rest.join(": ");
    const kind = who.toLowerCase();
    if (kind === "approval") return approval(ctx, list, text);
    if (kind === "joined") {
      const note = ctx.el("div", "", `${text} joined the room`, list);
      note.style.cssText = "justify-self:center; font-size:24px; color:var(--muted); padding:8px 20px; border-radius:20px; background:var(--glass);";
      return { node: note };
    }
    const person = people.includes(kind);
    if (!person && !agents.includes(kind)) agents.push(kind);
    const color = person ? "var(--ink)" : AGENT_COLORS[agents.indexOf(kind) % AGENT_COLORS.length];
    const row = ctx.el("div", "", undefined, list);
    row.style.cssText = `display:flex; gap:18px; align-items:flex-end; ${person ? "flex-direction:row-reverse;" : ""}`;
    const avatar = ctx.el("div", "", person ? who[0].toUpperCase() : "✦", row);
    avatar.style.cssText = `flex:none; width:72px; height:72px; border-radius:22px; display:grid; place-items:center; font-weight:800; font-size:34px; color:var(--bg); background:${color};`;
    const bubble = ctx.el("div", "", undefined, row);
    bubble.style.cssText = `max-width:1080px; padding:24px 34px; border-radius:30px; font-size:40px; line-height:1.35; ${person ? "background:linear-gradient(135deg,var(--accent),var(--accent2)); color:var(--bg);" : "background:var(--glass); border:1px solid var(--glass-edge);"}`;
    const name = ctx.el("div", "", person ? who : `@${who} · agent`, bubble);
    name.style.cssText = `font-size:26px; font-weight:700; margin-bottom:6px; opacity:.75; ${person ? "" : `color:${color}`}`;
    ctx.el("div", "", text, bubble);
    let typing;
    if (!person) {
      typing = ctx.el("div", "", undefined, row);
      typing.style.cssText = "position:absolute; display:flex; gap:10px; padding:28px 32px; border-radius:30px; background:var(--glass); margin-left:90px;";
      typing.dots = [0, 1, 2].map(() => {
        const d = ctx.el("span", "", undefined, typing);
        d.style.cssText = "width:15px; height:15px; border-radius:50%; background:var(--muted);";
        return d;
      });
    }
    return { node: row, bubble, typing, person };
  });

  // Every message starts hidden; the list slides up so the newest one is always in view.
  const visible = 920 - 84 - 60;
  items.forEach((item, s) => {
    item.node.style.opacity = 0;
    const at = ctx.at(s);
    let show = at;
    if (item.typing) {
      const think = Math.min(900, ctx.length(s) * 0.3);
      ctx.add(item.typing, { opacity: [0, 1], duration: 150 }, at);
      ctx.add(item.typing.dots, { translateY: [0, -10, 0], duration: 450, delay: stagger(120), loop: Math.floor(think / 600), ease: "inOutSine" }, at);
      ctx.add(item.typing, { opacity: 0, duration: 120 }, at + think);
      show = at + think;
      item.node.style.opacity = 1;
      item.bubble.style.opacity = 0;
      ctx.add(item.node.firstChild, { opacity: [0, 1], scale: [0.6, 1], duration: 300, ease: "outBack" }, at);
      ctx.add(item.bubble, { opacity: [0, 1], translateY: [24, 0], scale: [0.92, 1], duration: 500, ease: "outBack(1.4)" }, show);
    } else {
      ctx.add(item.node, { opacity: [0, 1], translateY: [40, 0], scale: [0.95, 1], duration: 550, ease: "outBack(1.3)" }, show);
    }
    if (item.after) item.after(show);
    const bottom = item.node.offsetTop + item.node.offsetHeight + 34;
    if (bottom > visible) ctx.add(list, { translateY: -(bottom - visible), duration: 600, ease: "outExpo" }, show);
  });
}

function approval(ctx, list, text) {
  const card = ctx.el("div", "", undefined, list);
  card.style.cssText = "justify-self:center; width:1080px; padding:36px 40px; border-radius:24px; border:2px solid var(--accent); background:color-mix(in srgb, var(--accent) 10%, transparent); display:grid; gap:20px;";
  ctx.el("div", "", "⚑  Approval needed", card).style.cssText = "font-size:28px; font-weight:800; letter-spacing:.12em; text-transform:uppercase; color:var(--accent);";
  ctx.el("div", "", text, card).style.cssText = "font-size:40px; line-height:1.3;";
  const row = ctx.el("div", "", undefined, card);
  row.style.cssText = "display:flex; gap:18px;";
  const allow = ctx.el("div", "", "Allow once", row);
  const decline = ctx.el("div", "", "Decline", row);
  for (const b of [allow, decline]) b.style.cssText = "padding:16px 40px; border-radius:18px; font-size:34px; font-weight:700; border:1px solid var(--glass-edge); background:var(--glass);";
  const done = ctx.el("div", "", "✓ Approved by a person in the room", card);
  done.style.cssText = "font-size:30px; font-weight:700; color:var(--accent); opacity:0;";
  return {
    node: card,
    after: (at) => {
      const press = at + 1300;
      ctx.add(allow, { scale: [1, 0.92, 1.05, 1], backgroundColor: ["rgba(0,0,0,0)", "#28c840"], color: ["#fff", "#06210c"], duration: 600, ease: "outQuad" }, press);
      ctx.add(decline, { opacity: [1, 0.35], duration: 400 }, press);
      ctx.add(done, { opacity: [0, 1], translateY: [10, 0], duration: 500, ease: "outExpo" }, press + 350);
    },
  };
}

// terminal: each step's `show` starting "$ " is typed as a command; other lines are printed output.
export function terminal(ctx) {
  const rows = ctx.steps.length;
  const body = win(ctx, 1500, Math.min(820, 230 + rows * 120), ctx.shot.title || "terminal");
  const list = ctx.el("div", "mono", undefined, body);
  list.style.cssText = "padding:44px 54px; display:grid; gap:26px; font-size:44px; line-height:1.4; align-content:start;";
  ctx.steps.forEach((step, s) => {
    const text = step.show || step.say;
    const command = text.startsWith("$ ");
    const line = ctx.el("div", "", undefined, list);
    line.style.cssText = `white-space:pre; overflow:hidden; color:${command ? "var(--accent)" : "var(--muted)"};`;
    line.textContent = text;
    line.style.opacity = 0;
    if (command) {
      const typing = Math.min(1600, 50 * text.length);
      ctx.add(line, { opacity: [0, 1], duration: 10 }, ctx.at(s));
      ctx.add(line, { width: ["0ch", `${text.length}ch`], duration: typing, ease: steps(text.length) }, ctx.at(s));
    } else {
      ctx.add(line, { opacity: [0, 1], translateY: [10, 0], duration: 400, ease: "outQuad" }, ctx.at(s, 200));
    }
  });
}

// code: an editor showing shot.code; each step with `lines` (1-based) moves the highlight to them
// and dims the rest.
export function code(ctx) {
  const source = String(ctx.shot.code || "").replace(/^\n+|\s+$/g, "");
  const lineCount = String(ctx.shot.code || "").replace(/^\n+|\s+$/g, "").split("\n").length;
  const body = win(ctx, 1640, Math.min(960, 150 + lineCount * 54), ctx.shot.title || ctx.shot.file || "code");
  const panel = ctx.el("div", "mono", undefined, body);
  panel.style.cssText = "position:relative; padding:30px 0; font-size:34px; line-height:54px;";
  const bar = ctx.el("div", "", undefined, panel);
  bar.style.cssText = "position:absolute; left:0; right:0; top:30px; height:54px; background:color-mix(in srgb, var(--accent) 16%, transparent); border-left:5px solid var(--accent); opacity:0;";
  const rows = source.split("\n").map((text, i) => {
    const row = ctx.el("div", "", undefined, panel);
    row.style.cssText = "position:relative; display:flex; white-space:pre; padding-right:40px;";
    ctx.el("span", "", String(i + 1).padStart(3, " "), row).style.cssText = "width:90px; text-align:right; padding-right:30px; color:var(--muted); opacity:.5;";
    const line = ctx.el("span", "", undefined, row);
    line.innerHTML = highlight(text);
    return row;
  });
  ctx.add(rows, { opacity: [0, 1], translateX: [-20, 0], duration: 400, delay: stagger(25), ease: "outQuad" }, ctx.start + 300);
  ctx.steps.forEach((step, s) => {
    if (!step.lines || step.lines.length === 0) return;
    const first = Math.min(...step.lines) - 1;
    const last = Math.max(...step.lines) - 1;
    ctx.add(bar, { opacity: 1, top: 30 + first * 54, height: (last - first + 1) * 54, duration: 600, ease: "outExpo" }, ctx.at(s));
    rows.forEach((row, i) => ctx.add(row, { opacity: step.lines.includes(i + 1) ? 1 : 0.38, duration: 500 }, ctx.at(s)));
  });
}

// highlight colours keywords, strings, numbers and comments; enough for Go, JavaScript, Python and
// shell on a slide.
const KEYWORDS = new Set("func return if else for range var const type struct interface package import go defer select case switch break continue nil true false let function async await class def from export default new this None True False in not and or".split(" "));
function highlight(line) {
  const esc = (s) => s.replace(/&/g, "&amp;").replace(/</g, "&lt;");
  const comment = line.search(/\/\/|#(?!!)/);
  const codePart = comment >= 0 ? line.slice(0, comment) : line;
  const html = codePart.replace(/("[^"]*"|'[^']*'|`[^`]*`)|\b(\d+(?:\.\d+)?)\b|\b([A-Za-z_]\w*)\b|([^"'`\w]+)/g, (m, str, num, word, other) => {
    if (str) return `<span style="color:var(--accent3)">${esc(str)}</span>`;
    if (num) return `<span style="color:var(--accent2)">${num}</span>`;
    if (word) return KEYWORDS.has(word) ? `<span style="color:var(--accent);font-weight:700">${word}</span>` : esc(word);
    return esc(other);
  });
  return html + (comment >= 0 ? `<span style="color:var(--muted);font-style:italic">${esc(line.slice(comment))}</span>` : "");
}
