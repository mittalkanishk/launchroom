"""A sample explainer shot that uses every ready-made shot. Render it in the renderer image:
    manim render the_loop.py Shot
"""
from launchroom import *

LOOP = '''
for {
    if reason, over := e.limitReached(ctx); over {
        e.d.Finish.Done(reason, e.d.State.LastReply())
        return
    }

    t, ok := e.callModel(ctx)
    if !ok {
        return
    }

    var keepGoing bool
    if len(t.reply.ToolCalls) > 0 {
        keepGoing = e.handleToolCalls(ctx, t)
    } else {
        keepGoing = e.handleText(ctx, t)
    }
    if !keepGoing {
        return
    }
}
'''


class Shot(Explainer):
    def construct(self):
        self.title("An agent is just a loop", "How Claude Code Works, part 1",
                   narration="Claude Code is a loop. Here it is, in a few lines of Go.")
        self.chat([
            Step("You ask for something.", show="user: Fix the failing test"),
            Step("The model answers with a request to run a tool.", show="model: run_tests()"),
            Step("The loop runs it and hands back the result.", show="tool: FAIL loop_test.go:42"),
        ])
        self.flow(["model", "tool", "result"], loop=True, steps=[
            Step("The loop calls the model.", lines=(1,)),
            Step("If the model asks for a tool, the loop runs it.", lines=(2,)),
            Step("The result goes back to the model, and it all repeats.", lines=(3, 1)),
        ])
        self.code(LOOP, heading="steerloop / engine.Run", steps=[
            Step("First, the loop checks its limits, so it can always stop.", lines=(2, 3, 4, 5)),
            Step("Then it calls the model.", lines=(7,)),
            Step("If the model asked for tools, the loop runs them. Otherwise it reads the answer.", lines=(13, 14, 15, 16, 17)),
        ])
        self.compare("What people think", "What happens", steps=[
            Step("People imagine the model running code.", show="The model runs your code", lines=(1,)),
            Step("In fact it only writes text that asks for it.", show="The model writes a request; the loop runs it", lines=(2,)),
        ])
        self.terminal(heading="your terminal", steps=[
            Step("You can watch it yourself.", show="$ go run ./cmd/steerloop2 'fix the test'"),
            Step("Every line is one turn of the loop.", show="iteration 1 · tool run_tests · 1.2s"),
        ])
        self.number("3", "parts: model, tools, loop", narration="That is the whole agent. Three parts.")
        self.statement("The model never runs anything. The loop does.")
