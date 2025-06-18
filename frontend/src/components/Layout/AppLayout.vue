<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Navigation -->
    <nav class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
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
            <!-- Data refresh indicator -->
            <div
              v-if="stocksStore.lastUpdated"
              class="hidden md:flex items-center text-sm text-gray-500"
            >
              <ClockIcon class="w-4 h-4 mr-1" />
              Last updated: {{ formatLastUpdated }}
            </div>

            <!-- Refresh button -->
            <button
              @click="refreshData"
              :disabled="stocksStore.isLoading"
              class="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ArrowPathIcon
                :class="[stocksStore.isLoading ? 'animate-spin' : '', 'w-4 h-4 mr-2']"
              />
              {{ stocksStore.isLoading ? 'Loading...' : 'Refresh' }}
            </button>

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
    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div class="px-4 py-6 sm:px-0">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  ChartBarIcon,
  ClockIcon,
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
const navigation: NavItem[] = [
  { name: 'Markets', href: '/', icon: TableCellsIcon },
  { name: 'Recommendations', href: '/recommendations', icon: StarIcon },
]

// Computed
const formatLastUpdated = computed(() => {
  if (!stocksStore.lastUpdated) return ''
  return new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(stocksStore.lastUpdated)
})

// Update navigation current state based on route
const updateNavigationState = () => {
  navigation.forEach((item) => {
    item.current = route.path === item.href
  })
}

// Methods
const refreshData = async () => {
  try {
    await Promise.all([stocksStore.fetchRatings(), stocksStore.fetchRecommendations()])
  } catch (error) {
    console.error('Failed to refresh data:', error)
  }
}

// Lifecycle
onMounted(() => {
  updateNavigationState()
  // Initial data load
  refreshData()
})

// Watch route changes to update navigation
import { watch } from 'vue'
watch(() => route.path, updateNavigationState)
</script>
 