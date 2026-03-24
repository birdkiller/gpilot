import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card, Descriptions, Tag, Button, Space, Typography, Spin, Divider, message, Timeline,
} from 'antd'
import {
  ArrowLeftOutlined, RobotOutlined, CheckOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import ReactMarkdown from 'react-markdown'
import { Alert, alertApi } from '../api/alerts'
import { Analysis, analysisApi } from '../api/analysis'
import { useWebSocket } from '../hooks/useWebSocket'

const severityColors: Record<string, string> = {
  critical: 'red', warning: 'orange', info: 'blue',
}

export default function AlertDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [alert, setAlert] = useState<Alert | null>(null)
  const [analyses, setAnalyses] = useState<Analysis[]>([])
  const [loading, setLoading] = useState(true)
  const [analyzing, setAnalyzing] = useState(false)
  const { analysisChunks } = useWebSocket()

  useEffect(() => {
    if (id) loadAlert()
  }, [id])

  const loadAlert = async () => {
    try {
      const [alertRes, analysesRes] = await Promise.all([
        alertApi.get(id!),
        analysisApi.listByAlert(id!),
      ])
      setAlert(alertRes.data)
      setAnalyses(analysesRes.data.analyses || [])
    } catch {
      message.error('Failed to load alert')
    } finally {
      setLoading(false)
    }
  }

  const handleAnalyze = async () => {
    setAnalyzing(true)
    try {
      const res = await analysisApi.analyzeAlert(id!)
      setAnalyses((prev) => [res.data, ...prev])
      message.success('Analysis completed')
    } catch (err: any) {
      message.error('Analysis failed: ' + (err.response?.data?.error || err.message))
    } finally {
      setAnalyzing(false)
    }
  }

  const handleAcknowledge = async () => {
    try {
      await alertApi.acknowledge(id!, 'admin')
      message.success('Acknowledged')
      loadAlert()
    } catch {
      message.error('Failed')
    }
  }

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  if (!alert) {
    return <div style={{ textAlign: 'center', padding: 100 }}>Alert not found</div>
  }

  const streamingText = analysisChunks[id!] || ''

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/alerts')}>
          Back
        </Button>
      </Space>

      <Card bordered={false}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <Typography.Title level={4} style={{ margin: 0 }}>
              <Tag color={severityColors[alert.severity]}>{alert.severity.toUpperCase()}</Tag>
              {alert.name}
            </Typography.Title>
            {alert.annotations?.summary && (
              <Typography.Paragraph type="secondary" style={{ marginTop: 8 }}>
                {alert.annotations.summary}
              </Typography.Paragraph>
            )}
          </div>
          <Space>
            {alert.status === 'firing' && (
              <Button icon={<CheckOutlined />} onClick={handleAcknowledge}>
                Acknowledge
              </Button>
            )}
            <Button
              type="primary"
              icon={<RobotOutlined />}
              loading={analyzing}
              onClick={handleAnalyze}
            >
              AI Root Cause Analysis
            </Button>
          </Space>
        </div>

        <Divider />

        <Descriptions column={2} size="small">
          <Descriptions.Item label="Status">
            <Tag color={alert.status === 'firing' ? 'red' : 'green'}>{alert.status}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Namespace">
            <Tag>{alert.namespace || 'N/A'}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Pod">{alert.pod || '-'}</Descriptions.Item>
          <Descriptions.Item label="Node">{alert.node || '-'}</Descriptions.Item>
          <Descriptions.Item label="Started">{dayjs(alert.started_at).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
          <Descriptions.Item label="Last Active">{dayjs(alert.last_active_at).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
          {alert.resolved_at && (
            <Descriptions.Item label="Resolved">{dayjs(alert.resolved_at).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
          )}
          {alert.acknowledged_by && (
            <Descriptions.Item label="Acknowledged By">{alert.acknowledged_by}</Descriptions.Item>
          )}
        </Descriptions>

        {Object.keys(alert.labels || {}).length > 0 && (
          <>
            <Divider orientation="left" plain>Labels</Divider>
            <Space wrap>
              {Object.entries(alert.labels).map(([k, v]) => (
                <Tag key={k}>{k}={v}</Tag>
              ))}
            </Space>
          </>
        )}
      </Card>

      {/* Streaming Analysis Output */}
      {(analyzing || streamingText) && (
        <Card
          title={<><RobotOutlined /> AI Analysis (Streaming)</>}
          bordered={false}
          style={{ marginTop: 16 }}
        >
          <div className={analyzing ? 'streaming-cursor' : ''}>
            <ReactMarkdown>{streamingText || 'Analyzing...'}</ReactMarkdown>
          </div>
        </Card>
      )}

      {/* Historical Analyses */}
      {analyses.length > 0 && (
        <Card
          title="Analysis History"
          bordered={false}
          style={{ marginTop: 16 }}
        >
          <Timeline
            items={analyses.map((a) => ({
              color: a.type === 'root_cause' ? 'red' : 'blue',
              children: (
                <div key={a.id}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Tag color={a.type === 'root_cause' ? 'red' : 'blue'}>{a.type}</Tag>
                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                      {dayjs(a.created_at).format('YYYY-MM-DD HH:mm:ss')} | {a.llm_model} | {a.llm_tokens_used} tokens
                    </Typography.Text>
                  </div>
                  <div style={{ marginTop: 8 }}>
                    <ReactMarkdown>{a.root_cause || a.summary}</ReactMarkdown>
                  </div>
                  {a.suggestions?.map((s, i) => (
                    <Card key={i} size="small" style={{ marginTop: 8, background: '#fafafa' }}>
                      <Typography.Text strong>{s.title}</Typography.Text>
                      <div><ReactMarkdown>{s.description}</ReactMarkdown></div>
                    </Card>
                  ))}
                </div>
              ),
            }))}
          />
        </Card>
      )}
    </div>
  )
}
