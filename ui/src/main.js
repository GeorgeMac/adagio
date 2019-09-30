import Vue from 'vue';
import Buefy from 'buefy';
import App from './App.vue';
import 'buefy/dist/buefy.css';

import router from './router';

import { library } from '@fortawesome/fontawesome-svg-core';
// internal icons
import { faCheck, faPlus, faArrowUp, faSpinner } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";

library.add(faCheck, faPlus, faArrowUp, faSpinner);
Vue.component('vue-fontawesome', FontAwesomeIcon);

Vue.config.productionTip = false
Vue.use(Buefy, {
  defaultIconComponent: 'vue-fontawesome',
  defaultIconPack: 'fas',
  customIconPacks: {
    fas: {
      sizes: {
        default: "1x",
        "is-small": "1x",
        "is-medium": "2x",
        "is-large": "3x"
      },
      iconPrefix: ""
    }
  }
})

new Vue({
  router,
  render: h => h(App),
}).$mount('#app')
