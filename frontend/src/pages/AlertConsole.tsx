import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Table, Tag, Space, Input, Select, Button, Typography, Badge, message,
} from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import { useAlertStore } from '../store/alertStore'
import { Alert, alertApi } from '../api/alerts'
import { useWebSocket } from '../hooks/useWebSocket'

dayjs.extend(relativeTime)

const severityColors: Record<string, string> = {
  critical: 'red',
  warning: 'orange',
  info: 'blue',
}

const statusColors: Record<string, string> = {
  firing: 'error',
  resolved: 'success',
  acknowledged: 'processing',
  suppressed: 'default',
}

export default function AlertConsole() {
  const navigate = useNavigate()
  const { alerts, total, loading, page, size, fetchAlerts, setFilters } = useAlertStore()
  const [search, setSearch] = useState('')
  useWebSocket()

  useEffect(() => {
    fetchAlerts()
  }, [fetchAlerts])

  const handleAcknowledge = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await alertApi.acknowledge(id, 'admin')
      message.success('Alert acknowledged')
      fetchAlerts()
    } catch {
      message.error('Failed to acknowledge')
    }
  }

  const columns: ColumnsType<Alert> = [
    {
      title: 'Severity',
      dataIndex: 'severity',
      width: 100,
      filters: [
        { text: 'Critical', value: 'critical' },
        { text: 'Warning', value: 'warning' },
        { text: 'Info', value: 'info' },
      ],
      render: (s: string) => (
        <Tag color={severityColors[s]}>{s.toUpperCase()}</Tag>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      width: 110,
      render: (s: string) => (
        <Badge status={statusColors[s] as any} text={s} />
      ),
    },
    {
      title: 'Alert Name',
      dataIndex: 'name',
      ellipsis: true,
      render: (name: string, record: Alert) => (
        <div>
          <Typography.Text strong>{name}</Typography.Text>
          {record.annotations?.summary && (
            <div>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                {record.annotations.summary}
              </Typography.Text>
            </div>
          )}
        </div>
      ),
    },
    {
      title: 'Namespace',
      dataIndex: 'namespace',
      width: 130,
      render: (ns: string) => ns ? <Tag>{ns}</Tag> : '-',
    },
    {
      title: 'Duration',
      dataIndex: 'started_at',
      width: 130,
      render: (t: string) => dayjs(t).fromNow(),
      sorter: (a, b) => dayjs(a.started_at).unix() - dayjs(b.started_at).unix(),
    },
    {
      title: 'Actions',
      width: 120,
      render: (_, record) => (
        <Space>
          {record.status === 'firing' && (
            <Button size="small" onClick={(e) => handleAcknowledge(record.id, e)}>
              ACK
            </Button>
          )}
          <Button size="small" type="link" onClick={() => navigate(`/alerts/${record.id}`)}>
            Detail
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Alert Console
        </Typography.Title>
        <Space>
          <Input
            placeholder="Search alerts..."
            prefix={<SearchOutlined />}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onPressEnter={() => setFilters({ search })}
            style={{ width: 240 }}
            allowClear
          />
          <Select
            placeholder="Status"
            style={{ width: 130 }}
            allowClear
            onChange={(v) => setFilters({ status: v })}
            options={[
              { label: 'Firing', value: 'firing' },
              { label: 'Resolved', value: 'resolved' },
              { label: 'Acknowledged', value: 'acknowledged' },
            ]}
          />
          <Select
            placeholder="Severity"
            style={{ width: 130 }}
            allowClear
            onChange={(v) => setFilters({ severity: v })}
            options={[
              { label: 'Critical', value: 'critical' },
              { label: 'Warning', value: 'warning' },
              { label: 'Info', value: 'info' },
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={fetchAlerts}>
            Refresh
          </Button>
        </Space>
      </div>

      <Table
        dataSource={alerts}
        columns={columns}
        rowKey="id"
        loading={loading}
        size="middle"
        onRow={(record) => ({
          onClick: () => navigate(`/alerts/${record.id}`),
          style: { cursor: 'pointer' },
        })}
        pagination={{
          current: page,
          pageSize: size,
          total,
          showTotal: (t) => `Total ${t} alerts`,
          showSizeChanger: true,
          onChange: (p, s) => {
            useAlertStore.setState({ page: p, size: s })
            fetchAlerts()
          },
        }}
      />
    </div>
  )
}
