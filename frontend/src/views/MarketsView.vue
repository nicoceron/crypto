<template>
  <div class="space-y-6">
    <!-- Page header with total market stats -->
    <div class="bg-white shadow rounded-lg p-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-gray-900">Stock Markets</h1>
          <p class="text-sm text-gray-600 mt-1">
            The global stock market ratings today.
            <span class="text-red-600">📉 -1.2%</span> change in the last 24 hours.
            <button class="text-blue-600 hover:underline ml-1">Read more</button>
          </p>
        </div>
        <div class="text-right">
          <div class="text-2xl font-bold text-gray-900">$2,847,392,841,820</div>
          <div class="text-sm text-gray-500">Total Market Cap 📉 -1.2%</div>
        </div>
      </div>

      <!-- Mini stats section -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-6">
        <div>
          <div class="text-lg font-semibold text-gray-900">$142,846,810,797</div>
          <div class="text-sm text-gray-500">24h Trading Volume</div>
        </div>
        <div>
          <div class="text-lg font-semibold text-gray-900">
            {{ stocksStore.totalRatings.toLocaleString() }}
          </div>
          <div class="text-sm text-gray-500">Total Rated Stocks</div>
        </div>
      </div>
    </div>

    <!-- Trending and Top Gainers section -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Trending (Recommendations) -->
      <div class="bg-white shadow rounded-lg p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900 flex items-center">
            🔥 Trending Recommendations
          </h3>
          <router-link to="/recommendations" class="text-blue-600 hover:underline text-sm">
            View more →
          </router-link>
        </div>

        <div v-if="stocksStore.recommendations.length === 0" class="text-center py-4">
          <div class="text-gray-500">Loading recommendations...</div>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="rec in stocksStore.topRecommendations.slice(0, 3)"
            :key="rec.ticker"
            class="flex items-center justify-between p-2 hover:bg-gray-50 rounded-lg cursor-pointer"
            @click="$router.push(`/stock/${rec.ticker}`)"
          >
            <div class="flex items-center space-x-3">
              <StockLogo :symbol="rec.ticker" size="xs" />
              <div>
                <div class="text-sm font-medium text-gray-900">{{ rec.ticker }}</div>
                <div class="text-xs text-gray-500">{{ rec.company || rec.ticker }}</div>
              </div>
            </div>
            <div class="text-right">
              <div class="text-sm font-medium text-green-600">
                ${{ (rec.target_price || 0).toFixed(2) }}
              </div>
              <div class="text-xs text-green-600">↗ {{ (rec.score || 0).toFixed(1) }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Top Gainers (from ratings) -->
      <div class="bg-white shadow rounded-lg p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900 flex items-center">🚀 Recent Upgrades</h3>
          <button class="text-blue-600 hover:underline text-sm">View more →</button>
        </div>

        <div v-if="recentUpgrades.length === 0" class="text-center py-4">
          <div class="text-gray-500">Loading recent upgrades...</div>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="rating in recentUpgrades.slice(0, 3)"
            :key="rating.rating_id"
            class="flex items-center justify-between p-2 hover:bg-gray-50 rounded-lg cursor-pointer"
            @click="$router.push(`/stock/${rating.ticker}`)"
          >
            <div class="flex items-center space-x-3">
              <StockLogo :symbol="rating.ticker" size="xs" />
              <div>
                <div class="text-sm font-medium text-gray-900">{{ rating.ticker }}</div>
                <div class="text-xs text-gray-500">{{ rating.brokerage }}</div>
              </div>
            </div>
            <div class="text-right">
              <div class="text-sm font-medium text-green-600">
                ${{ (rating.target_to || 0).toFixed(2) }}
              </div>
              <div class="text-xs text-green-600">↗ {{ rating.rating_to }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Main table -->
    <div class="bg-white shadow rounded-lg">
      <!-- Search and filters -->
      <div class="p-6 border-b border-gray-200">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <!-- Search -->
          <div>
            <div class="relative">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search by ticker, company..."
                class="block w-full pr-10 border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                @input="debouncedSearch"
              />
              <div class="absolute inset-y-0 right-0 pr-3 flex items-center">
                <MagnifyingGlassIcon class="h-5 w-5 text-gray-400" />
              </div>
            </div>
          </div>

          <!-- Sort by -->
          <div>
            <select
              v-model="sortBy"
              class="block w-full border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              @change="handleSort"
            >
              <option value="updated_at">Last Updated</option>
              <option value="ticker">Ticker</option>
              <option value="firm">Firm</option>
              <option value="rating">Rating</option>
              <option value="price_target">Price Target</option>
            </select>
          </div>

          <!-- Sort order -->
          <div>
            <select
              v-model="sortOrder"
              class="block w-full border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              @change="handleSort"
            >
              <option value="desc">High to Low</option>
              <option value="asc">Low to High</option>
            </select>
          </div>

          <!-- Page size -->
          <div>
            <select
              v-model="pageSize"
              class="block w-full border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              @change="handlePageSizeChange"
            >
              <option value="10">10 per page</option>
              <option value="20">20 per page</option>
              <option value="50">50 per page</option>
              <option value="100">100 per page</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Enhanced table -->
      <div class="overflow-x-auto">
        <table class="min-w-full">
          <thead class="bg-gray-50 border-b border-gray-200">
            <tr>
              <th
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-12"
              >
                #
              </th>
              <th
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
              >
                Stock
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
              >
                Price Target
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
              >
                Rating
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
              >
                Analyst
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
              >
                Last 7 Days
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr
              v-for="(rating, index) in stocksStore.ratings"
              :key="rating.rating_id"
              class="hover:bg-gray-50 cursor-pointer"
              @click="$router.push(`/stock/${rating.ticker}`)"
            >
              <!-- Rank -->
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ (stocksStore.currentPage - 1) * pageSize + index + 1 }}
              </td>

              <!-- Stock -->
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <StockLogo :symbol="rating.ticker" size="sm" />
                  <div class="ml-3">
                    <div class="text-sm font-medium text-gray-900">
                      {{ rating.ticker }}
                    </div>
                    <div class="text-sm text-gray-500 truncate max-w-32">
                      {{ rating.company }}
                    </div>
                  </div>
                </div>
              </td>

              <!-- Price Target with trend indicator -->
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="flex items-center justify-end space-x-2">
                  <div class="text-right">
                    <div class="text-sm font-medium text-gray-900">
                      ${{ (rating.target_to || 0).toFixed(2) }}
                    </div>
                    <div class="text-xs text-gray-500">USD</div>
                  </div>
                  <div class="flex items-center">
                    <component
                      :is="getTargetTrendIcon(rating)"
                      class="h-4 w-4"
                      :class="getTargetTrendColor(rating)"
                    />
                  </div>
                </div>
              </td>

              <!-- Rating -->
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <span
                  :class="getRatingColor(rating.rating_to)"
                  class="inline-flex px-2 py-1 text-xs font-semibold rounded-full"
                >
                  {{ rating.rating_to }}
                </span>
              </td>

              <!-- Analyst -->
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="text-sm text-gray-900">{{ rating.brokerage }}</div>
                <div class="text-xs text-gray-500">{{ formatDate(rating.time) }}</div>
              </td>

              <!-- Mini chart -->
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="w-20 h-8">
                  <MiniChart :symbol="rating.ticker" :rating="rating" period="1W" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div
        v-if="stocksStore.totalPages > 1"
        class="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6"
      >
        <div class="flex flex-1 justify-between sm:hidden">
          <button
            @click="stocksStore.changePage(stocksStore.currentPage - 1)"
            :disabled="stocksStore.currentPage <= 1"
            class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <button
            @click="stocksStore.changePage(stocksStore.currentPage + 1)"
            :disabled="stocksStore.currentPage >= stocksStore.totalPages"
            class="relative ml-3 inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Next
          </button>
        </div>
        <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
          <div>
            <p class="text-sm text-gray-700">
              Showing
              <span class="font-medium">
                {{ (stocksStore.currentPage - 1) * pageSize + 1 }}
              </span>
              to
              <span class="font-medium">
                {{ Math.min(stocksStore.currentPage * pageSize, stocksStore.totalRatings) }}
              </span>
              of
              <span class="font-medium">{{ stocksStore.totalRatings.toLocaleString() }}</span>
              results
            </p>
          </div>
          <div>
            <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
              <button
                @click="stocksStore.changePage(stocksStore.currentPage - 1)"
                :disabled="stocksStore.currentPage <= 1"
                class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronLeftIcon class="h-5 w-5" />
              </button>

              <button
                v-for="page in visiblePages"
                :key="page"
                @click="stocksStore.changePage(page)"
                :class="[
                  page === stocksStore.currentPage
                    ? 'z-10 bg-blue-50 border-blue-500 text-blue-600'
                    : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50',
                  'relative inline-flex items-center px-4 py-2 border text-sm font-medium',
                ]"
              >
                {{ page }}
              </button>

              <button
                @click="stocksStore.changePage(stocksStore.currentPage + 1)"
                :disabled="stocksStore.currentPage >= stocksStore.totalPages"
                class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronRightIcon class="h-5 w-5" />
              </button>
            </nav>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  MagnifyingGlassIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ArrowUpIcon,
  ArrowDownIcon,
} from '@heroicons/vue/24/outline'
import { useStocksStore } from '@/stores/stocks'
import type { RatingsFilters } from '@/types'
import StockLogo from '@/components/StockLogo.vue'
import MiniChart from '@/components/MiniChart.vue'

