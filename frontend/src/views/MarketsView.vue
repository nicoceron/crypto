<template>
  <div class="space-y-4">
    <!-- Page header with total market stats -->
    <div class="bg-white p-4">
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
    </div>

    <!-- Trending and Top Gainers section -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- Trending (Recommendations) -->
      <div class="bg-white shadow rounded-lg p-3 flex flex-col">
        <div class="flex items-center justify-between mb-2">
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

        <div v-else class="space-y-4 flex-1">
          <div
            v-for="rec in stocksStore.topRecommendations.slice(0, 2)"
            :key="rec.ticker"
            class="flex items-center justify-between p-5 hover:bg-gray-50 rounded-lg cursor-pointer border border-gray-100"
            @click="$router.push(`/stock/${rec.ticker}`)"
          >
            <div class="flex items-center space-x-3">
              <StockLogo :symbol="rec.ticker" size="sm" />
              <div>
                <div class="text-sm font-medium text-gray-900">{{ rec.ticker }}</div>
                <div class="text-xs text-gray-500">{{ rec.company || rec.ticker }}</div>
                <div class="text-xs text-gray-400 mt-1">
                  Score {{ (rec.score || 0).toFixed(1) }}
                </div>
              </div>
            </div>
            <div class="flex items-center space-x-4">
              <div class="text-right">
                <div class="text-lg font-bold text-gray-900">
                  ${{ (rec.target_price || 0).toFixed(2) }}
                </div>
                <div class="text-xs text-gray-500">Target Price</div>
              </div>
              <div class="w-48 h-12">
                <MiniChart :symbol="rec.ticker" period="1W" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Top Gainers (from ratings) -->
      <div class="bg-white shadow rounded-lg p-3 flex flex-col">
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-lg font-semibold text-gray-900 flex items-center">🚀 Recent Upgrades</h3>
          <button class="text-blue-600 hover:underline text-sm">View more →</button>
        </div>

        <div v-if="recentUpgrades.length === 0" class="text-center py-4">
          <div class="text-gray-500">Loading recent upgrades...</div>
        </div>

        <div v-else class="space-y-2 flex-1">
          <div
            v-for="rating in recentUpgrades.slice(0, 3)"
            :key="rating.rating_id"
            class="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg cursor-pointer border border-gray-100"
            @click="$router.push(`/stock/${rating.ticker}`)"
          >
            <div class="flex items-center space-x-3">
              <StockLogo :symbol="rating.ticker" size="sm" />
              <div>
                <div class="text-sm font-medium text-gray-900">{{ rating.ticker }}</div>
                <div class="text-xs text-gray-500">{{ rating.brokerage }}</div>
              </div>
            </div>
            <div class="flex items-center space-x-4">
              <!-- Price Target Section -->
              <div class="text-right">
                <div class="text-lg font-bold text-gray-900">
                  ${{ (rating.target_to || 0).toFixed(2) }}
                </div>
                <div class="text-xs text-gray-500">Price Target</div>
              </div>

              <!-- Rating Change Section -->
              <div class="flex items-center space-x-2">
                <div
                  v-if="rating.rating_from && rating.rating_from !== rating.rating_to"
                  class="flex items-center space-x-2"
                >
                  <span
                    class="inline-flex px-2 py-1 text-xs font-medium rounded-lg bg-gray-100 text-gray-500 border line-through"
                  >
                    {{ rating.rating_from }}
                  </span>
                  <div class="text-blue-500 text-sm font-bold">→</div>
                </div>
                <span
                  class="inline-flex px-3 py-1.5 text-xs font-bold rounded-lg bg-green-100 text-green-800 border border-green-200"
                >
                  {{ rating.rating_to }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Main table -->
    <div class="bg-white">
      <!-- Search and filters -->
      <div class="p-4">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <!-- Search -->
          <div>
            <div class="relative">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search by ticker, company..."
                class="block w-full pr-10 border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 text-sm font-medium text-gray-900 placeholder-gray-500"
                @input="debouncedSearch"
              />
              <div class="absolute inset-y-0 right-0 pr-3 flex items-center">
                <MagnifyingGlassIcon class="h-5 w-5 text-gray-400" />
              </div>
            </div>
          </div>

          <!-- Page size -->
          <div>
            <select
              v-model="pageSize"
              class="block w-full border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 text-sm font-medium text-gray-900"
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
          <thead class="border-t border-b border-gray-200">
            <tr>
              <th
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-12"
              >
                #
              </th>
              <th
                @click="handleHeaderClick('ticker')"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 select-none"
              >
                <div class="flex items-center justify-between">
                  <span>Stock</span>
                  <div class="flex flex-col ml-1">
                    <component
                      :is="getSortIcon('ticker', 'asc')"
                      class="h-3 w-3"
                      :class="getSortIconColor('ticker', 'asc')"
                    />
                    <component
                      :is="getSortIcon('ticker', 'desc')"
                      class="h-3 w-3 -mt-1"
                      :class="getSortIconColor('ticker', 'desc')"
                    />
                  </div>
                </div>
              </th>
              <th
                @click="handleHeaderClick('target_to')"
                class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 select-none"
              >
                <div class="flex items-center justify-end">
                  <span>Price Target</span>
                  <div class="flex flex-col ml-1">
                    <component
                      :is="getSortIcon('target_to', 'asc')"
                      class="h-3 w-3"
                      :class="getSortIconColor('target_to', 'asc')"
                    />
                    <component
                      :is="getSortIcon('target_to', 'desc')"
                      class="h-3 w-3 -mt-1"
                      :class="getSortIconColor('target_to', 'desc')"
                    />
                  </div>
                </div>
              </th>
              <th
                @click="handleHeaderClick('rating_to')"
                class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 select-none"
              >
                <div class="flex items-center justify-end">
                  <span>Rating</span>
                  <div class="flex flex-col ml-1">
                    <component
                      :is="getSortIcon('rating_to', 'asc')"
                      class="h-3 w-3"
                      :class="getSortIconColor('rating_to', 'asc')"
                    />
                    <component
                      :is="getSortIcon('rating_to', 'desc')"
                      class="h-3 w-3 -mt-1"
                      :class="getSortIconColor('rating_to', 'desc')"
                    />
                  </div>
                </div>
              </th>
              <th
                @click="handleHeaderClick('brokerage')"
                class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 select-none"
              >
                <div class="flex items-center justify-end">
                  <span>Analyst</span>
                  <div class="flex flex-col ml-1">
                    <component
                      :is="getSortIcon('brokerage', 'asc')"
                      class="h-3 w-3"
                      :class="getSortIconColor('brokerage', 'asc')"
                    />
                    <component
                      :is="getSortIcon('brokerage', 'desc')"
                      class="h-3 w-3 -mt-1"
                      :class="getSortIconColor('brokerage', 'desc')"
                    />
                  </div>
                </div>
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
              >
                Last 7 Days
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-100">
            <tr
              v-for="(rating, index) in sortedRatings"
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

              <!-- Price Target Change -->
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="flex items-center justify-end space-x-3">
                  <!-- Previous Price Target (if different) -->
                  <div
                    v-if="rating.target_from && rating.target_from !== rating.target_to"
                    class="text-right opacity-60"
                  >
                    <div
                      class="text-xs font-medium text-gray-500 bg-gray-50 px-2 py-1 rounded-lg border line-through"
                    >
                      ${{ (rating.target_from || 0).toFixed(2) }}
                    </div>
                    <div class="text-xs text-gray-400 mt-1">Previous</div>
                  </div>

                  <!-- Animated Arrow (if there's a change) -->
                  <div
                    v-if="rating.target_from && rating.target_from !== rating.target_to"
                    class="text-blue-500 text-lg font-bold animate-pulse"
                  >
                    →
                  </div>

                  <!-- Current Price Target -->
                  <div class="text-right">
                    <div
                      class="text-xs font-bold text-gray-900 bg-blue-50 px-3 py-1.5 rounded-lg border-2 border-blue-200"
                    >
                      ${{ (rating.target_to || 0).toFixed(2) }}
                    </div>
                    <div class="text-xs text-gray-600 mt-1">Current</div>
                  </div>

                  <!-- Enhanced Trend indicator -->
                  <div class="flex items-center">
                    <component
                      :is="getTargetTrendIcon(rating)"
                      class="h-5 w-5"
                      :class="getTargetTrendColor(rating) + ' drop-shadow-sm'"
                    />
                  </div>
                </div>
              </td>

              <!-- Rating -->
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="flex items-center justify-end space-x-3">
                  <!-- Previous Rating (if different) -->
                  <div
                    v-if="rating.rating_from && rating.rating_from !== rating.rating_to"
                    class="text-right opacity-60"
                  >
                    <span
                      class="inline-flex px-2 py-1 text-xs font-medium rounded-lg bg-gray-100 text-gray-500 border line-through"
                    >
                      {{ rating.rating_from }}
                    </span>
                    <div class="text-xs text-gray-400 mt-1">
                      ${{ (rating.target_from || 0).toFixed(2) }}
                    </div>
                  </div>

                  <!-- Animated Arrow (if there's a change) -->
                  <div
                    v-if="rating.rating_from && rating.rating_from !== rating.rating_to"
                    class="text-blue-500 text-lg font-bold animate-pulse"
                  >
                    →
                  </div>

                  <!-- Current Rating -->
                  <div class="text-right">
                    <span
                      :class="getRatingColor(rating.rating_to)"
                      class="inline-flex px-3 py-1.5 text-xs font-bold rounded-lg shadow-sm border-2 border-opacity-20"
                    >
                      {{ rating.rating_to }}
                    </span>
                    <div class="text-xs text-gray-700 mt-1">
                      ${{ (rating.target_to || 0).toFixed(2) }}
                    </div>
                  </div>
                </div>
              </td>

              <!-- Analyst -->
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="text-sm text-gray-900">{{ rating.brokerage }}</div>
                <div class="text-xs text-gray-500">{{ formatDate(rating.time) }}</div>
              </td>

              <!-- Mini chart -->
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="w-32 h-8">
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
        class="flex items-center justify-between bg-white px-4 py-3 sm:px-6"
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
  MinusIcon,
} from '@heroicons/vue/24/outline'
import { ChevronUpIcon, ChevronDownIcon } from '@heroicons/vue/20/solid'
import { useStocksStore } from '@/stores/stocks'
import type { RatingsFilters } from '@/types'
import StockLogo from '@/components/StockLogo.vue'
import MiniChart from '@/components/MiniChart.vue'

// Store
const stocksStore = useStocksStore()

// Local state
const searchQuery = ref('')
const sortBy = ref<RatingsFilters['sort_by']>('time')
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
  return sortedRatings.value
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

  // Only show arrows if there's actually a change
  if (targetFrom && targetTo > targetFrom) {
    return ArrowUpIcon
  } else if (targetFrom && targetTo < targetFrom) {
    return ArrowDownIcon
  } else {
    return MinusIcon // Show flat line for no change or no previous target
  }
}

const getTargetTrendColor = (rating: { target_to?: number; target_from?: number }) => {
  const targetTo = rating.target_to || 0
  const targetFrom = rating.target_from || 0

  // Only show colors if there's actually a change
  if (targetFrom && targetTo > targetFrom) {
    return 'text-green-600'
  } else if (targetFrom && targetTo < targetFrom) {
    return 'text-red-600'
  } else {
    return 'text-gray-400' // Gray for no change or no previous target
  }
}

const formatDate = (dateString: string) => {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
  }).format(new Date(dateString))
}

