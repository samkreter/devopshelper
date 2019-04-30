import Vue from 'vue'
import App from './App.vue'
import router from './router'
import {store} from './store/store'
import './registerServiceWorker'
import ArgonDashboard from './plugins/argon-dashboard'
import AuthService from './services/auth.service';

Vue.config.productionTip = false

Vue.prototype.$authService = new AuthService();

router.beforeEach((to, from, next) => {
  if (to.matched.some(record => record.meta.requiresAuth)) {
    
    let user = store.state.user
    if (!user){
      next({
        name: "login"
      })
    }
  }

  next()
})

Vue.use(ArgonDashboard)
new Vue({
  store: store,
  router: router,
  render: h => h(App)
}).$mount('#app')
