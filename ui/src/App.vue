<template>
  <div id="app">
    <header>
    </header>
    <section class="hero">
      <div class="hero-body">
        <div class="container">
          <h1 class="title">
            Runs
          </h1>
        </div>
      </div>
    </section>

    <Runs />
  </div>
</template>

<script>
import Runs from './components/Runs.vue'

export default {
  name: 'app',
  components: {
    Runs
  }
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
