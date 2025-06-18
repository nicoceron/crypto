import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import type {
  StockRating,
  StockRecommendation,
  PaginatedResponse,
  RatingsFilters,
  ApiError,
} from '@/types'

class ApiService {
  private api: AxiosInstance

  constructor() {
    this.api = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor
    this.api.interceptors.request.use(
      (config) => {
        console.log(`🚀 API Request: ${config.method?.toUpperCase()} ${config.url}`)
        return config
      },
      (error) => {
        console.error('❌ API Request Error:', error)
        return Promise.reject(error)
      },
    )

    // Response interceptor
    this.api.interceptors.response.use(
      (response: AxiosResponse) => {
        console.log(`✅ API Response: ${response.status} ${response.config.url}`)
        return response
      },
      (error) => {
        console.error('❌ API Response Error:', error.response?.data || error.message)

        // Transform backend errors to our ApiError format
        if (error.response?.data) {
          const apiError: ApiError = {
            error: error.response.data.error || 'An error occurred',
            code: error.response.data.code || 'UNKNOWN_ERROR',
            details: error.response.data.details,
          }
          return Promise.reject(apiError)
        }

        return Promise.reject({
          error: error.message || 'Network error',
          code: 'NETWORK_ERROR',
        } as ApiError)
      },
    )
  }

  // Get paginated stock ratings with optional filters
  async getRatings(filters: RatingsFilters = {}): Promise<PaginatedResponse<StockRating>> {
    const params = new URLSearchParams()

    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    if (filters.sort_by) params.append('sort_by', filters.sort_by)
    if (filters.order) params.append('order', filters.order)
    if (filters.search) params.append('search', filters.search)

    const response = await this.api.get<PaginatedResponse<StockRating>>(`/ratings?${params}`)
    return response.data
  }

  // Get ratings for a specific ticker
  async getRatingsByTicker(ticker: string): Promise<StockRating[]> {
    const response = await this.api.get<StockRating[]>(`/ratings/${ticker}`)
    return response.data
  }

  // Get stock recommendations
  async getRecommendations(): Promise<StockRecommendation[]> {
    const response = await this.api.get<StockRecommendation[]>('/recommendations')
    return response.data
  }

  // Trigger data ingestion (admin function)
  async triggerIngestion(): Promise<{ message: string; status: string }> {
    const response = await this.api.post<{ message: string; status: string }>('/ingest')
    return response.data
  }

  // Health check
  async healthCheck(): Promise<{ status: string; service: string; timestamp: string }> {
    // Health endpoint is at root level, not under /api/v1
    const response = await axios.get('http://localhost:8080/health')
    return response.data
  }

  // Helper method to handle retries for failed requests
  async retry<T>(fn: () => Promise<T>, maxRetries = 3, delay = 1000): Promise<T> {
    for (let i = 0; i < maxRetries; i++) {
      try {
        return await fn()
      } catch (error) {
        if (i === maxRetries - 1) throw error

        console.log(`🔄 Retrying request in ${delay}ms... (attempt ${i + 1}/${maxRetries})`)
        await new Promise((resolve) => setTimeout(resolve, delay))
        delay *= 2 // Exponential backoff
      }
    }
    throw new Error('Max retries exceeded')
  }
}

// Export a singleton instance
export const apiService = new ApiService()
export default apiService
 