import client from './client'

export interface Alert {
  id: string
  fingerprint: string
  group_id?: string
  status: 'firing' | 'resolved' | 'suppressed' | 'acknowledged'
  severity: 'critical' | 'warning' | 'info'
  adjusted_severity?: string
  name: string
  namespace: string
  pod?: string
  node?: string
  labels: Record<string, string>
  annotations: Record<string, string>
  started_at: string
  resolved_at?: string
  last_active_at: string
  acknowledged_by?: string
  acknowledged_at?: string
  created_at: string
  updated_at: string
}

export interface AlertListResult {
  alerts: Alert[]
  total: number
  page: number
  size: number
}

export interface AlertGroup {
  id: string
  name: string
  namespace: string
  status: string
  alert_count: number
  root_cause_id?: string
  created_at: string
  updated_at: string
}

export interface AlertListParams {
  status?: string
  severity?: string
  namespace?: string
  search?: string
  from?: string
  to?: string
  page?: number
  size?: number
}

export const alertApi = {
  list: (params: AlertListParams) =>
    client.get<AlertListResult>('/alerts', { params }),

  get: (id: string) =>
    client.get<Alert>(`/alerts/${id}`),

  acknowledge: (id: string, user: string) =>
    client.put(`/alerts/${id}/acknowledge`, { user }),

  listGroups: (page = 1, size = 20) =>
    client.get('/alert-groups', { params: { page, size } }),

  getGroup: (id: string) =>
    client.get<AlertGroup>(`/alert-groups/${id}`),
}
