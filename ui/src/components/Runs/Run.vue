<template>
  <section class="section">
    <div class="columns is-centered">
      <div class="column">
        <figure class="image">
          <Graph v-bind:run="run" @clicked="onNodeClicked"/>
        </figure>
      </div>
      <div class="column has-text-left">
        <div class="content">
          <p class="title">run details</p>
          <p class="subtitle">{{ run.id }}</p>
          <div class="content">
            <span :class="status.class">
              {{ run.status.toLowerCase() }}
              <b-icon pack="fas" :icon="status.icon" :custom-class="status.custom"></b-icon>
            </span>
          </div>
        </div>
        <hr />
        <Node :node="selectedNode" />
      </div>
    </div>
  </section>
</template>

<script>
import { Adagio } from '@/services/adagio';
import Graph from '../Graphs/Graph';
import Node from '../Nodes/Node';

export default {
  name: 'Runs',
  components: {
    Graph,
    Node
  },
  computed: {
    id() {
      return this.$route.params.id
    },
    statusText() {
      var status = this.run.status.toLowerCase();
      return `${status}`;
    },
    status() {
      switch (this.run.status) {
        case "RUNNING":
          return {
            icon: "spinner",
            custom: "fa-spin",
            class: "tag is-warning"
          }
        case "COMPLETED":
          return {
            icon: "check",
            custom: "",
            class: "tag is-success"
          }
        default:
          return {
            icon: "spinner",
            custom: "",
            class: "tag is-light"
          }
      }
    }
  },
  data() {
    return {
      run: {
        status: 'WAITING',
        nodes: [],
        edges: []
      },
      selectedNode: null
    }
  },
  mounted() {
    this.getRun(true);

    this.intervalID = setInterval((function() { this.getRun(false) }).bind(this), 500);
  },
  beforeDestroy() {
    clearInterval(this.intervalID);
  },
  methods: {
    getRun(first) {
      if (this.run.status == 'COMPLETED' && !first) {
        return
      }

      Adagio.then((client) => {
        client.apis.ControlPlane.Inspect({ id: this.id }).then((resp) => {
          this.run = resp.body.run;
        })
      });
    },
    onNodeClicked(n) {
      this.selectedNode = n;
    }
  }
}
</script>

<style>
.tag .icon {
  padding-left: 0.5rem;
}
</style>
