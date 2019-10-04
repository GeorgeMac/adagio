<template>
  <div class="container has-text-left" :style="style">
    <p class="title">node details</p>
    <b-field grouped>
      <b-field label="name" label-position="inside">
        <b-input disabled :value="name">{{ name }}</b-input>
      </b-field>
      <b-field label="runtime" label-position="inside">
        <b-input disabled :value="runtime">{{ runtime }}</b-input>
      </b-field>
    </b-field>
  </div>
</template>

<script>
export default {
  name: 'Node',
  props: {
    node: Object
  },
  computed: {
    name() {
      if (this.node == null) { return "" }
      return this.node.spec.name
    },
    runtime() {
      if (this.node == null) { return "" }
      return this.node.spec.runtime
    },
    style() {
      if (this.node === undefined || this.node == null) {
        return "display: none";
      }

      return ""
    }
  }
}
</script>

<style>
</style>
