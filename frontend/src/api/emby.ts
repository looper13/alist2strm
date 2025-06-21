import { http } from './http'

// Emby 媒体库信息类型
interface EmbyLibrary {
  Id: string // 媒体库 ID
  Name: string // 媒体库名称
  Locations?: string[] // 媒体库位置路径列表
  CollectionType?: string // 集合类型(movies,tvshows,music等)
  ItemId?: string // 项目ID(兼容旧版)
  Guid?: string // 全局唯一标识符
  PrimaryImageItemId?: string // 主图像项目ID
  PrimaryImageTag?: string // 主图像标签
  RefreshProgress?: number // 刷新进度
  RefreshStatus?: string // 刷新状态
  ItemCount: number // 媒体库中的项目数量
  Type?: string // 媒体类型
  LastUpdate?: string // 最后更新时间
  LibraryOptions?: any // 库配置选项
}

// Emby 最新入库媒体信息类型
interface EmbyLatestMedia {
  Id: string
  Name: string
  Path: string
  Type: string
  DateCreated: string
  PremiereDate?: string
  ProductionYear?: number
  Overview?: string
}

// Emby 连接测试结果类型
interface EmbyConnectionTestResult {
  connected: boolean
  error?: string
  version?: string
  serverName?: string
  operatingSystem?: string
}

export class EmbyAPI {
  private baseUrl = '/emby'

  /**
   * 测试 Emby 服务器连接
   * @returns 连接测试结果
   */
  async testConnection() {
    return http.get<EmbyConnectionTestResult>(`${this.baseUrl}/test`)
  }

  /**
   * 获取 Emby 媒体库列表
   * @returns 媒体库列表
   */
  async getLibraries() {
    return http.get<EmbyLibrary[]>(`${this.baseUrl}/libraries`)
  }

  /**
   * 获取 Emby 最新入库媒体
   * @param limit 返回结果数量限制，默认 10
   * @returns 最新入库媒体列表
   */
  async getLatestMedia(limit: number = 10) {
    return http.get<EmbyLatestMedia[]>(`${this.baseUrl}/latest`, {
      params: { limit },
    })
  }

  /**
   * 刷新指定的 Emby 媒体库
   * @param libraryId 媒体库 ID
   * @returns 刷新结果
   */
  async refreshLibrary(libraryId: string) {
    return http.post(`${this.baseUrl}/libraries/${libraryId}/refresh`)
  }

  /**
   * 刷新所有 Emby 媒体库
   * @returns 刷新结果
   */
  async refreshAllLibraries() {
    return http.post(`${this.baseUrl}/libraries/refresh`)
  }

  /**
   * 获取单个 Emby 媒体库详细信息
   * @param libraryId 媒体库 ID
   * @returns 媒体库详细信息
   */
  async getLibraryDetails(libraryId: string) {
    return http.get<EmbyLibrary>(`${this.baseUrl}/libraries/${libraryId}`)
  }

  /**
   * 获取 Emby 图片的 URL
   * @param itemId 项目 ID
   * @param imageType 图片类型 (Primary, Backdrop, Logo等)
   * @param options 可选参数对象
   * @param options.tag 可选的图片标签
   * @param options.maxWidth 可选的最大宽度
   * @param options.maxHeight 可选的最大高度
   * @param options.quality 可选的图片质量
   * @returns 图片 URL
   */
  getImageUrl(
    itemId: string,
    imageType: string,
    options?: {
      tag?: string
      maxWidth?: number
      maxHeight?: number
      quality?: number
    },
  ): string {
    // 在开发环境使用完整服务器 URL，生产环境使用相对路径
    const prefix = import.meta.env.VITE_DIRECT_SERVER_URL === 'true'
      ? import.meta.env.VITE_API_BASE_URL || ''
      : ''

    let url = `${prefix}/api/emby/items/${itemId}/images/${imageType}`

    // 添加可选查询参数
    const params = new URLSearchParams()
    if (options?.tag) {
      params.append('tag', options.tag)
    }
    if (options?.maxWidth) {
      params.append('max_width', options.maxWidth.toString())
    }
    if (options?.maxHeight) {
      params.append('max_height', options.maxHeight.toString())
    }
    if (options?.quality) {
      params.append('quality', options.quality.toString())
    }

    const queryString = params.toString()
    if (queryString) {
      url += `?${queryString}`
    }

    return url
  } /**
     * 直接获取 Emby 图片内容
     * @param itemId 项目 ID
     * @param imageType 图片类型 (Primary, Backdrop, Logo等)
     * @param options 可选参数对象
     * @param options.tag 可选的图片标签
     * @param options.maxWidth 可选的最大宽度
     * @param options.maxHeight 可选的最大高度
     * @param options.quality 可选的图片质量
     * @returns Promise<Blob> 图片数据的 Blob 对象
     */

  async getImageContent(
    itemId: string,
    imageType: string,
    options?: {
      tag?: string
      maxWidth?: number
      maxHeight?: number
      quality?: number
    },
  ): Promise<Blob> {
    const url = this.getImageUrl(itemId, imageType, options)

    // 使用原生 fetch API 获取图片数据
    const response = await fetch(url, {
      // 凭证策略，确保发送Cookie (如果有)
      credentials: 'include',
    })

    if (!response.ok) {
      throw new Error(`获取图片失败: ${response.status} ${response.statusText}`)
    }

    return await response.blob()
  }

  /**
   * 获取 Emby 图片并创建一个对象 URL (可以直接用于 img src)
   * @param itemId 项目 ID
   * @param imageType 图片类型 (Primary, Backdrop, Logo等)
   * @param options 图片参数对象
   * @param options.tag 可选的图片标签
   * @param options.maxWidth 可选的最大宽度
   * @param options.maxHeight 可选的最大高度
   * @param options.quality 可选的图片质量
   * @returns Promise<string> 图片的对象 URL
   */
  async getImageObjectUrl(
    itemId: string,
    imageType: string,
    options?: {
      tag?: string
      maxWidth?: number
      maxHeight?: number
      quality?: number
    },
  ): Promise<string> {
    const blob = await this.getImageContent(itemId, imageType, options)
    return URL.createObjectURL(blob)
  }
}

// 导出 Emby API 实例
export const embyAPI = new EmbyAPI()
