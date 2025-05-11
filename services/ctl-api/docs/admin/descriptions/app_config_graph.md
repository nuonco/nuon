This downloads a graphviz file that you can render into a graph image.

Run this command once you hit this endpoint and you'll generate an image.

```bash
dot -Tsvg -Kneato -O ~/Downloads/app.gv
```

- Red Nodes mean the component was not in this app config and the graph used the most recent config for that component.
- Blue Nodes mean the component was in this app config.

Note, for more complex viewing recommend to copy the output directly into [this
viewer](https://dreampuf.github.io/GraphvizOnline).
