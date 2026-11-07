// render.mjs renders one stage scene to a narrated MP4 and its subtitles:
//
//   node render.mjs --scene scene.js --out dir --width 1920 --height 1080 --fps 30
//
// 1. Load the scene in headless Chromium and ask it for its sentences.
// 2. Voice each sentence (Microsoft Edge's free voices) and measure it.
// 3. Have the scene lay out its timeline from those lengths.
// 4. Seek the timeline to every frame's time and capture it, straight into ffmpeg.
// 5. Mix the narration, and a soft whoosh at every change of shot, under the frames.
//
// Writes <out>/scene.mp4 and <out>/scene.srt.

import { execFile, spawn } from "node:child_process";
import { createHash } from "node:crypto";
import { existsSync, mkdirSync, readFileSync, renameSync, writeFileSync } from "node:fs";
import { dirname, extname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { promisify } from "node:util";
import { chromium } from "playwright";

const run = promisify(execFile);
const HERE = dirname(fileURLToPath(import.meta.url));
const ANIME = resolve(HERE, "../node_modules/animejs/dist/bundles/anime.esm.js");
const VOICE = process.env.LAUNCHROOM_VOICE || "en-US-AndrewNeural";
const CACHE = process.env.LAUNCHROOM_VOICE_CACHE || "/tmp/launchroom-voice";
const EDGE_TTS = process.env.EDGE_TTS || "edge-tts";

const args = Object.fromEntries(process.argv.slice(2).reduce((pairs, a, i, all) => (a.startsWith("--") ? [...pairs, [a.slice(2), all[i + 1]]] : pairs), []));
const width = Number(args.width || 1920);
const height = Number(args.height || 1080);
const fps = Number(args.fps || 30);
const out = resolve(args.out || ".");
const scenePath = resolve(args.scene || "scene.js");

async function main() {
  mkdirSync(out, { recursive: true });
  await checkSyntax(scenePath);

  const browser = await chromium.launch({ args: ["--force-color-profile=srgb", "--font-render-hinting=none"] });
  try {
    const page = await browser.newPage({ viewport: { width, height }, deviceScaleFactor: 1 });
    const errors = [];
    page.on("pageerror", (e) => errors.push(e.message));
    await page.route("http://stage.local/**", (route) => serve(route));
    await page.goto("http://stage.local/page.html");
    await page.waitForFunction(() => window.__stage !== undefined);

    const sentences = await page.evaluate(() => window.__stage.load()).catch((e) => fail(e, errors));
    const voiced = [];
    for (const text of sentences) voiced.push(await speak(text));
    const plan = await page.evaluate((s) => window.__stage.build(s), voiced.map((v) => v.seconds)).catch((e) => fail(e, errors));
    if (errors.length) fail(new Error(errors.join("\n")), []);

    const audio = join(out, "narration.wav");
    await mix(plan, voiced, audio);
    await capture(page, plan.total, audio, join(out, "scene.mp4"));
    writeFileSync(join(out, "scene.srt"), srt(plan.steps));
  } finally {
    await browser.close();
  }
}

// serve answers the page's requests from the runtime folder, the scene, and anime.js.
function serve(route) {
  const path = new URL(route.request().url()).pathname;
  const file = path === "/scene.js" ? scenePath : path === "/anime.js" ? ANIME : join(HERE, path.replace(/^\/+/, ""));
  if (!file.startsWith(HERE) && file !== scenePath && file !== ANIME) return route.fulfill({ status: 404 });
  if (!existsSync(file)) return route.fulfill({ status: 404, body: `not found: ${path}` });
  const types = { ".js": "text/javascript", ".mjs": "text/javascript", ".css": "text/css", ".html": "text/html" };
  return route.fulfill({ status: 200, contentType: types[extname(file)] || "application/octet-stream", body: readFileSync(file) });
}

// checkSyntax reports a broken scene with its line number before any browser starts.
async function checkSyntax(path) {
  const copy = join(out, "check.mjs");
  writeFileSync(copy, readFileSync(path));
  try {
    await run("node", ["--check", copy]);
  } catch (e) {
    throw new Error(`scene.js has a syntax error:\n${String(e.stderr).replaceAll(copy, "scene.js").trim()}`);
  }
}

function fail(error, errors) {
  const detail = [...errors, error.message].filter(Boolean).join("\n");
  throw new Error(`the scene failed in the browser:\n${detail.replace(/http:\/\/stage\.local\//g, "")}`);
}

// speak voices one sentence, cached by voice and text, retrying when the service sends no audio.
async function speak(text) {
  mkdirSync(CACHE, { recursive: true });
  const key = createHash("sha256").update(`${VOICE}\n${text}`).digest("hex").slice(0, 24);
  const file = join(CACHE, `${key}.mp3`);
  if (!existsSync(file)) {
    for (let attempt = 1; ; attempt++) {
      try {
        await run(EDGE_TTS, ["--voice", VOICE, "--text", text, "--write-media", `${file}.part`]);
        renameSync(`${file}.part`, file);
        break;
      } catch (e) {
        if (attempt === 5) throw new Error(`text-to-speech failed 5 times for "${text.slice(0, 60)}": ${e.stderr || e.message}`);
        await new Promise((r) => setTimeout(r, 1500 * attempt));
      }
    }
  }
  const { stdout } = await run("ffprobe", ["-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", file]);
  return { file, seconds: parseFloat(stdout) };
}

// mix lays each sentence at its start and a whoosh just before every new shot, as long as the video.
async function mix(plan, voiced, outFile) {
  const whoosh = join(out, "whoosh.wav");
  await run("ffmpeg", ["-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "anoisesrc=color=pink:duration=0.9:amplitude=0.6",
    "-af", "highpass=f=250,lowpass=f=4500,afade=t=in:d=0.5:curve=exp,afade=t=out:st=0.5:d=0.4,volume=0.16", whoosh]);
  const inputs = [];
  const chains = [];
  voiced.forEach((v, i) => {
    inputs.push("-i", v.file);
    const ms = Math.round(plan.steps[i].start);
    chains.push(`[${i}:a]aresample=48000,aformat=channel_layouts=stereo,adelay=${ms}|${ms}[v${i}]`);
  });
  plan.cuts.slice(1).forEach((cut, k) => {
    const i = voiced.length + k;
    inputs.push("-i", whoosh);
    const ms = Math.max(0, Math.round(cut - 300));
    chains.push(`[${i}:a]aresample=48000,aformat=channel_layouts=stereo,adelay=${ms}|${ms}[w${k}]`);
  });
  const labels = chains.map((c) => c.slice(c.lastIndexOf("[")));
  const seconds = (plan.total / 1000).toFixed(3);
  const graph = `${chains.join(";")};${labels.join("")}amix=inputs=${labels.length}:normalize=0,apad,atrim=end=${seconds}[a]`;
  const script = join(out, "mix.txt");
  writeFileSync(script, graph);
  await run("ffmpeg", ["-hide_banner", "-loglevel", "error", "-y", ...inputs, "-filter_complex_script", script, "-map", "[a]", outFile], { maxBuffer: 1 << 24 });
}

// capture seeks to every frame's time, screenshots it, and pipes the frames into ffmpeg with the sound.
async function capture(page, totalMs, audio, outFile) {
  const frames = Math.ceil((totalMs / 1000) * fps);
  const ffmpeg = spawn("ffmpeg", ["-hide_banner", "-loglevel", "error", "-y",
    "-f", "image2pipe", "-c:v", "mjpeg", "-framerate", String(fps), "-i", "-", "-i", audio,
    "-c:v", "libx264", "-preset", "veryfast", "-crf", "18", "-pix_fmt", "yuv420p",
    "-c:a", "aac", "-b:a", "192k", "-shortest", "-movflags", "+faststart", outFile], { stdio: ["pipe", "inherit", "inherit"] });
  const finished = new Promise((ok, no) => ffmpeg.on("close", (code) => (code === 0 ? ok() : no(new Error(`ffmpeg exited ${code}`)))));
  const cdp = await page.context().newCDPSession(page);
  for (let f = 0; f < frames; f++) {
    await page.evaluate((ms) => window.__stage.seek(ms), (f * 1000) / fps);
    const { data } = await cdp.send("Page.captureScreenshot", { format: "jpeg", quality: 92, optimizeForSpeed: true });
    if (!ffmpeg.stdin.write(Buffer.from(data, "base64"))) await new Promise((r) => ffmpeg.stdin.once("drain", r));
  }
  ffmpeg.stdin.end();
  await finished;
}

function srt(steps) {
  const time = (ms) => {
    const t = Math.round(ms);
    const pad = (n, w = 2) => String(n).padStart(w, "0");
    return `${pad(Math.floor(t / 3600000))}:${pad(Math.floor(t / 60000) % 60)}:${pad(Math.floor(t / 1000) % 60)},${pad(t % 1000, 3)}`;
  };
  return steps.map((s, i) => `${i + 1}\n${time(s.start)} --> ${time(s.end)}\n${s.say}\n`).join("\n");
}

main().catch((e) => {
  console.error(e.message);
  process.exit(1);
});
