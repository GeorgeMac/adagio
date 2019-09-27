<template>
  <section class="section">
    <div class="container">
      <b-table :data="runs">
        <template slot-scope="props">
          <b-table-column field="id" label="ID" width="40">
            <RouterLink :to="{ name: 'run', params: { id: props.row.id } }">
              {{ props.row.id }}
            </RouterLink>
          </b-table-column>

          <b-table-column field="created_at" label="Created">
            <span>
              {{ new Date(props.row.created_at).toString() }}
            </span>
          </b-table-column>

          <b-table-column label="Status">
            <span :class="statusClass(props.row.status)">
              {{ props.row.status.toLowerCase() }}
            </span>
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
  name: 'Runs',
  data() {
    return {
      runs: []
    }
  },
  mounted() {
    this.getRuns();

    this.intervalID = setInterval(this.getRuns, 2000);
  },
  beforeDestroy() {
    clearInterval(this.intervalID);
  },
  methods: {
    getRuns() {
      Adagio.then((client) => {
        client.apis.ControlPlane.ListRuns().then((resp) => {
          this.runs = resp.body.runs.reverse();
        })
      });
    },
    statusClass(status) {
      switch (status) {
        case "RUNNING":
          return "tag is-warning"
        case "COMPLETED":
          return "tag is-success"
        default:
          return "tag is-light"
      }
    }
  }
}
</script>
