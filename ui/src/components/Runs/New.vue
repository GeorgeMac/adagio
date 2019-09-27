<template>
  <section class="columns">
    <div class="column">
      <b-tabs v-model="tab" type="is-boxed">
        <b-tab-item label="Editor" id="editor">
          <div class="card">
            <div class="card-content">
              <!-- nodes -->
              <b-field label="Name" horizontal>
                <b-input placeholder="Give it a name" v-model="currentNode.name"></b-input>
              </b-field>
              <b-field label="Runtime" horizontal>
                <b-select placeholder="Select a runtime" v-model="currentNode.runtime">
                  <option v-for="runtime in runtimes" :value="runtime" :key="runtime">
                  {{ runtime }}
                  </option>
                </b-select>
              </b-field>
            </div>
            <div class="card-footer">
              <a
                class="is-primary card-footer-item"
                @click.prevent="createNode()"
                >
                Add Node
              </a>
            </div>
          </div>
          <div class="card">
            <div class="card-content">
              <!-- edges -->
              <b-field label="Source" horizontal>
                <b-select placeholder="Select a node" v-model="currentEdge.source">
                  <option v-for="node in spec.nodes" :value="node.name" :key="node.name">
                  {{ node.name }}
                  </option>
                </b-select>
              </b-field>
              <b-field label="Destination" horizontal>
                <b-select placeholder="Select an edge" v-model="currentEdge.destination">
                  <option v-for="node in spec.nodes" :value="node.name" :key="node.name">
                  {{ node.name }}
                  </option>
                </b-select>
              </b-field>
            </div>
            <footer class="card-footer">
              <a
                class="is-primary card-footer-item"
                @click.prevent="createEdge()"
                >
                Add Edge
              </a>
            </footer>
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
        Create
      </button>
    </div>
    <div class="column" id="graphEditorColumn">
      <div class="container" id="editGraph">
        <Graph v-bind:run="specToRun()" />
      </div>
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
          this.$router.push('/runs/' + resp.body.run.id)
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
div#graphEditorColumn {
  padding-top: 20%;
}

div.card {
  margin-bottom: 1rem;
}
</style>
