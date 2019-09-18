<template>
  <div class="level notification">
    <div class="level-left">
      <p class="level-item"><strong>{{ id }}:</strong></p>
      <p class="level-item">{{ createdAt }}</p>
    </div>
    <div class="level-right">
      <b-icon class="level-item" :icon="icon"></b-icon> 
    </div>
  </div>
  <!--
  <div class="tile is-ancestor notification">
    <div class="tile notification is-vertical">
      <div class="tile is-parent">
        <article class="tile is-child  has-text-left">
          <p class="title">{{ id }}</p>
          <p>{{ createdAt }}</p>
        </article>
      </div>
      <div class="tile is-parent">
        <div class="tile is-child">
          <b-icon :icon="icon"></b-icon> 
        </div>
      </div>
    </div>
  </div>
  -->
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
