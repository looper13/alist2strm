// 仪表盘 API 模块

// 仪表盘统计数据接口
export interface DashboardStats {
  // 任务统计
  taskStats: {
    total: number // 任务总数
    enabled: number // 启用任务数
    disabled: number // 禁用任务数
    totalExecutions: number // 总执行次数
    todayExecutions: number // 今日执行次数
    todaySuccess: number // 今日成功次数
    todayFailed: number // 今日失败次数
  }

  // STRM 文件统计
  strmStats: {
    totalGenerated: number // 总生成数
    todayGenerated: number // 今日生成数
    downloadSize: number // 原数据下载量(MB)
  }

  // Emby 统计
  embyStats: {
    serverStatus: 'online' | 'offline' | 'error' // 服务器状态
    userCount: number // Emby 用户数
    libraryCount: number // 媒体库数量
    recentItemsCount: number // 近期入库数量
  }
}

export const dashboardAPI = {
  /**
   * 获取仪表盘统计数据
   */
  async getStats(): Promise<Api.Common.HttpResponse<DashboardStats>> {
    // 实际项目中应该有一个统一的仪表盘 API
    // 这里为了演示，我们模拟一些数据
    return Promise.resolve({
      code: 0,
      message: '获取成功',
      data: {
        taskStats: {
          total: 12,
          enabled: 8,
          disabled: 4,
          totalExecutions: 156, // 总执行次数
          todayExecutions: 24,
          todaySuccess: 22,
          todayFailed: 2,
        },
        strmStats: {
          totalGenerated: 1562,
          todayGenerated: 45,
          downloadSize: 24680, // 约25GB
        },
        embyStats: {
          serverStatus: 'online',
          userCount: 4,
          libraryCount: 8,
          recentItemsCount: 15,
        },
      },
    })
  },
}
