import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider, theme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import AppLayout from './components/layout/AppLayout'
import Dashboard from './pages/Dashboard'
import AlertConsole from './pages/AlertConsole'
import AlertDetail from './pages/AlertDetail'
import LogExplorer from './pages/LogExplorer'
import EventTimeline from './pages/EventTimeline'

function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: {
          colorPrimary: '#1677ff',
          borderRadius: 6,
        },
      }}
    >
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<AppLayout />}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="alerts" element={<AlertConsole />} />
            <Route path="alerts/:id" element={<AlertDetail />} />
            <Route path="logs" element={<LogExplorer />} />
            <Route path="events" element={<EventTimeline />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}

export default App