// Rating priority mapping for proper sorting
const getRatingPriority = (rating: string): number => {
  const r = rating.toLowerCase()
  // Higher numbers = better ratings (for desc sort)
  if (r.includes('strong buy') || r.includes('outperform')) return 5
  if (r.includes('buy')) return 4
  if (r.includes('hold') || r.includes('neutral')) return 3
  if (r.includes('underweight') || r.includes('underperform')) return 2
  if (r.includes('sell') || r.includes('strong sell')) return 1
  return 0 // Unknown ratings get lowest priority
}

// Computed property for sorted ratings
const sortedRatings = computed(() => {
  if (sortBy.value === 'rating_to') {
    return [...stocksStore.ratings].sort((a, b) => {
      const priorityA = getRatingPriority(a.rating_to || '')
      const priorityB = getRatingPriority(b.rating_to || '')

      if (sortOrder.value === 'desc') {
        return priorityB - priorityA // Higher priority first
      } else {
        return priorityA - priorityB // Lower priority first
      }
    })
  } else {
    return stocksStore.ratings
  }
})

const handleSort = () => {
  // For rating sorting, we use the computed property above
  // For other fields, use backend sorting
  if (sortBy.value !== 'rating_to') {
    stocksStore.sortRatings(sortBy.value, sortOrder.value)
  }
  // For rating sorting, the computed property will handle it automatically
}

const handlePageSizeChange = () => {
  stocksStore.changePageSize(pageSize.value)
}

// Header click handler for sorting
const handleHeaderClick = (column: string) => {
  if (sortBy.value === column) {
    // Toggle order if same column
    sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
  } else {
    // New column, default to desc
    sortBy.value = column as RatingsFilters['sort_by']
    sortOrder.value = 'desc'
  }
  handleSort()
}

// Get sort icon for header
const getSortIcon = (column: string, direction: 'asc' | 'desc') => {
  if (sortBy.value === column && sortOrder.value === direction) {
    return direction === 'asc' ? ChevronUpIcon : ChevronDownIcon
  }
  return direction === 'asc' ? ChevronUpIcon : ChevronDownIcon
}

// Get sort icon color
const getSortIconColor = (column: string, direction: 'asc' | 'desc') => {
  if (sortBy.value === column && sortOrder.value === direction) {
    return 'text-blue-600'
  }
  return 'text-gray-300'
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
