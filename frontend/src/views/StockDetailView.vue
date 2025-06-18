<template>
  <div class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <!-- Stock Header -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <StockLogo :symbol="ticker" size="lg" />
            <div>
              <h1 class="text-3xl font-bold text-gray-900">{{ ticker }}</h1>
              <p class="text-gray-500">{{ companyName || 'Stock Analysis' }}</p>
            </div>
          </div>

          <div class="text-right">
            <div class="text-2xl font-bold text-gray-900">
              ${{ lastPrice?.toFixed(2) || 'N/A' }}
            </div>
            <div :class="priceChangeClass">
              {{ priceChange >= 0 ? '+' : '' }}{{ priceChange?.toFixed(2) || '0.00' }} ({{
                priceChangePercent >= 0 ? '+' : ''
              }}{{ priceChangePercent?.toFixed(2) || '0.00' }}%)
            </div>
          </div>
        </div>
      </div>

      <!-- Chart Section -->
      <div class="mb-6">
        <StockChart :symbol="ticker" />
      </div>

      <!-- Stats Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
        <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <div class="text-sm font-medium text-gray-500 mb-1">Total Ratings</div>
          <div class="text-2xl font-bold text-gray-900">{{ sortedRatings.length }}</div>
        </div>

        <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <div class="text-sm font-medium text-gray-500 mb-1">Avg Price Target</div>
          <div class="text-2xl font-bold text-gray-900">${{ averagePriceTarget.toFixed(2) }}</div>
        </div>

        <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <div class="text-sm font-medium text-gray-500 mb-1">Analyst Firms</div>
          <div class="text-2xl font-bold text-gray-900">{{ uniqueFirms.length }}</div>
        </div>

        <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <div class="text-sm font-medium text-gray-500 mb-1">Last Update</div>
          <div class="text-lg font-semibold text-gray-900">
            {{ formatDate(lastRatingDate) }}
          </div>
        </div>
      </div>

      <!-- Rating Distribution -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <h2 class="text-xl font-semibold text-gray-900 mb-4">Rating Distribution</h2>
        <div class="space-y-3">
          <div
            v-for="(count, rating) in ratingDistribution"
            :key="rating"
            class="flex items-center"
          >
            <div class="w-20 text-sm font-medium text-gray-600">{{ rating }}</div>
            <div class="flex-1 bg-gray-200 rounded-full h-3 mx-3">
              <div
                class="h-3 rounded-full transition-all duration-300"
                :class="getRatingColor(rating)"
                :style="{ width: `${(count / sortedRatings.length) * 100}%` }"
              ></div>
            </div>
            <div class="w-12 text-sm font-semibold text-gray-900 text-right">{{ count }}</div>
          </div>
        </div>
      </div>

      <!-- Recent Ratings Table -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-xl font-semibold text-gray-900">Recent Analyst Ratings</h2>
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Date
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Firm
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Action
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Rating
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  Price Target
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="rating in sortedRatings" :key="rating.rating_id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatDate(rating.time) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-medium text-gray-900">{{ rating.brokerage }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span
                    class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800"
                  >
                    {{ rating.action || 'Updated' }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span
                    class="inline-flex px-2 py-1 text-xs font-semibold rounded-full"
                    :class="getRatingBadgeClass(rating.rating_to)"
                  >
                    {{ rating.rating_to }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  ${{ (rating.target_to || 0).toFixed(2) }}
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
    const response = await fetch(
      `http://localhost:8080/api/v1/stocks/${currentTicker}/price?period=1M`,
    )
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

const getRatingColor = (rating: string) => {
  const ratingLower = rating.toLowerCase()
  if (ratingLower.includes('buy') || ratingLower.includes('strong buy')) return 'bg-green-500'
  if (ratingLower.includes('hold') || ratingLower.includes('neutral')) return 'bg-yellow-500'
  if (ratingLower.includes('sell')) return 'bg-red-500'
  return 'bg-gray-500'
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

onMounted(async () => {
  const currentTicker = route.params.ticker as string
  if (currentTicker) {
    stockRatings.value = await stocksStore.fetchRatingsByTicker(currentTicker)
    loadPriceData()
  }
})
</script>
