import { create } from 'zustand'
import { Alert, alertApi, AlertListParams } from '../api/alerts'

interface AlertState {
  alerts: Alert[]
  total: number
  page: number
  size: number
  loading: boolean
  firingCount: number
  filters: AlertListParams
  setFilters: (filters: Partial<AlertListParams>) => void
  fetchAlerts: () => Promise<void>
  updateAlert: (alert: Alert) => void
  addAlert: (alert: Alert) => void
}

export const useAlertStore = create<AlertState>((set, get) => ({
  alerts: [],
  total: 0,
  page: 1,
  size: 20,
  loading: false,
  firingCount: 0,
  filters: {},

  setFilters: (filters) => {
    set((s) => ({ filters: { ...s.filters, ...filters } }))
    get().fetchAlerts()
  },

  fetchAlerts: async () => {
    set({ loading: true })
    try {
      const { filters, page, size } = get()
      const res = await alertApi.list({ ...filters, page, size })
      const data = res.data
      const firingCount = data.alerts?.filter((a) => a.status === 'firing').length ?? 0
      set({
        alerts: data.alerts || [],
        total: data.total,
        page: data.page,
        size: data.size,
        firingCount,
        loading: false,
      })
    } catch {
      set({ loading: false })
    }
  },

  updateAlert: (updated) => {
    set((s) => ({
      alerts: s.alerts.map((a) => (a.id === updated.id ? updated : a)),
      firingCount: s.alerts.filter((a) =>
        a.id === updated.id ? updated.status === 'firing' : a.status === 'firing'
      ).length,
    }))
  },

  addAlert: (alert) => {
    set((s) => {
      const exists = s.alerts.find((a) => a.id === alert.id)
      if (exists) {
        return {
          alerts: s.alerts.map((a) => (a.id === alert.id ? alert : a)),
        }
      }
      return {
        alerts: [alert, ...s.alerts],
        total: s.total + 1,
        firingCount: alert.status === 'firing' ? s.firingCount + 1 : s.firingCount,
      }
    })
  },
}))
