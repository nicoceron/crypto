<template>
  <div class="stock-logo" :class="sizeClass">
    <img
      v-if="logoUrl && !imageError"
      :src="logoUrl"
      :alt="`${symbol} logo`"
      @error="handleImageError"
      @load="handleImageLoad"
      class="logo-image"
      :class="{ loading: loading }"
    />
    <div v-else class="logo-fallback" :style="{ backgroundColor: fallbackColor }">
      <span class="logo-text">{{ symbolInitials }}</span>
    </div>

    <!-- Loading spinner overlay -->
    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'

interface Props {
  symbol: string
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  shape?: 'circle' | 'rounded' | 'square'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  shape: 'rounded',
})

// Reactive state
const logoUrl = ref('')
const loading = ref(false)
const imageError = ref(false)

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

// Methods
const loadLogo = async () => {
  if (!props.symbol) return

  loading.value = true
  imageError.value = false

  try {
    const response = await fetch(`http://localhost:8080/api/v1/stocks/${props.symbol}/logo`)
    if (response.ok) {
      const data = await response.json()
      logoUrl.value = data.logo_url
    } else {
      throw new Error('Logo not found')
    }
  } catch {
    console.log(`Logo not found for ${props.symbol}, using fallback`)
    imageError.value = true
  } finally {
    loading.value = false
  }
}

const handleImageError = () => {
  imageError.value = true
  loading.value = false
}

const handleImageLoad = () => {
  loading.value = false
}

// Watch for symbol changes
watch(
  () => props.symbol,
  () => {
    if (props.symbol) {
      loadLogo()
    }
  },
  { immediate: true },
)

onMounted(() => {
  if (props.symbol) {
    loadLogo()
  }
})
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
  @apply w-full h-full object-contain transition-opacity duration-200;
}

.logo-image.loading {
  @apply opacity-50;
}

/* Fallback */
.logo-fallback {
  @apply w-full h-full flex items-center justify-center text-white font-semibold;
}

.size-xs .logo-text {
  @apply text-xs;
}

.size-sm .logo-text {
  @apply text-sm;
}

.size-md .logo-text {
  @apply text-base;
}

.size-lg .logo-text {
  @apply text-lg;
}

.size-xl .logo-text {
  @apply text-xl;
}

/* Loading overlay */
.loading-overlay {
  @apply absolute inset-0 bg-white bg-opacity-50 flex items-center justify-center;
}

.spinner {
  @apply w-4 h-4 border-2 border-gray-200 border-t-gray-600 rounded-full animate-spin;
}

.size-xs .spinner,
.size-sm .spinner {
  @apply w-3 h-3 border;
}

.size-lg .spinner,
.size-xl .spinner {
  @apply w-6 h-6 border-2;
}
</style>
 