// Store
const stocksStore = useStocksStore()

// Local state
const searchQuery = ref('')
const sortBy = ref<RatingsFilters['sort_by']>('updated_at')
const sortOrder = ref<RatingsFilters['order']>('desc')
const pageSize = ref(20)

// Debounced search
let searchTimeout: ReturnType<typeof setTimeout> | null = null
const debouncedSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    stocksStore.searchRatings(searchQuery.value)
  }, 300)
}

// Computed
const visiblePages = computed(() => {
  const current = stocksStore.currentPage
  const total = stocksStore.totalPages
  const delta = 2

  const pages = []
  const start = Math.max(1, current - delta)
  const end = Math.min(total, current + delta)

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }

  return pages
})

const recentUpgrades = computed(() => {
  return stocksStore.ratings
    .filter((rating) => rating.rating_to && rating.rating_to.toLowerCase().includes('buy'))
    .sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime())
})

// Methods
const getRatingColor = (rating: string) => {
  const r = rating.toLowerCase()
  if (r.includes('buy') || r.includes('strong')) {
    return 'bg-green-100 text-green-800'
  } else if (r.includes('sell')) {
    return 'bg-red-100 text-red-800'
  } else if (r.includes('hold')) {
    return 'bg-yellow-100 text-yellow-800'
  } else {
    return 'bg-gray-100 text-gray-800'
  }
}

