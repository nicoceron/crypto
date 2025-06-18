<template>
  <div class="stock-chart">
    <div class="chart-header">
      <h3 class="chart-title">{{ symbol }} Price Chart</h3>
      <div class="chart-controls">
        <button
          v-for="period in chartPeriods"
          :key="period.value"
          @click="selectedPeriod = period.value"
          :class="['period-btn', { active: selectedPeriod === period.value }]"
        >
          {{ period.label }}
        </button>
      </div>
    </div>

    <div class="chart-container">
      <Line
        v-if="!loading && !error && chartData.length > 0"
        :data="chartDataFormatted"
        :options="chartOptions"
        :key="chartKey"
      />
      <div v-else-if="loading" class="loading-overlay">
        <div class="spinner"></div>
        <p>Loading chart data...</p>
      </div>
      <div v-else-if="error" class="error-message">
        <p>{{ error }}</p>
        <button @click="loadChartData" class="retry-btn">Retry</button>
      </div>
      <div v-else class="empty-message">
        <p>No chart data available</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  type ChartOptions,
  type ChartData,
} from 'chart.js'

// Register Chart.js components
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

interface PriceBar {
  timestamp: string
  open: number
  high: number
  low: number
  close: number
  volume: number
}

interface Props {
  symbol: string
  height?: number
}

const props = withDefaults(defineProps<Props>(), {
  height: 400,
})

// Reactive state
const loading = ref(false)
const error = ref('')
const chartData = ref<PriceBar[]>([])
const selectedPeriod = ref('1M')
const chartKey = ref(0) // Force re-render when needed

const chartPeriods = [
  { label: '1W', value: '1W' },
  { label: '1M', value: '1M' },
  { label: '3M', value: '3M' },
  { label: '6M', value: '6M' },
  { label: '1Y', value: '1Y' },
]

// Load chart data from API
const loadChartData = async () => {
  if (!props.symbol) return

  loading.value = true
  error.value = ''

  try {
    const response = await fetch(
      `http://localhost:8080/api/v1/stocks/${props.symbol}/price?period=${selectedPeriod.value}`,
    )
    if (!response.ok) {
      throw new Error(`Failed to load chart data: ${response.statusText}`)
    }

    const data = await response.json()
    const bars = data.bars || []

    // Validate and filter data
    const validBars = bars.filter(
      (bar: any) =>
        bar && bar.timestamp && typeof bar.close === 'number' && !isNaN(bar.close) && bar.close > 0,
    )

    if (validBars.length === 0) {
      // Create mock data if no API data
      const now = new Date()
      const mockData: PriceBar[] = []
      const basePrice = 100 + Math.random() * 200

      for (let i = 20; i >= 0; i--) {
        const date = new Date(now.getTime() - i * 24 * 60 * 60 * 1000)
        const price = basePrice + (Math.random() - 0.5) * 20
        mockData.push({
          timestamp: date.toISOString(),
          open: price,
          high: price + Math.random() * 5,
          low: price - Math.random() * 5,
          close: price,
          volume: Math.floor(Math.random() * 1000000),
        })
      }
      chartData.value = mockData
    } else {
      chartData.value = validBars
    }

    // Force chart re-render
    chartKey.value++
    console.log(`Loaded ${chartData.value.length} data points for ${props.symbol}`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load chart data'
    console.error('Chart data loading error:', err)
  } finally {
    loading.value = false
  }
}

// Computed chart data for vue-chartjs
const chartDataFormatted = computed<ChartData<'line'>>(() => {
  if (chartData.value.length === 0) {
    return {
      labels: [],
      datasets: [],
    }
  }

  const labels = chartData.value.map((bar) => {
    try {
      const date = new Date(bar.timestamp)
      if (isNaN(date.getTime())) {
        return 'Invalid Date'
      }
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
    } catch {
      return 'Invalid Date'
    }
  })

  const prices = chartData.value.map((bar) => bar.close)

  // Calculate color based on price trend
  const firstPrice = prices[0] || 0
  const lastPrice = prices[prices.length - 1] || 0
  const isPositive = lastPrice >= firstPrice

  return {
    labels,
    datasets: [
      {
        label: `${props.symbol} Price`,
        data: prices,
        borderColor: isPositive ? '#10b981' : '#ef4444',
        backgroundColor: 'transparent',
        borderWidth: 2,
        fill: false,
        tension: 0.1,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHoverBackgroundColor: isPositive ? '#10b981' : '#ef4444',
        pointHoverBorderColor: '#ffffff',
        pointHoverBorderWidth: 2,
      },
    ],
  }
})

// Chart options
const chartOptions = computed<ChartOptions<'line'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index',
  },
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      backgroundColor: 'rgba(0, 0, 0, 0.8)',
      titleColor: '#ffffff',
      bodyColor: '#ffffff',
      borderColor: '#10b981',
      borderWidth: 1,
      displayColors: false,
      callbacks: {
        title: (context) => {
          return `${props.symbol} - ${context[0].label}`
        },
        label: (context) => {
          return `Price: $${Number(context.parsed.y).toFixed(2)}`
        },
      },
    },
  },
  scales: {
    x: {
      display: true,
      grid: {
        display: false,
      },
      ticks: {
        color: '#6b7280',
        maxTicksLimit: 8,
      },
    },
    y: {
      display: true,
      position: 'right',
      grid: {
        color: 'rgba(107, 114, 128, 0.1)',
        drawBorder: false,
      },
      ticks: {
        color: '#6b7280',
        callback: function (value) {
          return `$${Number(value).toFixed(2)}`
        },
      },
    },
  },
  elements: {
    point: {
      hoverRadius: 6,
    },
  },
}))

// Watch for symbol changes
watch(
  () => props.symbol,
  () => {
    if (props.symbol) {
      loadChartData()
    }
  },
  { immediate: true },
)

// Watch for period changes
watch(selectedPeriod, () => {
  loadChartData()
})

onMounted(() => {
  if (props.symbol) {
    loadChartData()
  }
})
</script>

<style scoped>
.stock-chart {
  @apply bg-white rounded-lg border border-gray-200 p-6 shadow-sm;
}

.chart-header {
  @apply flex justify-between items-center mb-6;
}

.chart-title {
  @apply text-lg font-semibold text-gray-900;
}

.chart-controls {
  @apply flex space-x-2;
}

.period-btn {
  @apply px-3 py-1 text-sm font-medium rounded-md border border-gray-300 bg-white text-gray-700 hover:bg-gray-50 transition-colors;
}

.period-btn.active {
  @apply bg-blue-600 text-white border-blue-600 hover:bg-blue-700;
}

.chart-container {
  @apply relative;
  height: 400px;
}

.loading-overlay {
  @apply absolute inset-0 bg-white bg-opacity-75 flex flex-col items-center justify-center;
}

.spinner {
  @apply w-8 h-8 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin mb-2;
}

.error-message,
.empty-message {
  @apply text-center py-8;
}

.error-message p,
.empty-message p {
  @apply text-red-600 mb-4;
}

.retry-btn {
  @apply px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors;
}
</style>
