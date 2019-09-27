<template>
  <section class="section">
    <div class="columns">
      <div class="column">
        <div class="card">
          <div class="card-content">
            <p class="title">
            {{ this.stats.run_count || 0 }} runs
            </p>
            <p>{{ this.stats.node_counts.waiting_count || 0}} waiting nodes</p>
            <p>{{ this.stats.node_counts.ready_count || 0}} ready nodes</p>
            <p>{{ this.stats.node_counts.running_count || 0 }} running nodes</p>
            <p>{{ this.stats.node_counts.completed_count || 0}} completed nodes</p>
          </div>
          <footer class="card-footer">
            <p class="card-footer-item">
            <span>
              <RouterLink :to="{ name: 'runs' }">
              View Runs
              </RouterLink>
            </span>
            </p>
          </footer>
        </div>
      </div>
      <div class="column">
        <div class="card">
          <div class="card-content">
            <p class="title">
            {{ this.agents.length }} agents
            </p>
          </div>
          <footer class="card-footer">
            <p class="card-footer-item">
            <span>
              <RouterLink :to="{ name: 'agents' }">
              View Agents
              </RouterLink>
            </span>
            </p>
          </footer>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
import { Adagio } from '@/services/adagio';

export default {
  name: 'Home',
  data() {
    return {
      stats: {
        run_count: 0,
        node_counts: {
          waiting_count: 0,
          ready_count: 0,
          running_count: 0,
          completed_count: 0
        }
      },
      agents: []
    }
  },
  mounted() {
    this.getAll();

    this.intervalID = setInterval(this.getAll, 2000);
  },
  beforeDestroy() {
    clearInterval(this.intervalID);
  },
  methods: {
    getAll() {
      this.getStats();
      this.getAgents();
    },
    getStats() {
      Adagio.then((client) => {
        client.apis.ControlPlane.Stats().then((resp) => {
          this.stats = resp.body.stats;
        })
      });
    },
    getAgents() {
      Adagio.then((client) => {
        client.apis.ControlPlane.ListAgents().then((resp) => {
          if (resp.body.agents !== undefined) {
            this.agents = resp.body.agents;
          }
        })
      });
    }
  }
}
</script>
