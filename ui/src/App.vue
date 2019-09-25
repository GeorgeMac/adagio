<template>
  <div id="app" class="container">
    <header>
      <b-navbar>
        <template slot="brand">
          <b-navbar-item tag="router-link" :to="{ name: 'home' }">
            <img src="./assets/images/logo.png" />
            <strong>ADAGIO</strong>
          </b-navbar-item>
        </template>
        <template slot="start">
          <b-navbar-item tag="router-link" :to="{ name: 'runs' }">
            Runs
          </b-navbar-item>
          <b-navbar-item href="#">
            Agents
          </b-navbar-item>

          <b-navbar-dropdown label="Info">
            <b-navbar-item href="#">
              About
            </b-navbar-item>
            <b-navbar-item href="#">
              Documentation
            </b-navbar-item>
          </b-navbar-dropdown>
        </template>

        <template slot="end">
          <b-navbar-item tag="div">
            <div class="buttons">
              <RouterLink class="button is-default" :to="{ name: 'new_run' }">
                <strong>New Run</strong>
                <b-icon icon="plus"></b-icon>
              </RouterLink>
            </div>
          </b-navbar-item>
        </template>
      </b-navbar>
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
