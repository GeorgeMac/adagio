<template>
  <div class="level notification">
      <p class="level-item">
        <strong>
          <RouterLink :to="{ name: 'run', params: { id: id } }">{{ id }}</RouterLink>:
        </strong>
      </p>
      <p class="level-item">{{ createdAt }}</p>
      <b-icon class="level-item" :icon="icon"></b-icon> 
  </div>
</template>

<script>
export default {
  name: "ListItem",
  props: {
    id: String,
    created_at: String,
    status: String,
    nodes: Array,
    edges: Array
  },
  computed: {
    createdAt() {
      return new Date(this.created_at).toString()
    },
    icon() {
      switch (this.status) {
        case "RUNNING":
          return "spinner"
        case "COMPLETED":
          return "check"
        default:
          return "clock"
      }
    }
  }
};
</script>
