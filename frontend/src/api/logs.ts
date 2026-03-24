import client from './client'

export interface LogEntry {
  timestamp: string
  line: string
  labels?: Record<string, string>
  stream?: string
}

export interface LogQueryResult {
  entries: LogEntry[]
  total: number
  query_expr: string
}

export interface NaturalQueryResult {
  translated_query: string
  explanation: string
  result: LogQueryResult
}

export const logApi = {
  query: (data: { source?: string; query: string; from?: string; to?: string; limit?: number }) =>
    client.post<LogQueryResult>('/logs/query', data),

  naturalQuery: (question: string) =>
    client.post<NaturalQueryResult>('/logs/natural-query', { question }),

  analyze: (logs: string[], context?: string) =>
    client.post('/logs/analyze', { logs, context }),
}
