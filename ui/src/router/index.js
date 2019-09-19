import Vue from "vue";
import Router from "vue-router";
import Runs from "@/components/Runs";
import Run from "@/components/Runs/Run";
import New from "@/components/Runs/New";

Vue.use(Router);

export default new Router({
  routes: [
    {
      path: "/",
      name: "runs",
      component: Runs
    },
    {
      path: "/runs/new",
      name: "new_run",
      component: New
    },
    {
      path: "/runs/:id",
      name: "run",
      component: Run 
    }
  ]
});
