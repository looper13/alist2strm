declare namespace AList {
  interface AlistListResponse<T> {
    code: number
    message: string
    data: {
      content: T
    }
  }
  
  interface AlistGetResponse<T> {
    code: number
    message: string
    data: T
  }

  interface AlistFile {
    name: string
    size?: number
    is_dir: boolean
    modified?: string
    created?: string
    sign?: string
    thumb?: string
    type?: number
    hashinfo?: string
    hash_info?: any
    raw_url?: string
    readme?: string
    header?: string
    provider?: string
    related?: any
  }

  interface AlistDir {
    name: string
    modified: string
  }
}

export as namespace AList