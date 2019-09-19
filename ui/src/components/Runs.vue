<template>
  <section class="section">
    <div class="container">
      <ListItem v-for="run in runs" v-bind="run" :key="run.id" />
    </div>
  </section>
</template>

<script>
import { Adagio } from '@/services/adagio';
import ListItem from './Runs/ListItem';

export default {
  name: 'Runs',
  components: {
    ListItem
  },
  data() {
    return {
      runs: []
    }
  },
  mounted() {
    this.getRuns();

    setInterval(this.getRuns, 2000);
  },
  methods: {
    getRuns() {
      Adagio.then((client) => {
        client.apis.ControlPlane.List().then((resp) => {
          this.runs = resp.body.runs.reverse();
        })
      });
    }
  }
}
</script>
