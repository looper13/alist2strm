export interface AlistConfig {
  host: string
  token: string
  perPage?: number
  maxRetries?: number
  retryDelay?: number
  reqDelay?: number
}

export interface AlistOptions {
  perPage?: number
  maxRetries?: number
  retryDelay?: number
  reqDelay?: number
}

export interface RetryOptions {
  maxRetries: number
  retryDelay: number
  reqDelay: number
} 