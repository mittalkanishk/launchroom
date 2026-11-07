# Launch Room

Explainer videos made in your team chat by a crew of AI agents, with a person deciding what goes
out. Built on [Switch](https://github.com/sandbox-quantum/switch) for the Flip the Switch hackathon.

```
Sam       @producer make a 3-minute explainer about Switch, for developers. Use its docs so it's all true.
producer  switch-in-the-room. Script: https://preview-switch-in-the-room.example.dev/script/
Sam       section 4 is too long, cut it to the approval part. ok otherwise
producer  Done. @director the brief and script are ready.
director  Style aurora. Scenes 1-3, 5-6 on stage (anime.js); scene 4, the approval code, in Manim. Rendering.
director  @reviewer review the scenes of switch-in-the-room
reviewer  Scene 2: the same shot twice in a row. Scene 4: line 7 is cut off. Rest looks good.
director  Fixed both and re-rendered 2 and 4. @editor ready for the cut.
editor    Cut, finished, QC passed. @reviewer review the video. … Preview: https://preview-switch-in-the-room.example.dev
Sam       publish it
editor    ┌ Approval needed: video publish switch-in-the-room ┐  [Allow once] [Decline]
```

## Two engines, one video

Every scene is a file the director writes; the file's extension picks the engine, and the cut
joins them, so one video mixes both without the viewer noticing a seam.

| A section that shows | Engine | File | Good at |
| --- | --- | --- | --- |
| Ideas, products, people, processes, numbers | **stage**: anime.js in headless Chromium | `scenes/NN.js` | Motion graphics: kinetic type, flows, cards, chat and terminal screens |
| Code explained line by line, formulas, graphs, exact diagrams | **Manim** | `scenes/NN.py` | Precise drawing, code walkthroughs |

Both speak their own narration (free Edge voices, timed first so every animation lands on its
sentence) and hand back an MP4, subtitles and a sheet with a frame per sentence. Studio (ffmpeg)
cuts, grades, brings loudness to -14 LUFS and measures. Each engine runs in its own keyless
container, because the scenes are code the agents wrote.

## Why Switch

The crew is five managed agents in one room: `producer` (the script), `director` (style, engine
and shot per scene, written down in `shots.md`), `editor` (cut, finish, QC), `reviewer` (Codex:
the scenes before the cut, the video before publish) and `builder` (sites and decks). Switch gives
them the room, the handovers (`@director`), and the approval card: `video publish` is a gated
command, and only a person pressing Allow lets it run. Every decision is a file in the video's
folder, and every step is a `video` command, so the room can read, change and redo any of it.

```
producer  video new, script preview        director  video scenes, check, render [--shots N]
editor    video cut, finish, qc, publish   reviewer  reads shots.md, scenes/, renders/scenes/*.jpg, renders/qc/
```

## Set it up

Docker, [mise](https://mise.jdx.dev), a Cloudflare domain with a tunnel, a Telegram bot, a Claude
subscription or API key, and an OpenAI key for Codex. Then:

1. `mise run switch:init-env`, fill in `.env` (`.env.example` explains every value).
2. Route `switch.<domain>` → `http://switch:8000` and `*.<domain>` → `http://agent-controller:8080` in the tunnel.
3. `mise run switch:up`, sign in, connect your Telegram group under **Apps**.
4. **Machines → Add machine**, put the code in `.env` as `SWITCH_CONTROLLER_ENROLL_CODE`, `mise run switch:up` again.
5. Create the managed agents on that machine, named `producer`, `director`, `editor`, `reviewer`
   (Codex), `builder`; set each to **Open** addressing; `/invite_agent @producer` and the rest.
6. Develop: `GOWORK=off go test ./...`. Prompts in `internal/prompts/`, engines in `docker/stage/runtime/` and `docker/renderer/launchroom/`.
