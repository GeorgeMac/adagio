<template>
  <div id="run">
    <div :id="run.id" class="container">
    </div>
  </div>
</template>

<script>
import * as d3 from 'd3';
import dagreD3 from 'dagre-d3';

export default {
  name: 'Graph',
  props: {
    run: Object
  },
  mounted() {
    this.generateGraph();
  },
  updated() {
    this.generateGraph();
  },
  methods: {
    generateGraph() {
      if (this.run.nodes.length == 0) {
        return
      }

      d3.select("div.container svg#graph").remove();

      // Create the input graph
      var g = new dagreD3.graphlib.Graph()
      .setGraph({rankdir: 'LR'})
      .setDefaultEdgeLabel(function() { return {}; });

      this.run.nodes.forEach((node) => {
        var cls = "node-default";
        if (node.status == "RUNNING") {
          cls = "node-running";
        }

        var attempts = node.attempts;
        if (attempts !== undefined && attempts.length > 0) {
          switch (attempts[attempts.length - 1].conclusion) {
            case "SUCCESS":
              cls = "node-success";
              break;
            case "FAIL":
              cls = "node-fail";
              break;
          case "ERROR":
              cls = "node-error";
              break;
          }
        }

        var label = `"${node.spec.name}" runtime: "${node.spec.runtime}"`;

        var n = g.setNode(node.spec.name, {
          label: label,
          class: cls,
          id: `node-${node.spec.name}`
        });

        n.rx = n.ry = 5;
      });

      if (this.run.edges !== undefined) {
        this.run.edges.forEach((edge) => {
          g.setEdge(edge.source, edge.destination);
        });
      }

      // Create the renderer
      var render = new dagreD3.render();

      var container = document.getElementById(this.run.id);
      var svg = d3.select(container).append("svg");
      svg.attr("id", "graph");
      var svgGroup = svg.append("g");

      // Run the renderer. This is what draws the final graph.
      render(svgGroup, g);

      // Register handlers
      var parent = this;
      svgGroup.selectAll("g.node").on("click", function() {
        var node = d3.select(this);
        var cls = node.attr("class");
        if (cls.includes("selected")) {
          cls = cls.replace(" selected", "");
        } else {
          cls += " selected";
        }

        node.attr("class", cls);

        parent.$emit('clicked', node.attr('id').slice(5));
      });

      // Center the graph
      svg.attr("width", g.graph().width + 100);
      var xCenterOffset = (svg.attr("width") - g.graph().width) / 2;
      svgGroup.attr("transform", "translate(" + xCenterOffset + ", 20)");
      svg.attr("height", g.graph().height + 40);
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style>
  g.node-success > rect {
    fill: hsl(141, 71%, 48%);
  }

  g.node-fail > rect {
    fill: hsl(348, 100%, 61%);
  }

  g.node-error > rect {
    fill: hsl(48, 100%, 67%);
  }

  g.node-running > rect {
    fill: hsl(0, 0%, 96%)
  }

  text {
    font-weight: 300;
    font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
    font-size: 14px;
  }

  .node rect {
    stroke: #999;
    fill: #fff;
    stroke-width: 1.5px;
  }

  #editGraph .node rect:hover {
    stroke: hsl(48, 100%, 67%);
    cursor: pointer;
  }

  #editGraph .node .label tspan {
    pointer-events: none;
  }

  #editGraph .node.selected rect {
    stroke: hsl(48, 100%, 67%);
    stroke-width: 3px;
  }

  .edgePath path {
    stroke: #333;
    stroke-width: 1.5px;
  }
</style>
