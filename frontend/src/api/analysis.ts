import client from './client'

export interface Analysis {
  id: string
  alert_id?: string
  group_id?: string
  type: 'root_cause' | 'log_analysis' | 'event_diagnosis'
  summary: string
  root_cause: string
  suggestions: { title: string; description: string; commands?: string[] }[]
  severity_suggestion?: string
  llm_model: string
  llm_tokens_used: number
  created_at: string
}

export const analysisApi = {
  analyzeAlert: (alertId: string) =>
    client.post<Analysis>(`/alerts/${alertId}/analyze`),

  get: (id: string) =>
    client.get<Analysis>(`/analyses/${id}`),

  listByAlert: (alertId: string) =>
    client.get<{ analyses: Analysis[] }>(`/alerts/${alertId}/analyses`),
}
