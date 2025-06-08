export interface AlistConfig {
  token: string
  host: string
  domain: string
  reqInterval: number
  reqRetryCount: number
  reqRetryInterval: number
}

export interface StrmConfig {
  defaultSuffix: string
  replaceSuffix: boolean
  urlEncode: boolean
}
