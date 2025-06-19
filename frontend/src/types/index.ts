// API Response Types
export interface ApiError {
  error: string
  code: string
  details?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total_items: number
    total_pages: number
  }
}

// Stock Data Types - matching Go backend models exactly
export interface StockRating {
  rating_id: string
  ticker: string
  company: string
  brokerage: string
  action: string
  rating_from: string
  rating_to: string
  target_from: number
  target_to: number
  time: string
  created_at: string
}

export interface StockRecommendation {
  ticker: string
  company: string
  score: number
  rationale: string
  latest_rating: string
  target_price: number
  technical_signal: string
  sentiment_score: number | null
  generated_at: string
}

// Filter and Search Types
export interface RatingsFilters {
  page?: number
  limit?: number
  sort_by?: 'ticker' | 'brokerage' | 'rating_to' | 'target_to' | 'updated_at' | 'time'
  order?: 'asc' | 'desc'
  search?: string
  ticker?: string
  firm?: string
  rating?: string
  sector?: string
}

// UI State Types
export interface LoadingState {
  isLoading: boolean
  error: string | null
  lastUpdated: Date | null
}

export interface TableState {
  currentPage: number
  itemsPerPage: number
  sortBy: string
  sortOrder: 'asc' | 'desc'
  searchQuery: string
}

// Chart Data Types for financial visualization
export interface ChartDataPoint {
  x: string | number
  y: number
  label?: string
}

export interface ChartConfig {
  type: 'line' | 'bar' | 'pie' | 'doughnut'
  data: ChartDataPoint[]
  options?: Record<string, unknown>
}

// Theme Types
export type Theme = 'light' | 'dark' | 'system'

// Navigation Types - using component type for icons
export interface NavItem {
  name: string
  href: string
  icon?: unknown
  current?: boolean
}

// Form Types
export interface SearchForm {
  query: string
  filters: {
    sector?: string
    rating?: string
    firm?: string
    dateRange?: string
  }
}

// Stock Detail View Types
export interface StockDetail {
  ticker: string
  company_name?: string
  current_price?: number
  ratings: StockRating[]
  recommendations: StockRecommendation[]
  sector?: string
  market_cap?: number
  volume?: number
  price_change?: number
  price_change_percent?: number
}

// Component Props Types
export interface BaseComponentProps {
  loading?: boolean
  error?: string | null
  className?: string
}

export interface DataTableColumn {
  key: string
  label: string
  sortable?: boolean
  type?: 'text' | 'number' | 'date' | 'currency' | 'rating'
  width?: string
  align?: 'left' | 'center' | 'right'
}

export interface PaginationProps {
  currentPage: number
  totalPages: number
  totalItems: number
  itemsPerPage: number
  onPageChange: (page: number) => void
  onPageSizeChange?: (size: number) => void
}

// Notification Types
export interface Notification {
  id: string
  type: 'success' | 'warning' | 'error' | 'info'
  title: string
  message?: string
  timeout?: number
  actions?: NotificationAction[]
}

export interface NotificationAction {
  label: string
  action: () => void
  style?: 'primary' | 'secondary'
}

// API Health Check Type
export interface HealthCheck {
  status: string
  service: string
  timestamp: string
  version?: string
  uptime?: string
}
