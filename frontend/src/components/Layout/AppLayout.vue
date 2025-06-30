<template>
  <div class="min-h-screen bg-white">
    <!-- Navigation -->
    <nav class="bg-white shadow-sm border-b border-gray-200">
      <div class="mx-auto px-4 sm:px-6 lg:px-8" style="max-width: 1344px">
        <div class="flex justify-between h-16">
          <!-- Logo and main navigation -->
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <router-link to="/" class="flex items-center space-x-2">
                <div
                  class="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center"
                >
                  <ChartBarIcon class="w-5 h-5 text-white" />
                </div>
                <h1 class="text-xl font-bold text-gray-900">StockAnalyzer</h1>
              </router-link>
            </div>

            <!-- Desktop navigation -->
            <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
              <router-link
                v-for="item in navigation"
                :key="item.name"
                :to="item.href"
                :class="[
                  item.current
                    ? 'border-blue-500 text-gray-900'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                  'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium',
                ]"
              >
                <component :is="item.icon" class="w-4 h-4 mr-2" />
                {{ item.name }}
              </router-link>
            </div>
          </div>

          <!-- Right side items -->
          <div class="flex items-center space-x-4">
            <!-- Mobile menu button -->
            <button
              @click="mobileMenuOpen = !mobileMenuOpen"
              class="sm:hidden inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
            >
              <Bars3Icon v-if="!mobileMenuOpen" class="block h-6 w-6" />
              <XMarkIcon v-else class="block h-6 w-6" />
            </button>
          </div>
        </div>
      </div>

      <!-- Mobile menu -->
      <div v-show="mobileMenuOpen" class="sm:hidden">
        <div class="pt-2 pb-3 space-y-1 bg-white border-t border-gray-200">
          <router-link
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            @click="mobileMenuOpen = false"
            :class="[
              item.current
                ? 'bg-blue-50 border-blue-500 text-blue-700'
                : 'border-transparent text-gray-600 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-800',
              'block pl-3 pr-4 py-2 border-l-4 text-base font-medium',
            ]"
          >
            <component :is="item.icon" class="inline w-4 h-4 mr-2" />
            {{ item.name }}
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Error alert -->
    <div v-if="stocksStore.hasError" class="bg-red-50 border-l-4 border-red-400 p-4">
      <div class="flex items-center justify-between">
        <div class="flex">
          <ExclamationTriangleIcon class="h-5 w-5 text-red-400" />
          <div class="ml-3">
            <p class="text-sm text-red-700">
              {{ stocksStore.error }}
            </p>
          </div>
        </div>
        <button @click="stocksStore.clearError" class="text-red-400 hover:text-red-600">
          <XMarkIcon class="h-5 w-5" />
        </button>
      </div>
    </div>

    <!-- Loading indicator -->
    <div v-if="stocksStore.isLoading" class="bg-blue-50 border-l-4 border-blue-400 p-4">
      <div class="flex items-center">
        <ArrowPathIcon class="h-5 w-5 text-blue-400 animate-spin" />
        <div class="ml-3">
          <p class="text-sm text-blue-700">Loading data...</p>
        </div>
      </div>
    </div>

    <!-- Main content -->
    <main class="mx-auto py-6 sm:px-6 lg:px-8" style="max-width: 1375px">
      <div class="mx-auto px-4 sm:px-6 lg:px-8" style="max-width: 1375px">
        <div class="px-4 py-6 sm:px-0">
          <router-view />
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  ChartBarIcon,
  ArrowPathIcon,
  Bars3Icon,
  XMarkIcon,
  ExclamationTriangleIcon,
  StarIcon,
  TableCellsIcon,
} from '@heroicons/vue/24/outline'
import { useStocksStore } from '@/stores/stocks'
import type { NavItem } from '@/types'

// Store
const stocksStore = useStocksStore()
const route = useRoute()

// Local state
const mobileMenuOpen = ref(false)

// Navigation items
const navigation = ref<NavItem[]>([
  { name: 'Markets', href: '/', icon: TableCellsIcon, current: false },
  { name: 'Recommendations', href: '/recommendations', icon: StarIcon, current: false },
])

// Update navigation current state based on route
const updateNavigationState = () => {
  navigation.value.forEach((item) => {
    item.current =
      route.path === item.href || (route.path.startsWith('/stock/') && item.href === '/')
  })
}

// Lifecycle
onMounted(() => {
  updateNavigationState()
  // Don't fetch data here - let individual views handle their own data needs
  // This prevents duplicate API calls when AppLayout and views both try to fetch on mount
})

// Watch route changes to update navigation
import { watch } from 'vue'
watch(() => route.path, updateNavigationState)
</script>
