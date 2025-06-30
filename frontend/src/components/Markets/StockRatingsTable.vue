<template>
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
          :key="`rating-${rating.ticker}-${index}`"
          class="hover:bg-gray-50 cursor-pointer"
          @click="$router.push(`/stock/${rating.ticker}`)"
        >
          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            {{ (currentPage - 1) * pageSize + index + 1 }}
          </td>

          <td class="px-6 py-4 whitespace-nowrap">
            <div class="flex items-center">
              <StockLogo :symbol="rating.ticker" size="sm" />
              <div class="ml-3">
                <div class="text-sm font-medium text-gray-900">{{ rating.ticker }}</div>
                <div class="text-sm text-gray-500 truncate max-w-32">{{ rating.company }}</div>
              </div>
            </div>
          </td>

          <td class="px-6 py-4 whitespace-nowrap text-right">
            <div class="flex items-center justify-end space-x-3">
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

              <div
                v-if="rating.target_from && rating.target_from !== rating.target_to"
                class="text-blue-500 text-lg font-bold animate-pulse"
              >
                →
              </div>

              <div class="text-right">
                <div
                  class="text-xs font-bold text-gray-900 bg-blue-50 px-3 py-1.5 rounded-lg border-2 border-blue-200"
                >
                  ${{ (rating.target_to || 0).toFixed(2) }}
                </div>
                <div class="text-xs text-gray-600 mt-1">Current</div>
              </div>

              <div class="flex items-center">
                <component
                  :is="getTargetTrendIcon(rating)"
                  class="h-5 w-5"
                  :class="getTargetTrendColor(rating) + ' drop-shadow-sm'"
                />
              </div>
            </div>
          </td>

          <td class="px-6 py-4 whitespace-nowrap text-right">
            <div class="flex items-center justify-end space-x-3">
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

              <div
                v-if="rating.rating_from && rating.rating_from !== rating.rating_to"
                class="text-blue-500 text-lg font-bold animate-pulse"
              >
                →
              </div>

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

          <td class="px-6 py-4 whitespace-nowrap text-right">
            <div class="text-sm text-gray-900">{{ rating.brokerage }}</div>
            <div class="text-xs text-gray-500">{{ formatDate(rating.time) }}</div>
          </td>

          <td class="px-6 py-4 whitespace-nowrap text-right">
            <div class="w-32 h-8">
              <MiniChart :symbol="rating.ticker" :rating="rating" period="1W" />
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUpIcon, ArrowDownIcon, MinusIcon } from '@heroicons/vue/24/outline'
import { ChevronUpIcon, ChevronDownIcon } from '@heroicons/vue/20/solid'
import StockLogo from '@/components/StockLogo.vue'
import MiniChart from '@/components/MiniChart.vue'
import type { StockRating, RatingsFilters } from '@/types'

interface Props {
  ratings: StockRating[]
  sortBy: RatingsFilters['sort_by']
  sortOrder: RatingsFilters['order']
  currentPage: number
  pageSize: number
}

interface Emits {
  (e: 'sort', column: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const getRatingPriority = (rating: string): number => {
  const r = rating.toLowerCase()
  if (r.includes('strong buy') || r.includes('outperform')) return 5
  if (r.includes('buy')) return 4
  if (r.includes('hold') || r.includes('neutral')) return 3
  if (r.includes('underweight') || r.includes('underperform')) return 2
  if (r.includes('sell') || r.includes('strong sell')) return 1
  return 0
}

const sortedRatings = computed(() => {
  if (!Array.isArray(props.ratings)) return []

  const sortedArray = [...props.ratings]

  switch (props.sortBy) {
    case 'rating_to':
      return sortedArray.sort((a, b) => {
        const priorityA = getRatingPriority(a.rating_to || '')
        const priorityB = getRatingPriority(b.rating_to || '')
        return props.sortOrder === 'desc' ? priorityB - priorityA : priorityA - priorityB
      })

    case 'ticker':
      return sortedArray.sort((a, b) => {
        const tickerA = (a.ticker || '').toLowerCase()
        const tickerB = (b.ticker || '').toLowerCase()
        return props.sortOrder === 'desc'
          ? tickerB.localeCompare(tickerA)
          : tickerA.localeCompare(tickerB)
      })

    case 'target_to':
      return sortedArray.sort((a, b) => {
        const targetA = a.target_to || 0
        const targetB = b.target_to || 0
        return props.sortOrder === 'desc' ? targetB - targetA : targetA - targetB
      })

    case 'brokerage':
      return sortedArray.sort((a, b) => {
        const brokerageA = (a.brokerage || '').toLowerCase()
        const brokerageB = (b.brokerage || '').toLowerCase()
        return props.sortOrder === 'desc'
          ? brokerageB.localeCompare(brokerageA)
          : brokerageA.localeCompare(brokerageB)
      })

    case 'time':
    case 'updated_at':
      return sortedArray.sort((a, b) => {
        const timeA = new Date(a.time || '').getTime()
        const timeB = new Date(b.time || '').getTime()
        return props.sortOrder === 'desc' ? timeB - timeA : timeA - timeB
      })

    default:
      return sortedArray
  }
})

const getRatingColor = (rating: string) => {
  const r = rating.toLowerCase()
  if (r.includes('buy') || r.includes('strong')) return 'bg-green-100 text-green-800'
  if (r.includes('sell')) return 'bg-red-100 text-red-800'
  if (r.includes('hold')) return 'bg-yellow-100 text-yellow-800'
  return 'bg-gray-100 text-gray-800'
}

const getTargetTrendIcon = (rating: { target_to?: number; target_from?: number }) => {
  const targetTo = rating.target_to || 0
  const targetFrom = rating.target_from || 0

  if (targetFrom && targetTo > targetFrom) return ArrowUpIcon
  if (targetFrom && targetTo < targetFrom) return ArrowDownIcon
  return MinusIcon
}

const getTargetTrendColor = (rating: { target_to?: number; target_from?: number }) => {
  const targetTo = rating.target_to || 0
  const targetFrom = rating.target_from || 0

  if (targetFrom && targetTo > targetFrom) return 'text-green-600'
  if (targetFrom && targetTo < targetFrom) return 'text-red-600'
  return 'text-gray-400'
}

const formatDate = (dateString: string) => {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
  }).format(new Date(dateString))
}

const handleHeaderClick = (column: string) => {
  emit('sort', column)
}

const getSortIcon = (column: string, direction: 'asc' | 'desc') => {
  return direction === 'asc' ? ChevronUpIcon : ChevronDownIcon
}

const getSortIconColor = (column: string, direction: 'asc' | 'desc') => {
  if (props.sortBy === column && props.sortOrder === direction) {
    return 'text-blue-600'
  }
  return 'text-gray-300'
}
</script>
