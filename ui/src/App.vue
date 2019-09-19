<template>
  <div id="app">
    <header>
      <nav class="level">
        <p class="level-item has-text-centered">
          <RouterLink class="link is-info" :to="{ name: 'runs' }">runs</RouterLink> 
        </p>
        <p class="level-item has-text-centered">
          <strong>ADAGIO</strong>
        </p>
        <p class="level-item has-text-centered">
          <RouterLink class="link is-info" :to="{ name: 'new_run' }">
            <b-icon class="level-item" icon="plus"></b-icon>
          </RouterLink> 
        </p>
      </nav>
    </header>

    <RouterView />
  </div>
</template>

<script>
export default {
  name: 'app'
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
