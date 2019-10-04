<template>
  <section class="columns is-centered">
    <div class="column is-8">
      <b-tabs v-model="tab" type="is-boxed">
        <b-tab-item label="Editor" id="editor">
          <!-- nodes -->
          <b-field label="Node" label-position="on-border">
            <b-input placeholder="Give it a name" v-model="currentNode.name"></b-input>
            <b-select placeholder="Select a runtime" v-model="currentNode.runtime">
              <option v-for="runtime in runtimes" :value="runtime" :key="runtime">
              {{ runtime }}
              </option>
            </b-select>
            <p class="control">
              <b-button class="button is-primary" @click.prevent="createNode()">
                Add Node
              </b-button>
            </p>
          </b-field>
          <!-- graph demo container -->
          <div class="container" id="editGraph">
            <Graph v-bind:run="specToRun()" @clicked="onNodeClicked" />
          </div>
        </b-tab-item>
        <b-tab-item label="Raw" id="raw">
          <div class="container">
            <b-field label="Graph Specification">
              <b-input type="textarea" rows="20" v-model="specRaw"></b-input>
            </b-field>
          </div>
        </b-tab-item>
      </b-tabs>
      <button
        class="button is-primary"
        @click.prevent="createRun()"
        >
        Create Run
      </button>
    </div>
  </section>
</template>

<script>
import { Adagio } from '@/services/adagio';
import Graph from '@/components/Graphs/Graph';

export default {
  name: 'New',
  components: {
    Graph
  },
  data() {
    return {
      tab: 0,
      runtimes: [],
      selectedNode: null,
      currentNode: {
        name:    "",
        runtime: null
      },
      currentEdge: {
        source:      null,
        destination: null
      },
      spec: {
        nodes: [],
        edges: []
      },
      specRaw: "",
    }
  },
  beforeMount() {
    this.getRuntimes();
  },
  beforeUpdate() {
    if (this.editorInFocus()) {
      this.editorToRaw();
      return
    }

    this.rawToEditor();
  },
  methods: {
    editorInFocus() {
      return this.tab == 0;
    },
    editorToRaw() {
      this.specRaw = JSON.stringify(this.spec, null, 2);
    },
    rawToEditor() {
      this.spec = JSON.parse(this.specRaw);
    },
    nodes() {
      return this.spec.nodes.map((n) => { n.name })
    },
    createNode() {
      this.spec.nodes.push(this.currentNode);
      this.currentNode = {
        name:    "",
        runtime: ""
      };
    },
    createEdge() {
      this.spec.edges.push(this.currentEdge);
      this.currentEdge = {
        source:      "",
        destination: ""
      };
    },
    specToRun() {
      return {
        id: "editGraph",
        nodes: this.spec.nodes.map((spec) => {
          return {
            spec: spec
          }
        }),
        edges: this.spec.edges
      }
    },
    onNodeClicked(n) {
      if (this.selectedNode === null) {
        this.selectedNode = n.spec.name;
        return
      }

      if (this.selectedNode === n.spec.name) {
        return
      }

      this.spec.edges.push({
        source:      this.selectedNode,
        destination: n.spec.name,
      });

      this.selectedNode = null;
    },
    specPayload() {
      if (this.editorInFocus()) {
        return this.spec;
      }

      return JSON.parse(this.specRaw)
    },
    createPayload() {
      return {
        body: {
          spec: this.specPayload()
        }
      }
    },
    createRun() {
      Adagio.then((client) => {
        client.apis.ControlPlane.Start(this.createPayload()).then((resp) => {
          this.$buefy.snackbar.open({
              duration: 5000,
              message: 'run started',
              type: 'is-success',
              position: 'is-top',
              actionText: 'View',
              queue: false,
              onAction: () => {
                this.$router.push('/runs/' + resp.body.run.id);
              }
          })
        })
      });
    },
    getRuntimes() {
      Adagio.then((client) => {
        client.apis.ControlPlane.ListAgents().then((resp) => {
          var agents = resp.body.agents;
          if (agents !== undefined) {
            this.runtimes = agents
            .flatMap((agent) => {
              return agent.runtimes.map((r) => {
                return r.name
              })
            })
            .filter((r, i, all) => {
              return all.indexOf(r) === i;
            });
          }
        })
      });
    },
  }
}
</script>

<style scoped>
div.card {
  margin-bottom: 1rem;
}

#editGraph {
  min-height: 5rem;
  border: 1px dashed;
}
</style>
