// Scene 1. The director replaces this with the first section of the script.
// `video scenes` prints every ready-made shot; scenes/NN.py files are animated with Manim instead.
import { scene } from "stage";

export default scene({
  style: "aurora",
  shots: [
    { type: "title", kicker: "Series name", title: "Title of the video", steps: ["The first sentence the viewer hears."] },
  ],
});
