// A sample stage scene that uses every ready-made shot. Render it in the stage image:
//   node /opt/stage/runtime/render.mjs --scene switch-room.js --out out --width 960 --height 540 --fps 15
import { scene } from "stage";

export default scene({
  style: "aurora",
  shots: [
    {
      type: "title", kicker: "Launch room", title: "Agents in the room",
      subtitle: "People and AI agents, getting real work done together",
      steps: [
        "Your AI agent shouldn't live in your terminal.",
        "It should be in the room, with your team.",
      ],
    },
    {
      type: "kinetic",
      steps: [
        { say: "Today, an agent works alone.", show: "Today, an agent works *alone*." },
        { say: "One person, one terminal, one conversation.", show: "One person. One terminal." },
        { say: "Switch puts it in a chat room instead.", show: "Switch puts it in the *room*." },
      ],
    },
    {
      type: "flow",
      nodes: [{ label: "You", icon: "💬" }, { label: "Switch", icon: "⇄" }, { label: "Agent", icon: "✦" }],
      steps: [
        { say: "You mention an agent in the chat.", path: [1, 2] },
        { say: "Switch hands the message to the agent's controller.", path: [2, 3] },
        { say: "And the answer comes back into the room, for everyone.", path: [3, 1] },
      ],
    },
    {
      type: "chat", room: "launch",
      steps: [
        { say: "Ask for something, the way you'd ask a colleague.", show: "you: @builder make a launch page for Quill" },
        { say: "The builder answers in the thread.", show: "builder: Done. Preview: preview-quill.example.dev" },
        { say: "Going live needs a person to press Allow.", show: "approval: builder wants to run site live quill" },
      ],
    },
    {
      type: "cards", heading: "What the room gives you",
      steps: [
        { say: "Anyone on the team can drive the agents.", show: "Shared: anyone can ask, from their phone" },
        { say: "Agents hand work to each other in the open.", show: "Open: agents talk where you can see" },
        { say: "And risky steps wait for a human.", show: "Safe: approvals are buttons in the chat" },
      ],
    },
    {
      type: "terminal", title: "setup",
      steps: [
        { say: "Self-hosting it is one command.", show: "$ mise run switch:up" },
        { say: "Every service starts in Docker.", show: "✓ switch  ✓ agent-controller  ✓ studio  ✓ stage" },
      ],
    },
    {
      type: "code", file: "gate.go",
      code: `
var Steps = []Gated{
    {Command: "site live", Reason: "Going live needs a person."},
    {Command: "video publish", Reason: "Publishing needs a person."},
}

// Check answers "ask" for a gated command.
func Check(in io.Reader, out io.Writer) error {
    return ask(out, step.Reason)
}`,
      steps: [
        { say: "The gate is a few lines of Go.", lines: [1, 2, 3, 4] },
        { say: "It answers ask, and Switch turns that into buttons.", lines: [7, 8, 9] },
      ],
    },
    {
      type: "compare", left: "An agent alone", right: "An agent in the room",
      steps: [
        { say: "Alone, only you see what it does.", show: "Only you see its work", side: "left" },
        { say: "In the room, the whole team does.", show: "The whole team sees it", side: "right" },
        { say: "And a person approves what matters.", show: "People approve what matters", side: "right" },
      ],
    },
    {
      type: "stat", value: "6", caption: "agents, one room",
      steps: ["This video was made by six agents, in one room."],
    },
  ],
});
