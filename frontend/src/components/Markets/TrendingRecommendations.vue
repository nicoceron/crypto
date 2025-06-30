<template>
  <div class="bg-white shadow rounded-lg p-3 flex flex-col">
    <div class="flex items-center justify-between mb-2">
      <h3 class="text-lg font-semibold text-gray-900 flex items-center">
        🔥 Trending Recommendations
        <span v-if="isLoading" class="ml-2 text-xs text-blue-600">⚡ Loading priority...</span>
      </h3>
      <router-link to="/recommendations" class="text-blue-600 hover:underline text-sm">
        View more →
      </router-link>
    </div>

    <div v-if="recommendations && recommendations.length > 0" class="space-y-4 flex-1">
      <div
        v-for="rec in topRecommendations.slice(0, 2)"
        :key="rec.ticker"
        class="flex items-center justify-between p-5 hover:bg-gray-50 rounded-lg cursor-pointer border border-gray-100"
        @click="$router.push(`/stock/${rec.ticker}`)"
      >
        <div class="flex items-center space-x-3">
          <StockLogo :symbol="rec.ticker" size="sm" />
          <div>
            <div class="text-sm font-medium text-gray-900">{{ rec.ticker }}</div>
            <div class="text-xs text-gray-500">{{ rec.company || rec.ticker }}</div>
            <div class="text-xs text-gray-400 mt-1">Score {{ (rec.score || 0).toFixed(1) }}</div>
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
</template>

<script setup lang="ts">
import { computed } from 'vue'
import StockLogo from '@/components/StockLogo.vue'
import MiniChart from '@/components/MiniChart.vue'
import type { StockRecommendation } from '@/types'

interface Props {
  recommendations?: StockRecommendation[]
  isLoading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  recommendations: () => [],
  isLoading: false,
})

const topRecommendations = computed(() => {
  if (!props.recommendations) return []
  return [...props.recommendations].sort((a, b) => (b.score || 0) - (a.score || 0))
})
</script>
