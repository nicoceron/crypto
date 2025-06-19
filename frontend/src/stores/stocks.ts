import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  StockRating,
  StockRecommendation,
  PaginatedResponse,
  RatingsFilters,
  LoadingState,
  ApiError,
} from '@/types'
import { apiService } from '@/services/api'

export const useStocksStore = defineStore('stocks', () => {
  // State
  const ratings = ref<StockRating[]>([])
  const recommendations = ref<StockRecommendation[]>([])
  const pagination = ref({
    page: 1,
    limit: 20,
    total_items: 0,
    total_pages: 0,
  })

  const loadingState = ref<LoadingState>({
    isLoading: false,
    error: null,
    lastUpdated: null,
  })

  const filters = ref<RatingsFilters>({
    page: 1,
    limit: 20,
    sort_by: 'updated_at',
    order: 'desc',
    search: '',
  })

  // Computed
  const isLoading = computed(() => loadingState.value.isLoading)
  const error = computed(() => loadingState.value.error)
  const hasError = computed(() => !!loadingState.value.error)
  const lastUpdated = computed(() => loadingState.value.lastUpdated)

  const totalRatings = computed(() => pagination.value.total_items || 0)
  const currentPage = computed(() => pagination.value.page || 1)
  const totalPages = computed(() => pagination.value.total_pages || 0)

  const topRecommendations = computed(() =>
    recommendations.value.slice(0, 10).sort((a, b) => (b.score || 0) - (a.score || 0)),
  )

  // Utility functions
  function setLoading(loading: boolean) {
    loadingState.value.isLoading = loading
  }

  function setError(error: string | null) {
    loadingState.value.error = error
  }

  function clearError() {
    loadingState.value.error = null
  }

  function updateLastUpdated() {
    loadingState.value.lastUpdated = new Date()
  }

  // Actions
  async function fetchRatings(newFilters?: Partial<RatingsFilters>) {
    try {
      setLoading(true)
      clearError()

      if (newFilters) {
        filters.value = { ...filters.value, ...newFilters }
      }

      const response: PaginatedResponse<StockRating> = await apiService.getRatings(filters.value)

      ratings.value = response.data
      pagination.value = response.pagination
      updateLastUpdated()

      console.log(
        `📊 Loaded ${response.data.length} ratings (page ${response.pagination.page}/${response.pagination.total_pages})`,
      )
    } catch (err) {
      const error = err as ApiError
      console.error('❌ Failed to fetch ratings:', error)
      setError(error.error || 'Failed to load ratings')
    } finally {
      setLoading(false)
    }
  }

  async function fetchRatingsByTicker(ticker: string) {
    try {
      setLoading(true)
      clearError()

      const data = await apiService.getRatingsByTicker(ticker)

      console.log(`📊 Loaded ${data.length} ratings for ${ticker}`)
      return data
    } catch (err) {
      const error = err as ApiError
      console.error(`❌ Failed to fetch ratings for ${ticker}:`, error)
      setError(error.error || `Failed to load ratings for ${ticker}`)
      return []
    } finally {
      setLoading(false)
    }
  }

  async function fetchRecommendations() {
    try {
      setLoading(true)
      clearError()

      const data = await apiService.getRecommendations()

      recommendations.value = data
      updateLastUpdated()

      console.log(`🎯 Loaded ${data.length} recommendations`)
    } catch (err) {
      const error = err as ApiError
      console.error('❌ Failed to fetch recommendations:', error)
      setError(error.error || 'Failed to load recommendations')
    } finally {
      setLoading(false)
    }
  }

  async function searchRatings(searchQuery: string) {
    await fetchRatings({
      search: searchQuery,
      page: 1, // Reset to first page when searching
    })
  }

  async function sortRatings(
    sortBy: RatingsFilters['sort_by'],
    order: RatingsFilters['order'] = 'desc',
  ) {
    await fetchRatings({
      sort_by: sortBy,
      order,
      page: 1, // Reset to first page when sorting
    })
  }

  async function changePage(page: number) {
    await fetchRatings({ page })
  }

  async function changePageSize(limit: number) {
    await fetchRatings({
      limit,
      page: 1, // Reset to first page when changing page size
    })
  }

  async function triggerDataIngestion() {
    try {
      setLoading(true)
      clearError()

      const result = await apiService.triggerIngestion()
      console.log('🔄 Data ingestion triggered:', result)

      // Refresh data after ingestion
      await Promise.all([fetchRatings(), fetchRecommendations()])

      return result
    } catch (err) {
      const error = err as ApiError
      console.error('❌ Failed to trigger ingestion:', error)
      setError(error.error || 'Failed to trigger data ingestion')
      throw error
    } finally {
      setLoading(false)
    }
  }

  function resetFilters() {
    filters.value = {
      page: 1,
      limit: 20,
      sort_by: 'time',
      order: 'desc',
      search: '',
    }
  }

  function reset() {
    ratings.value = []
    recommendations.value = []
    pagination.value = {
      page: 1,
      limit: 20,
      total_items: 0,
      total_pages: 0,
    }
    loadingState.value = {
      isLoading: false,
      error: null,
      lastUpdated: null,
    }
    resetFilters()
  }

  return {
    // State
    ratings,
    recommendations,
    pagination,
    filters,
    loadingState,

    // Computed
    isLoading,
    error,
    hasError,
    lastUpdated,
    totalRatings,
    currentPage,
    totalPages,
    topRecommendations,

    // Actions
    fetchRatings,
    fetchRatingsByTicker,
    fetchRecommendations,
    searchRatings,
    sortRatings,
    changePage,
    changePageSize,
    triggerDataIngestion,
    resetFilters,
    reset,
    clearError,
  }
})
