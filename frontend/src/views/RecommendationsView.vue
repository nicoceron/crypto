<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="md:flex md:items-center md:justify-between">
      <div class="min-w-0 flex-1">
        <h2
          class="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight"
        >
          Stock Recommendations
        </h2>
        <p class="mt-1 text-sm text-gray-500">
          AI-generated investment recommendations based on analyst consensus and market data
        </p>
      </div>
    </div>

    <div v-if="stocksStore.isLoading" class="text-center py-12">
      <ArrowPathIcon class="mx-auto h-12 w-12 text-gray-400 animate-spin" />
      <p class="mt-4 text-lg text-gray-500">Loading recommendations...</p>
    </div>

    <div v-else-if="stocksStore.recommendations.length === 0" class="text-center py-12">
      <StarIcon class="mx-auto h-16 w-16 text-gray-400" />
      <h3 class="mt-4 text-xl font-medium text-gray-900">No recommendations available</h3>
      <p class="mt-2 text-gray-500">
        Recommendations will appear here once the analysis engine processes the latest data.
      </p>
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-2">
      <div
        v-for="recommendation in sortedRecommendations"
        :key="recommendation.ticker"
        @click="navigateToStock(recommendation.ticker)"
        class="bg-white shadow rounded-lg overflow-hidden hover:shadow-lg transition-shadow cursor-pointer"
      >
        <div class="p-6">
          <!-- Header -->
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center space-x-3">
              <div class="flex-shrink-0">
                <StockLogo :symbol="recommendation.ticker" size="md" />
              </div>
              <div>
                <h3 class="text-lg font-medium text-gray-900">
                  {{ recommendation.company || recommendation.ticker }}
                </h3>
                <p class="text-sm text-gray-500">
                  Latest Rating: {{ recommendation.latest_rating || 'N/A' }}
                </p>
              </div>
            </div>

            <!-- Score badge -->
            <div class="flex-shrink-0">
              <div
                :class="getScoreColor(recommendation.score || 0)"
                class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium"
              >
                <TrophyIcon class="w-4 h-4 mr-1" />
                {{ (recommendation.score || 0).toFixed(1) }}
              </div>
            </div>
          </div>

          <!-- Recommendation reason -->
          <div class="mb-4">
            <h4 class="text-sm font-medium text-gray-900 mb-2">Investment Thesis</h4>
            <p class="text-sm text-gray-600">
              {{ recommendation.rationale || 'No rationale provided' }}
            </p>
          </div>

          <!-- Details grid -->
          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="bg-gray-50 rounded-lg p-3">
              <div class="text-xs font-medium text-gray-500 uppercase tracking-wide">
                Latest Rating
              </div>
              <div class="mt-1 text-sm font-medium text-gray-900">
                {{ recommendation.latest_rating || 'N/A' }}
              </div>
            </div>

            <div class="bg-gray-50 rounded-lg p-3">
              <div class="text-xs font-medium text-gray-500 uppercase tracking-wide">
                Technical Signal
              </div>
              <div class="mt-1 text-sm font-medium text-gray-900">
                {{ recommendation.technical_signal || 'N/A' }}
              </div>
            </div>

            <div class="bg-gray-50 rounded-lg p-3">
              <div class="text-xs font-medium text-gray-500 uppercase tracking-wide">
                Price Target
              </div>
              <div class="mt-1 text-sm font-medium text-gray-900">
                ${{ (recommendation.target_price || 0).toFixed(2) }}
              </div>
            </div>

            <div class="bg-gray-50 rounded-lg p-3">
              <div class="text-xs font-medium text-gray-500 uppercase tracking-wide">Sentiment</div>
              <div class="mt-1 text-sm font-medium text-gray-900">
                {{ recommendation.sentiment_score || 'N/A' }}
              </div>
            </div>
          </div>

          <!-- Additional details -->
          <div
            v-if="recommendation.technical_signal || recommendation.sentiment_score"
            class="space-y-2"
          >
            <div
              v-if="
                recommendation.technical_signal &&
                recommendation.technical_signal !== 'Pending Analysis'
              "
              class="flex items-start space-x-2"
            >
              <ChartBarIcon class="w-4 h-4 text-blue-500 mt-0.5 flex-shrink-0" />
              <div>
                <div class="text-xs font-medium text-gray-500 uppercase tracking-wide">
                  Technical Signal
                </div>
                <div class="text-sm text-gray-600">
                  {{ recommendation.technical_signal }}
                </div>
              </div>
            </div>

            <div v-if="recommendation.sentiment_score" class="flex items-start space-x-2">
              <HeartIcon class="w-4 h-4 text-purple-500 mt-0.5 flex-shrink-0" />
              <div>
                <div class="text-xs font-medium text-gray-500 uppercase tracking-wide">
                  Sentiment Score
                </div>
                <div class="text-sm text-gray-600">
                  {{ recommendation.sentiment_score }}
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="mt-4 pt-4 border-t border-gray-200">
            <div class="flex items-center justify-between text-xs text-gray-500">
              <span>Generated {{ formatDate(recommendation.generated_at) }}</span>
              <div class="flex items-center space-x-1">
                <ClockIcon class="w-3 h-3" />
                <span>{{ getTimeAgo(recommendation.generated_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Summary stats -->
    <div v-if="stocksStore.recommendations.length > 0" class="bg-white shadow rounded-lg p-6">
      <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Recommendation Summary</h3>

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div class="text-center">
          <div class="text-2xl font-bold text-blue-600">
            {{ stocksStore.recommendations.length }}
          </div>
          <div class="text-sm text-gray-500">Total Recommendations</div>
        </div>

        <div class="text-center">
          <div class="text-2xl font-bold text-green-600">
            {{ averageScore.toFixed(1) }}
          </div>
          <div class="text-sm text-gray-500">Average Score</div>
        </div>

        <div class="text-center">
          <div class="text-2xl font-bold text-purple-600">
            {{ topScore.toFixed(1) }}
          </div>
          <div class="text-sm text-gray-500">Highest Score</div>
        </div>

        <div class="text-center">
          <div class="text-2xl font-bold text-orange-600">
            {{ uniqueRatings }}
          </div>
          <div class="text-sm text-gray-500">Rating Types</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  StarIcon,
  ArrowPathIcon,
  TrophyIcon,
  ChartBarIcon,
  HeartIcon,
  ClockIcon,
} from '@heroicons/vue/24/outline'
import { useStocksStore } from '@/stores/stocks'
import StockLogo from '@/components/StockLogo.vue'

// Store and router
const stocksStore = useStocksStore()
const router = useRouter()

// Computed
const sortedRecommendations = computed(() => {
  return [...stocksStore.recommendations].sort((a, b) => (b.score || 0) - (a.score || 0))
})

const averageScore = computed(() => {
  if (stocksStore.recommendations.length === 0) return 0
  const sum = stocksStore.recommendations.reduce((acc, rec) => acc + (rec.score || 0), 0)
  return sum / stocksStore.recommendations.length
})

const topScore = computed(() => {
  if (stocksStore.recommendations.length === 0) return 0
  return Math.max(...stocksStore.recommendations.map((r) => r.score || 0))
})

const uniqueRatings = computed(() => {
  const ratings = new Set(stocksStore.recommendations.map((r) => r.latest_rating || 'Unknown'))
  return ratings.size
})

// Methods
const getScoreColor = (score: number) => {
  if (score >= 8) {
    return 'bg-green-100 text-green-800'
  } else if (score >= 6) {
    return 'bg-yellow-100 text-yellow-800'
  } else if (score >= 4) {
    return 'bg-orange-100 text-orange-800'
  } else {
    return 'bg-red-100 text-red-800'
  }
}

const formatDate = (dateString: string) => {
  if (!dateString) return 'Unknown'

  try {
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(dateString))
  } catch {
    return 'Unknown'
  }
}

const getTimeAgo = (dateString: string) => {
  if (!dateString) return 'Unknown'

  try {
    const now = new Date()
    const date = new Date(dateString)
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / 60000)

    if (diffInMinutes < 1) return 'Just now'
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`

    const diffInHours = Math.floor(diffInMinutes / 60)
    if (diffInHours < 24) return `${diffInHours}h ago`

    const diffInDays = Math.floor(diffInHours / 24)
    if (diffInDays < 30) return `${diffInDays}d ago`

    return formatDate(dateString)
  } catch {
    return 'Unknown'
  }
}

const navigateToStock = (ticker: string) => {
  router.push(`/stock/${ticker}`)
}

// Lifecycle
onMounted(async () => {
  // Always fetch recommendations - the store will handle duplicate prevention
  try {
    await stocksStore.fetchRecommendations()
  } catch (error) {
    console.error('Failed to load recommendations:', error)
  }
})
</script>
