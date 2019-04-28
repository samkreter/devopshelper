import Vue from 'vue'
import Router from 'vue-router'
import DashboardLayout from '@/layout/DashboardLayout'

Vue.use(Router)

export default new Router({
  linkExactActiveClass: 'active',
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import( './views/Login.vue')
    },
    {
      path: '/',
      redirect: 'dashboard',
      component: DashboardLayout,
      children: [
        {
          path: '/repositories',
          name: 'repositories',
          component: () => import( './views/Repositories.vue')
        },
        {
          path: '/dashboard',
          name: 'dashboard',
          // route level code-splitting
          // this generates a separate chunk (about.[hash].js) for this route
          // which is lazy-loaded when the route is visited.
          component: () => import('./views/Dashboard.vue')
        }
        // {
        //   path: '/icons',
        //   name: 'icons',
        //   component: () => import(/* webpackChunkName: "demo" */ './views/Icons.vue')
        // },
        // {
        //   path: '/profile',
        //   name: 'profile',
        //   component: () => import(/* webpackChunkName: "demo" */ './views/UserProfile.vue')
        // },
        // {
        //   path: '/maps',
        //   name: 'maps',
        //   component: () => import(/* webpackChunkName: "demo" */ './views/Maps.vue')
        // }
      ]
    },
    // {
    //   path: '/',
    //   redirect: 'login',
    //   component: AuthLayout,
    //   children: [
    //     {
    //       path: '/login',
    //       name: 'login',
    //       component: () => import(/* webpackChunkName: "demo" */ './views/Login.vue')
    //     },
    //     {
    //       path: '/register',
    //       name: 'register',
    //       component: () => import(/* webpackChunkName: "demo" */ './views/Register.vue')
    //     }
    //   ]
    // }
  ]
})
