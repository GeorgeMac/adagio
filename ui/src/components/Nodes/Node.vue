<template>
  <div class="container has-text-left" v-if="this.node != null">
    <p class="title">node details</p>
    <b-field grouped>
      <b-field label="name" label-position="inside">
        <b-input disabled :value="node.spec.name">{{ node.spec.name }}</b-input>
      </b-field>
      <b-field label="runtime" label-position="inside">
        <b-input disabled :value="node.spec.runtime">{{ node.spec.runtime }}</b-input>
      </b-field>
    </b-field >
    <div v-show="retryCount">
      <p class="has-gutter">Retries:</p>
      <b-field v-for="(value, key) in node.spec.retry" :key="key" grouped>
        <b-field label="condition" label-position="inside">
          <b-input disabled :value="key">{{ key }}</b-input>
        </b-field>
        <b-field label="maximum attempts" label-position="inside">
          <b-input disabled :value="value.max_attempts">{{ value.max_attempts }}</b-input>
        </b-field>
      </b-field>
    </div>
    <div v-if="node.attempts !== undefined">
      <p class="has-gutter">Attempts:</p>
      <b-field v-for="(result, key) in attemptsReversed()" :key="key" grouped>
        <b-field label="conclusion" label-position="inside">
          <b-input disabled :value="result.conclusion">{{ result.conclusion }}</b-input>
        </b-field>
      </b-field>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Node',
  props: {
    node: Object
  },
  computed: {
    retryCount() {
      var retry = this.node.spec.retry;
      return (retry ? Object.keys(retry).length : 0)
    }
  },
  methods: {
    attemptsReversed() {
      return this.node.attempts.slice().reverse()
    }
  }
}
</script>

<style>
.has-gutter {
  margin-bottom: 1rem;
}
</style>
