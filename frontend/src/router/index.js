import { createRouter, createWebHashHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Connections from '../views/Connections.vue'
import Query from '../views/Query.vue'
import Settings from '../views/Settings.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/connections',
    name: 'Connections',
    component: Connections
  },
  {
    path: '/query',
    name: 'Query',
    component: Query
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
