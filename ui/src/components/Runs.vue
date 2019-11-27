<template>
  <section class="section">
    <div class="container">
      <b-table :data="orderedRuns">
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
      <b-button @click="loadMore">Load More</b-button>
    </div>
  </section>
</template>

<script>
import { Adagio } from '@/services/adagio';
import Timestamp from 'timestamp-nano';

export default {
  name: 'Runs',
  data() {
    return {
      active: {},
      runs:   {}
    }
  },
  computed: {
    orderedRuns() {
      return Object.values(this.runs).sort((a, b) => {
        var aT = this.unixNano(a['created_at']);
        var bT = this.unixNano(b['created_at']);
        return bT - aT;
      });
    }
  },
  mounted() {
    this.getPage();

    this.intervalID = setInterval(this.getRuns, 2000);
  },
  beforeDestroy() {
    clearInterval(this.intervalID);
  },
  methods: {
    getRuns() {
      this.getPage();

      this.pollActiveRuns();
    },
    loadMore() {
      var offset = this.orderedRuns[this.orderedRuns.length - 1]['created_at'];
      this.getPage(offset);
    },
    unixNano(ts) {
      var t = Timestamp.fromString(ts)
      return (t.getTimeT() * 1000000000) + t.getNano()
    },
    getPage(offset = null) {
      Adagio.then((client) => {
        var request = {'limit': 10};
        if (offset != null) {
          request['start_ns'] = this.unixNano(offset) + 1;
        }

        client.apis.ControlPlane.ListRuns(request).then((resp) => {
          if (resp.body.runs) {
            this.updateRuns(resp.body.runs);
          }
        })
      });
    },
    updateRuns(runs) {
      runs.forEach((run) => {
        if (!this.runs[run['id']]) {
          this.$set(this.runs, run['id'], run);

          if (run.status != "COMPLETED") {
            this.active[run['id']] = run;
          }
        }
      });
    },
    pollActiveRuns() {
      Adagio.then((client) => {
        Object.keys(this.active).map((runID) => {
          client.apis.ControlPlane.Inspect({id: runID}).then((resp) => {
            this.$set(this.runs, runID, resp.body.run);
            if (resp.body.run.status == "COMPLETED") {
              delete this.active[runID]
            }
          });
        });
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
