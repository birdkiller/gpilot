import client from './client'

export interface DashboardOverview {
  alert_stats: {
    firing: number
    resolved: number
    acknowledged: number
    suppressed: number
    total_groups: number
  }
  recent_analyses: any[]
  top_namespaces: { namespace: string; count: number }[]
}

export const dashboardApi = {
  overview: () =>
    client.get<DashboardOverview>('/dashboard/overview'),
}
