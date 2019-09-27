<template>
  <section class="section">
    <div class="container">
      <b-table :data="agents">
        <template slot-scope="props">
          <b-table-column field="id" label="ID" width="40">
            <RouterLink :to="{ name: 'agent', params: { id: props.row.id } }">
              {{ props.row.id }}
            </RouterLink>
          </b-table-column>

          <b-table-column field="runtimes" label="Runtimes">
            <b-tag v-for="(runtime, idx) in props.row.runtimes" :key="idx">
              {{ runtime.name }}
            </b-tag>
          </b-table-column>
        </template>

        <template slot="empty">
          <section class="section">
            <div class="content has-text-grey has-text-centered">
              <p>Nothing here.</p>
            </div>
          </section>
        </template>
      </b-table>
    </div>
  </section>
</template>

<script>
import { Adagio } from '@/services/adagio';

export default {
  name: 'Agents',
  data() {
    return {
      agents: []
    }
  },
  mounted() {
    this.getAgents();

    this.intervalID = setInterval(this.getAgents, 5000);
  },
  beforeDestroy() {
    clearInterval(this.intervalID);
  },
  methods: {
    getAgents() {
      Adagio.then((client) => {
        client.apis.ControlPlane.ListAgents().then((resp) => {
          if (this.agents !== undefined) {
            this.agents = resp.body.agents;
          }
        })
      });
    }
  }
}
</script>

<style scoped>
span.tag {
  margin-right: 0.75em;
}
</style>
