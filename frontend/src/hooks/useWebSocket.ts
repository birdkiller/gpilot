import { useEffect, useRef, useCallback, useState } from 'react'
import { useAlertStore } from '../store/alertStore'

interface WSMessage {
  type: string
  payload: any
}

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const [connected, setConnected] = useState(false)
  const [analysisChunks, setAnalysisChunks] = useState<Record<string, string>>({})
  const addAlert = useAlertStore((s) => s.addAlert)
  const updateAlert = useAlertStore((s) => s.updateAlert)

  const connect = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws/alerts`)

    ws.onopen = () => {
      setConnected(true)
      console.log('WebSocket connected')
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        switch (msg.type) {
          case 'alert:new':
            addAlert(msg.payload)
            break
          case 'alert:updated':
          case 'alert:resolved':
            updateAlert(msg.payload)
            break
          case 'analysis:chunk':
            setAnalysisChunks((prev) => {
              const key = msg.payload.alert_id || 'current'
              return { ...prev, [key]: (prev[key] || '') + msg.payload.chunk }
            })
            break
          case 'analysis:done':
            // Analysis complete - could trigger UI update
            break
          case 'analysis:started':
            setAnalysisChunks((prev) => {
              const key = msg.payload.alert_id || 'current'
              return { ...prev, [key]: '' }
            })
            break
        }
      } catch (err) {
        console.error('Failed to parse WS message:', err)
      }
    }

    ws.onclose = () => {
      setConnected(false)
      console.log('WebSocket disconnected, reconnecting in 3s...')
      setTimeout(connect, 3000)
    }

    ws.onerror = () => {
      ws.close()
    }

    wsRef.current = ws
  }, [addAlert, updateAlert])

  useEffect(() => {
    connect()
    return () => {
      wsRef.current?.close()
    }
  }, [connect])

  const clearChunks = useCallback((key: string) => {
    setAnalysisChunks((prev) => {
      const next = { ...prev }
      delete next[key]
      return next
    })
  }, [])

  return { connected, analysisChunks, clearChunks }
}
