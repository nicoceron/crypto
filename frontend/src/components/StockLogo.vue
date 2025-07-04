<template>
  <div class="stock-logo" :class="sizeClass">
    <img
      v-if="logoUrl && !imageError"
      :src="logoUrl"
      :alt="`${symbol} logo`"
      @error="handleImageError"
      class="logo-image"
    />
    <div v-else class="logo-fallback" :style="{ backgroundColor: fallbackColor }">
      <span class="logo-text">{{ symbolInitials }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useStocksStore } from '@/stores/stocks'

interface Props {
  symbol: string
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  shape?: 'circle' | 'rounded' | 'square'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  shape: 'rounded',
})

// Store
const stocksStore = useStocksStore()

// Local state
const imageError = ref(false)

// Get logo URL from store (uses cached data from batch loading)
const logoUrl = computed(() => {
  if (!props.symbol) return ''
  return stocksStore.getLogoUrl(props.symbol)
})

// Computed properties
const sizeClass = computed(() => `size-${props.size} shape-${props.shape}`)

const symbolInitials = computed(() => {
  if (!props.symbol) return ''
  return props.symbol.slice(0, 2).toUpperCase()
})

const fallbackColor = computed(() => {
  // Generate consistent color based on symbol
  const colors = [
    '#3b82f6',
    '#ef4444',
    '#10b981',
    '#f59e0b',
    '#8b5cf6',
    '#06b6d4',
    '#84cc16',
    '#f97316',
    '#ec4899',
    '#6b7280',
  ]

  let hash = 0
  for (let i = 0; i < props.symbol.length; i++) {
    hash = props.symbol.charCodeAt(i) + ((hash << 5) - hash)
  }

  return colors[Math.abs(hash) % colors.length]
})

const handleImageError = () => {
  console.log(`💥 ${props.symbol}: Image failed to load, showing initials`)
  imageError.value = true
}
</script>

<style scoped>
.stock-logo {
  @apply relative inline-flex items-center justify-center overflow-hidden;
}

/* Sizes */
.size-xs {
  @apply w-6 h-6;
}

.size-sm {
  @apply w-8 h-8;
}

.size-md {
  @apply w-12 h-12;
}

.size-lg {
  @apply w-16 h-16;
}

.size-xl {
  @apply w-20 h-20;
}

/* Shapes */
.shape-circle {
  @apply rounded-full;
}

.shape-rounded {
  @apply rounded-lg;
}

.shape-square {
  @apply rounded-none;
}

/* Logo image */
.logo-image {
  @apply w-full h-full object-contain;
}

/* Logo fallback */
.logo-fallback {
  @apply w-full h-full flex items-center justify-center text-white font-bold;
}

.logo-text {
  @apply text-xs font-bold;
}

/* Responsive text sizing */
.size-xs .logo-text {
  @apply text-xs;
}

.size-sm .logo-text {
  @apply text-xs;
}

.size-md .logo-text {
  @apply text-sm;
}

.size-lg .logo-text {
  @apply text-base;
}

.size-xl .logo-text {
  @apply text-lg;
}
</style>
