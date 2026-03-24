import { useEffect, useState } from 'react'
import { Card, Col, Row, Statistic, Typography, Table, Tag, Spin } from 'antd'
import {
  AlertOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  StopOutlined,
  TeamOutlined,
} from '@ant-design/icons'
import { dashboardApi, DashboardOverview } from '../api/dashboard'
import { useWebSocket } from '../hooks/useWebSocket'

export default function Dashboard() {
  const [data, setData] = useState<DashboardOverview | null>(null)
  const [loading, setLoading] = useState(true)
  useWebSocket() // Connect for real-time updates

  useEffect(() => {
    loadData()
    const interval = setInterval(loadData, 30000)
    return () => clearInterval(interval)
  }, [])

  const loadData = async () => {
    try {
      const res = await dashboardApi.overview()
      setData(res.data)
    } catch (err) {
      console.error('Failed to load dashboard:', err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Spin size="large" />
      </div>
    )
  }

  const stats = data?.alert_stats

  return (
    <div>
      <Typography.Title level={4} style={{ marginBottom: 16 }}>
        gPilot Dashboard
      </Typography.Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false}>
            <Statistic
              title="Firing Alerts"
              value={stats?.firing ?? 0}
              prefix={<AlertOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false}>
            <Statistic
              title="Resolved (24h)"
              value={stats?.resolved ?? 0}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false}>
            <Statistic
              title="Acknowledged"
              value={stats?.acknowledged ?? 0}
              prefix={<ClockCircleOutlined style={{ color: '#1677ff' }} />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false}>
            <Statistic
              title="Alert Groups"
              value={stats?.total_groups ?? 0}
              prefix={<TeamOutlined style={{ color: '#722ed1' }} />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <Card title="Recent AI Analyses" bordered={false}>
            {data?.recent_analyses && data.recent_analyses.length > 0 ? (
              <Table
                dataSource={data.recent_analyses}
                rowKey="id"
                pagination={false}
                size="small"
                columns={[
                  {
                    title: 'Type',
                    dataIndex: 'type',
                    render: (t: string) => (
                      <Tag color={t === 'root_cause' ? 'red' : t === 'log_analysis' ? 'blue' : 'green'}>
                        {t === 'root_cause' ? 'Root Cause' : t === 'log_analysis' ? 'Log Analysis' : 'Event Diagnosis'}
                      </Tag>
                    ),
                  },
                  { title: 'Summary', dataIndex: 'summary', ellipsis: true },
                  {
                    title: 'Model',
                    dataIndex: 'llm_model',
                    width: 120,
                    render: (m: string) => <Tag>{m}</Tag>,
                  },
                  {
                    title: 'Tokens',
                    dataIndex: 'llm_tokens_used',
                    width: 80,
                  },
                ]}
              />
            ) : (
              <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
                No analyses yet. Trigger an analysis from the Alert Console.
              </div>
            )}
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="System Status" bordered={false}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <StatusItem label="Backend API" status="online" />
              <StatusItem label="PostgreSQL" status="online" />
              <StatusItem label="Redis" status="online" />
              <StatusItem label="WebSocket" status="online" />
              <StatusItem label="Prometheus" status="configured" />
              <StatusItem label="Loki" status="configured" />
              <StatusItem label="K8s Cluster" status="optional" />
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

function StatusItem({ label, status }: { label: string; status: string }) {
  const colorMap: Record<string, string> = {
    online: '#52c41a',
    configured: '#1677ff',
    optional: '#d9d9d9',
    offline: '#ff4d4f',
  }
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <span>{label}</span>
      <Tag color={colorMap[status] || '#d9d9d9'}>{status}</Tag>
    </div>
  )
}
