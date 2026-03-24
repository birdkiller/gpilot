import { useState } from 'react'
import {
  Card, Input, Button, Space, Typography, Select, Table, Tag, Spin, Divider, message,
} from 'antd'
import { SearchOutlined, RobotOutlined, ThunderboltOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import ReactMarkdown from 'react-markdown'
import { logApi, LogEntry, LogQueryResult, NaturalQueryResult } from '../api/logs'

export default function LogExplorer() {
  const [mode, setMode] = useState<'natural' | 'logql'>('natural')
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(false)
  const [entries, setEntries] = useState<LogEntry[]>([])
  const [queryExpr, setQueryExpr] = useState('')
  const [explanation, setExplanation] = useState('')
  const [selectedLogs, setSelectedLogs] = useState<string[]>([])
  const [analysisResult, setAnalysisResult] = useState('')
  const [analyzingLogs, setAnalyzingLogs] = useState(false)

  const handleSearch = async () => {
    if (!query.trim()) return
    setLoading(true)
    setAnalysisResult('')

    try {
      if (mode === 'natural') {
        const res = await logApi.naturalQuery(query)
        const data: NaturalQueryResult = res.data
        setEntries(data.result?.entries || [])
        setQueryExpr(data.translated_query)
        setExplanation(data.explanation)
      } else {
        const res = await logApi.query({ query, limit: 200 })
        const data: LogQueryResult = res.data
        setEntries(data.entries || [])
        setQueryExpr(data.query_expr)
        setExplanation('')
      }
    } catch (err: any) {
      message.error('Query failed: ' + (err.response?.data?.error || err.message))
    } finally {
      setLoading(false)
    }
  }

  const handleAnalyzeLogs = async () => {
    if (selectedLogs.length === 0) {
      message.warning('Please select log lines to analyze')
      return
    }
    setAnalyzingLogs(true)
    try {
      const res = await logApi.analyze(selectedLogs)
      setAnalysisResult(res.data.root_cause || res.data.summary || JSON.stringify(res.data))
    } catch (err: any) {
      message.error('Log analysis failed')
    } finally {
      setAnalyzingLogs(false)
    }
  }

  const logColumns = [
    {
      title: 'Timestamp',
      dataIndex: 'timestamp',
      width: 200,
      render: (t: string) => (
        <Typography.Text code style={{ fontSize: 11 }}>
          {dayjs(t).format('YYYY-MM-DD HH:mm:ss.SSS')}
        </Typography.Text>
      ),
    },
    {
      title: 'Log Line',
      dataIndex: 'line',
      render: (line: string) => {
        const isError = /error|fatal|panic|exception/i.test(line)
        const isWarn = /warn|warning/i.test(line)
        return (
          <div
            className={`log-line ${isError ? 'log-line-error' : isWarn ? 'log-line-warn' : ''}`}
          >
            {line}
          </div>
        )
      },
    },
    {
      title: 'Labels',
      dataIndex: 'labels',
      width: 200,
      render: (labels: Record<string, string>) =>
        labels ? (
          <Space wrap size={2}>
            {Object.entries(labels).slice(0, 3).map(([k, v]) => (
              <Tag key={k} style={{ fontSize: 11 }}>{k}={v}</Tag>
            ))}
          </Space>
        ) : null,
    },
  ]

  return (
    <div>
      <Typography.Title level={4}>Log Explorer</Typography.Title>

      <Card bordered={false}>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Space>
            <Select
              value={mode}
              onChange={setMode}
              style={{ width: 160 }}
              options={[
                { label: 'Natural Language', value: 'natural' },
                { label: 'LogQL', value: 'logql' },
              ]}
            />
          </Space>

          <Input.Search
            placeholder={
              mode === 'natural'
                ? 'e.g. "Show me OOM errors in production namespace in the last hour"'
                : '{namespace="production"} |= "error"'
            }
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onSearch={handleSearch}
            enterButton={<><SearchOutlined /> Search</>}
            size="large"
            loading={loading}
          />

          {queryExpr && (
            <div>
              <Typography.Text type="secondary">Translated Query: </Typography.Text>
              <Typography.Text code>{queryExpr}</Typography.Text>
            </div>
          )}

          {explanation && (
            <div style={{ background: '#f6f8fa', padding: 12, borderRadius: 6 }}>
              <Typography.Text type="secondary">
                <ThunderboltOutlined /> {explanation}
              </Typography.Text>
            </div>
          )}
        </Space>
      </Card>

      {entries.length > 0 && (
        <Card
          bordered={false}
          style={{ marginTop: 16 }}
          title={`Results (${entries.length} entries)`}
          extra={
            <Button
              type="primary"
              icon={<RobotOutlined />}
              loading={analyzingLogs}
              onClick={handleAnalyzeLogs}
              disabled={selectedLogs.length === 0}
            >
              AI Analyze Selected ({selectedLogs.length})
            </Button>
          }
        >
          <Table
            dataSource={entries}
            columns={logColumns}
            rowKey={(_, i) => String(i)}
            size="small"
            pagination={{ pageSize: 50, showSizeChanger: true }}
            scroll={{ y: 500 }}
            rowSelection={{
              onChange: (_, rows) => setSelectedLogs(rows.map((r) => r.line)),
            }}
          />
        </Card>
      )}

      {analysisResult && (
        <Card
          title={<><RobotOutlined /> AI Log Analysis</>}
          bordered={false}
          style={{ marginTop: 16 }}
        >
          <ReactMarkdown>{analysisResult}</ReactMarkdown>
        </Card>
      )}
    </div>
  )
}
