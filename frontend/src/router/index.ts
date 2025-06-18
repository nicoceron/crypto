import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'markets',
      component: () => import('../views/MarketsView.vue'),
    },
    {
      path: '/recommendations',
      name: 'recommendations',
      component: () => import('../views/RecommendationsView.vue'),
    },
    {
      path: '/stock/:ticker',
      name: 'stock-detail',
      component: () => import('../views/StockDetailView.vue'),
      props: true,
    },
  ],
})

export default router
