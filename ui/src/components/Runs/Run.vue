<template>
  <section class="section">
    <div class="container">
      <Graph v-bind:run="run" />
    </div>
  </section>
</template>

<script>
import { Adagio } from '@/services/adagio';
import Graph from '../Graphs/Graph';

export default {
  name: 'Runs',
  components: {
    Graph
  },
  computed: {
    id() {
      return this.$route.params.id
    }
  },
  data() {
    return {
      run: {
        status: 'WAITING',
        nodes: [],
        edges: []
      }
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
    }
  }
}
</script>
