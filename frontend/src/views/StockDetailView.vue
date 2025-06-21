<template>
  <div class="min-h-screen bg-white">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <!-- Breadcrumb -->
      <nav class="flex mb-4" aria-label="Breadcrumb">
        <ol class="inline-flex items-center space-x-1 md:space-x-3">
          <li class="inline-flex items-center">
            <router-link to="/" class="text-gray-500 hover:text-gray-700 text-sm">
              Markets
            </router-link>
          </li>
          <li>
            <div class="flex items-center">
              <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
                  clip-rule="evenodd"
                ></path>
              </svg>
              <span class="text-gray-500 text-sm ml-1 md:ml-2">{{ ticker }} Price</span>
            </div>
          </li>
        </ol>
      </nav>

      <!-- Main Content Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Left Column - Stock Information -->
        <div class="lg:col-span-1 space-y-6">
          <!-- Stock Header -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div class="flex items-center space-x-3 mb-4">
              <StockLogo :symbol="ticker" size="md" />
              <div>
                <h1 class="text-2xl font-bold text-gray-900">{{ ticker }}</h1>
                <p class="text-gray-500 text-sm">{{ companyName || 'Stock' }}</p>
              </div>
            </div>

            <!-- Current Price -->
            <div class="mb-4">
              <div class="text-3xl font-bold text-gray-900 mb-1">
                ${{ lastPrice?.toFixed(2) || 'N/A' }}
                <span v-if="priceChange !== 0" :class="priceChangeClass" class="text-base ml-2">
                  {{ priceChange >= 0 ? '▲' : '▼' }} {{ Math.abs(priceChangePercent).toFixed(2) }}%
                </span>
              </div>

              <!-- Price Range Indicator -->
              <div v-if="lastPrice" class="mb-4">
                <div class="flex justify-between text-sm text-gray-600 mb-1">
                  <span>${{ (lastPrice * 0.95).toFixed(2) }}</span>
                  <span class="text-xs text-gray-500">24h Range</span>
                  <span>${{ (lastPrice * 1.05).toFixed(2) }}</span>
                </div>
                <div
                  class="w-full bg-gradient-to-r from-red-300 via-yellow-300 to-green-300 rounded-full h-2"
                >
                  <div
                    class="bg-gray-800 w-1 h-4 rounded-full relative -top-1"
                    style="margin-left: 60%"
                  ></div>
                </div>
              </div>
            </div>
          </div>

          <!-- Key Metrics -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div class="grid grid-cols-1 gap-4">
              <!-- Analyst Consensus -->
              <div class="flex justify-between py-2 border-b border-gray-100">
                <span class="text-gray-600 text-sm">Analyst Consensus</span>
                <span class="font-medium text-gray-900">
                  {{ getMostCommonRating() }}
                </span>
              </div>

              <!-- Average Price Target -->
              <div class="flex justify-between py-2 border-b border-gray-100">
                <span class="text-gray-600 text-sm">Avg Price Target</span>
                <span class="font-medium text-gray-900">
                  ${{ averagePriceTarget.toFixed(2) }}
                </span>
              </div>

              <!-- Number of Analysts -->
              <div class="flex justify-between py-2 border-b border-gray-100">
                <span class="text-gray-600 text-sm">Analyst Coverage</span>
                <span class="font-medium text-gray-900"> {{ sortedRatings.length }} ratings </span>
              </div>

              <!-- Analyst Firms -->
              <div class="flex justify-between py-2 border-b border-gray-100">
                <span class="text-gray-600 text-sm">Analyst Firms</span>
                <span class="font-medium text-gray-900"> {{ uniqueFirms.length }} firms </span>
              </div>

              <!-- Last Update -->
              <div class="flex justify-between py-2">
                <span class="text-gray-600 text-sm">Last Updated</span>
                <span class="font-medium text-gray-900">
                  {{ formatDate(lastRatingDate) }}
                </span>
              </div>
            </div>
          </div>

          <!-- Rating Distribution -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">Rating Breakdown</h3>
            <div class="space-y-3">
              <div
                v-for="(count, rating) in ratingDistribution"
                :key="rating"
                class="flex items-center justify-between"
              >
                <div class="flex items-center space-x-3 flex-1">
                  <div class="w-3 h-3 rounded-full" :class="getRatingDotColor(rating)"></div>
                  <span class="text-sm text-gray-700">{{ rating }}</span>
                </div>
                <div class="flex items-center space-x-2">
                  <div class="w-16 bg-gray-200 rounded-full h-2">
                    <div
                      class="h-2 rounded-full"
                      :class="getRatingBarColor(rating)"
                      :style="{ width: `${(count / sortedRatings.length) * 100}%` }"
                    ></div>
                  </div>
                  <span class="text-sm font-medium text-gray-900 w-6 text-right">{{ count }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Recent Analyst Activity -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h3>
            <div class="space-y-3">
              <div
                v-for="rating in sortedRatings.slice(0, 5)"
                :key="rating.rating_id"
                class="flex items-center justify-between py-2 border-b border-gray-100 last:border-b-0"
              >
                <div>
                  <div class="text-sm font-medium text-gray-900">{{ rating.brokerage }}</div>
                  <div class="text-xs text-gray-500">{{ formatDate(rating.time) }}</div>
                </div>
                <div class="text-right">
                  <!-- Show rating change if different -->
                  <div
                    v-if="rating.rating_from && rating.rating_from !== rating.rating_to"
                    class="flex items-center justify-end space-x-2 mb-2"
                  >
                    <div class="text-right opacity-60">
                      <span
                        class="inline-flex px-2 py-1 text-xs font-medium rounded-lg bg-gray-100 text-gray-500 border line-through"
                      >
                        {{ rating.rating_from }}
                      </span>
                      <div class="text-xs text-gray-400 mt-1">
                        ${{ (rating.target_from || 0).toFixed(2) }}
                      </div>
                    </div>
                    <div class="text-blue-500 text-sm font-bold animate-pulse">→</div>
                  </div>

                  <!-- Current rating -->
                  <span
                    class="inline-flex px-3 py-1.5 text-xs font-bold rounded-lg shadow-md border-2 border-opacity-20"
                    :class="getRatingBadgeClass(rating.rating_to)"
                  >
                    {{ rating.rating_to }}
                  </span>
                  <div class="text-xs text-gray-700 mt-1 bg-blue-50 px-2 py-1 rounded-md">
                    ${{ (rating.target_to || 0).toFixed(2) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Column - Chart -->
        <div class="lg:col-span-2">
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
            <!-- Chart Header -->
            <div class="flex items-center justify-between mb-6">
              <div>
                <h2 class="text-xl font-semibold text-gray-900">Price Chart</h2>
                <p class="text-sm text-gray-500">Historical price movement</p>
              </div>

              <!-- Time Period Selector -->
              <div
                class="flex items-center space-x-1 bg-white border border-gray-200 rounded-lg p-1"
              >
                <button
                  v-for="period in chartPeriods"
                  :key="period.value"
                  @click="selectedPeriod = period.value"
                  :class="[
                    selectedPeriod === period.value
                      ? 'bg-white text-gray-900 shadow-sm'
                      : 'text-gray-600 hover:text-gray-900',
                    'px-3 py-1 text-sm font-medium rounded-md transition-colors',
                  ]"
                >
                  {{ period.label }}
                </button>
              </div>
            </div>

            <!-- Chart Component -->
            <div class="h-96">
              <StockChart :symbol="ticker" :period="selectedPeriod" />
            </div>
          </div>

          <!-- All Analyst Ratings Table -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
            <div class="px-6 py-4 border-b border-gray-200">
              <h2 class="text-xl font-semibold text-gray-900">All Analyst Ratings</h2>
              <p class="text-sm text-gray-500 mt-1">
                Complete history of analyst ratings and price targets
              </p>
            </div>

            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200">
                <thead class="border-t border-b border-gray-200">
                  <tr>
                    <th
                      class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Date
                    </th>
                    <th
                      class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Analyst Firm
                    </th>
                    <th
                      class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Rating
                    </th>
                    <th
                      class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Price Target
                    </th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                  <tr
                    v-for="rating in sortedRatings"
                    :key="rating.rating_id"
                    class="hover:bg-gray-50"
                  >
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {{ formatDate(rating.time) }}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                      <div class="text-sm font-medium text-gray-900">{{ rating.brokerage }}</div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-right">
                      <div class="flex items-center justify-end space-x-3">
                        <!-- Previous Rating (if different) -->
                        <div
                          v-if="rating.rating_from && rating.rating_from !== rating.rating_to"
                          class="opacity-60 text-right"
                        >
                          <span
                            class="inline-flex px-2 py-1 text-xs font-medium rounded-lg bg-gray-100 text-gray-500 border line-through"
                          >
                            {{ rating.rating_from }}
                          </span>
                        </div>

                        <!-- Animated Arrow (if there's a change) -->
                        <div
                          v-if="rating.rating_from && rating.rating_from !== rating.rating_to"
                          class="text-blue-500 text-lg font-bold animate-pulse"
                        >
                          →
                        </div>

                        <!-- Current Rating -->
                        <span
                          class="inline-flex px-3 py-1.5 text-xs font-bold rounded-lg shadow-md border-2 border-opacity-20"
                          :class="getRatingBadgeClass(rating.rating_to)"
                        >
                          {{ rating.rating_to }}
                        </span>
                      </div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 text-right">
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
                        </div>

                        <!-- Animated Arrow (if there's a change) -->
                        <div
                          v-if="rating.target_from && rating.target_from !== rating.target_to"
                          class="text-blue-500 text-lg font-bold animate-pulse"
                        >
                          →
                        </div>

                        <!-- Current Price Target -->
                        <div
                          class="text-xs font-bold text-gray-900 bg-blue-50 px-3 py-1.5 rounded-lg border-2 border-blue-200 shadow-sm"
                        >
                          ${{ (rating.target_to || 0).toFixed(2) }}
                        </div>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-if="sortedRatings.length === 0" class="text-center py-12">
              <svg
                class="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
              <h3 class="mt-2 text-sm font-medium text-gray-900">No ratings available</h3>
              <p class="mt-1 text-sm text-gray-500">
                There are currently no analyst ratings for this stock.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/* eslint-disable vue/no-side-effects-in-computed-properties */
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useStocksStore } from '../stores/stocks'
import type { StockRating } from '../types'
import StockChart from '../components/StockChart.vue'
import StockLogo from '../components/StockLogo.vue'

const route = useRoute()
const stocksStore = useStocksStore()

const ticker = computed(() => (route.params.ticker as string) || '')

// State for stock-specific ratings
const stockRatings = ref<StockRating[]>([])

// Computed properties based on current data structure
const sortedRatings = computed(() => {
  return stockRatings.value.sort(
    (a: StockRating, b: StockRating) => new Date(b.time).getTime() - new Date(a.time).getTime(),
  )
})

const companyName = computed(() => {
  if (sortedRatings.value.length > 0) {
    return sortedRatings.value[0].company
  }
  return ticker.value
})

const averagePriceTarget = computed(() => {
  if (sortedRatings.value.length === 0) return 0
  const validTargets = sortedRatings.value
    .map((r: StockRating) => r.target_to)
    .filter((target: number | null | undefined) => target && target > 0) as number[]

  if (validTargets.length === 0) return 0
  return validTargets.reduce((sum: number, target: number) => sum + target, 0) / validTargets.length
})

const uniqueFirms = computed(() => {
  return [...new Set(sortedRatings.value.map((r: StockRating) => r.brokerage))]
})

const lastRatingDate = computed(() => {
  if (sortedRatings.value.length === 0) return null
  return sortedRatings.value[0].time
})

const ratingDistribution = computed(() => {
  const distribution: Record<string, number> = {}
  sortedRatings.value.forEach((rating: StockRating) => {
    const ratingValue = rating.rating_to || 'Unknown'
    distribution[ratingValue] = (distribution[ratingValue] || 0) + 1
  })
  return distribution
})

// Real price data from API
const lastPrice = ref<number | null>(null)
const priceChange = ref<number>(0)

// Chart period selection
const selectedPeriod = ref('1M')
const chartPeriods = [
  { label: '1W', value: '1W' },
  { label: '1M', value: '1M' },
  { label: '3M', value: '3M' },
  { label: '6M', value: '6M' },
  { label: '1Y', value: '1Y' },
]

const priceChangePercent = computed(() => {
  const price = lastPrice.value
  const change = priceChange.value
  if (!price || price <= 0) {
    return 0
  }
  return (change / price) * 100
})

const priceChangeClass = computed(() =>
  priceChange.value >= 0 ? 'text-green-600 font-medium' : 'text-red-600 font-medium',
)

// Load real price data
const loadPriceData = async () => {
  const currentTicker = route.params.ticker as string
  if (!currentTicker) return

  try {
    const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    const response = await fetch(`${baseURL}/api/v1/stocks/${currentTicker}/price?period=1M`)
    if (!response.ok) {
      throw new Error('Failed to load price data')
    }

    const data = await response.json()
    const bars = data.bars || []

    if (bars.length >= 2) {
      const latestBar = bars[bars.length - 1]
      const previousBar = bars[bars.length - 2]

      lastPrice.value = latestBar.close
      priceChange.value = latestBar.close - previousBar.close
    } else if (bars.length === 1) {
      lastPrice.value = bars[0].close
      priceChange.value = 0
    }
  } catch (error) {
    console.error('Failed to load price data:', error)
    lastPrice.value = null
    priceChange.value = 0
  }
}

// Methods
const formatDate = (dateString: string | null) => {
  if (!dateString) return 'N/A'

  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) return 'Invalid Date'

    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  } catch (error) {
    console.error('Date formatting error:', error)
    return 'Invalid Date'
  }
}

const getRatingBadgeClass = (rating: string | null) => {
  if (!rating) return 'bg-gray-100 text-gray-800'

  const ratingLower = rating.toLowerCase()
  if (ratingLower.includes('buy')) return 'bg-green-100 text-green-800'
  if (ratingLower.includes('hold') || ratingLower.includes('neutral'))
    return 'bg-yellow-100 text-yellow-800'
  if (ratingLower.includes('sell')) return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}

const getMostCommonRating = () => {
  if (sortedRatings.value.length === 0) return 'N/A'

  const ratingCounts: Record<string, number> = {}
  sortedRatings.value.forEach((rating) => {
    const ratingValue = rating.rating_to || 'Unknown'
    ratingCounts[ratingValue] = (ratingCounts[ratingValue] || 0) + 1
  })

  let mostCommon = 'N/A'
  let maxCount = 0

  Object.entries(ratingCounts).forEach(([rating, count]) => {
    if (count > maxCount) {
      maxCount = count
      mostCommon = rating
    }
  })

  return mostCommon
}

const getRatingDotColor = (rating: string) => {
  const ratingLower = rating.toLowerCase()
  if (ratingLower.includes('buy')) return 'bg-green-500'
  if (ratingLower.includes('hold') || ratingLower.includes('neutral')) return 'bg-yellow-500'
  if (ratingLower.includes('sell')) return 'bg-red-500'
  return 'bg-gray-500'
}

const getRatingBarColor = (rating: string) => {
  const ratingLower = rating.toLowerCase()
  if (ratingLower.includes('buy')) return 'bg-green-500'
  if (ratingLower.includes('hold') || ratingLower.includes('neutral')) return 'bg-yellow-500'
  if (ratingLower.includes('sell')) return 'bg-red-500'
  return 'bg-gray-500'
}

onMounted(async () => {
  const currentTicker = route.params.ticker as string
  if (currentTicker) {
    stockRatings.value = await stocksStore.fetchRatingsByTicker(currentTicker)
    loadPriceData()
  }
})
</script>
