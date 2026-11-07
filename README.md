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

## Live deployment

Launch Room is deployed on Railway and running now.

- **Dashboard:** https://gateway-production-776f.up.railway.app
- **Judge login:** `judge@switchai.com` / `judge123`

The machine `railway` runs the crew: five Claude Code agents, `producer`, `director`, `editor`,
`reviewer` and `builder`. The room is reachable from Telegram (`@launch_room_bot`) and Slack (the
`launch_room_live` app). Every release (`mise run git-tag-release`) builds the images and deploys
them (`.github/workflows/`), and `docker/docker-compose.yml` runs the same images locally.

![The Railway project: switch, its dashboard (gateway) and Postgres; the agent controller; and the
video back end, studio with its stage and renderer engines](docs/railway-deployment.jpg)

### Connect Telegram

1. In Telegram, message **@BotFather**: `/newbot` gives you a bot token; `/setprivacy` → **Disable**
   lets the bot read every message in a group.
2. Dashboard → **Apps → Register messaging app**: type `telegram`, a display name, the bot token
   and the bot's username → **Register**.
3. Add the bot to your Telegram group as an admin, and link the group to a room from **Apps**.
4. In the group: `/invite_agent @producer`, then `@director`, `@editor`, `@reviewer` and `@builder`.
5. Ask: `@producer make a 60-second explainer about <topic>`.

### Connect Slack

1. Dashboard → **Apps → Register messaging app**: type `slack`.
2. Create the Slack app from Switch's manifest (Socket Mode, its scopes, events and slash commands
   are set for you): signed in as an admin, open `/gateway/collaborations/setup/slack` on the
   dashboard and follow its `create_url`, or paste its `manifest` at https://api.slack.com/apps →
   **Create New App → From a manifest**.
3. In the Slack app: **Basic Information → App-Level Tokens → Generate** with the
   `connections:write` scope (the `xapp-…` token), and **Install App → Install to Workspace** (the
   bot token, `xoxb-…`). The workspace id (`T…`) is in the app's settings URL.
4. Paste the bot token, the app-level token and the workspace id → **Register**.
5. In a channel: `/invite @<your app>`, link the channel to a room from **Apps**, then
   `/invite-agent @producer` and the rest of the crew.

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

The crew is five managed Claude Code agents in one room: `producer` (the script), `director` (style, engine
and shot per scene, written down in `shots.md`), `editor` (cut, finish, QC), `reviewer` (the
scenes before the cut, the video before publish) and `builder` (sites and decks). Switch gives
them the room, the handovers (`@director`), and the approval card: `video publish` is a gated
command, and only a person pressing Allow lets it run. Every decision is a file in the video's
folder, and every step is a `video` command, so the room can read, change and redo any of it.

```
producer  video new, script preview        director  video scenes, check, render [--shots N]
editor    video cut, finish, qc, publish   reviewer  reads shots.md, scenes/, renders/scenes/*.jpg, renders/qc/
```

## Set it up

Docker, [mise](https://mise.jdx.dev), a Cloudflare domain with a tunnel, a Telegram bot, a Claude
subscription or API key, and an OpenAI key for transcribing voice notes. Then:

1. `mise run switch:init-env`, fill in `.env` (`.env.example` explains every value).
2. Route `switch.<domain>` → `http://switch:8000` and `*.<domain>` → `http://agent-controller:8080` in the tunnel.
3. `mise run switch:up`, sign in, connect your Telegram group under **Apps**.
4. **Machines → Add machine**, put the code in `.env` as `SWITCH_CONTROLLER_ENROLL_CODE`, `mise run switch:up` again.
5. Create the managed agents on that machine, named `producer`, `director`, `editor`, `reviewer`,
   `builder`, all Claude Code; set each to **Open** addressing; `/invite_agent @producer` and the rest.
6. Develop: `GOWORK=off go test ./...`. Prompts in `internal/prompts/`, engines in `docker/stage/runtime/` and `docker/renderer/launchroom/`.
