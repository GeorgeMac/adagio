<template>
  <section class="section">
    <div class="container">
      <b-field label="Graph Specification">
        <b-input type="textarea" rows="20" v-model="spec"></b-input>
      </b-field>
      <button
        class="button is-primary"
        @click.prevent="createRun()"
      >
        Create
      </button>
    </div>
  </section>
</template>

<script>
import { Adagio } from '@/services/adagio';

export default {
  name: 'New',
  data() {
    return {
      spec: ""
    }
  },
  methods: {
    specJSON() {
      return JSON.parse(this.spec)
    },
    createRun() {
      Adagio.then((client) => {
        client.apis.ControlPlane.Start({ body: { spec: this.specJSON() } }).then((resp) => {
          this.$router.push('/runs/' + resp.body.run.id)
        })
      });
    }
  }
}
</script>