const getTargetTrendIcon = (rating: { target_to?: number; target_from?: number }) => {
  // Compare target_to with target_from to determine trend
  const targetTo = rating.target_to || 0
  const targetFrom = rating.target_from || 0

  if (targetTo > targetFrom) {
    return ArrowUpIcon
  } else if (targetTo < targetFrom) {
    return ArrowDownIcon
  } else {
    return ArrowUpIcon // Default to up if no previous target
  }
}

const getTargetTrendColor = (rating: { target_to?: number; target_from?: number }) => {
  const targetTo = rating.target_to || 0
  const targetFrom = rating.target_from || 0

  if (targetTo > targetFrom) {
    return 'text-green-600'
  } else if (targetTo < targetFrom) {
    return 'text-red-600'
  } else {
    return 'text-gray-400'
  }
}

const formatDate = (dateString: string) => {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
  }).format(new Date(dateString))
}

const handleSort = () => {
  stocksStore.sortRatings(sortBy.value, sortOrder.value)
}

const handlePageSizeChange = () => {
  stocksStore.changePageSize(pageSize.value)
}

// Lifecycle
onMounted(() => {
  if (stocksStore.ratings.length === 0) {
    stocksStore.fetchRatings()
  }
  if (stocksStore.recommendations.length === 0) {
    stocksStore.fetchRecommendations()
  }
})
</script>
