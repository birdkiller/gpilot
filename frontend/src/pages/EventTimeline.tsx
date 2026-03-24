import { useEffect, useState } from 'react'
import { Card, Timeline, Tag, Typography, Space, Spin, Select, Empty } from 'antd'
import { WarningOutlined, InfoCircleOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import client from '../api/client'

interface K8sEvent {
  id: string
  uid: string
  type: string
  reason: string
  message: string
  namespace: string
  involved_object: { kind: string; name: string }
  first_seen: string
  last_seen: string
  count: number
}

export default function EventTimeline() {
  const [events, setEvents] = useState<K8sEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [typeFilter, setTypeFilter] = useState<string | undefined>()

  useEffect(() => {
    loadEvents()
    const interval = setInterval(loadEvents, 15000)
    return () => clearInterval(interval)
  }, [])

  const loadEvents = async () => {
    try {
      const res = await client.get('/events', { params: { size: 100 } })
      setEvents(res.data.events || res.data || [])
    } catch {
      // K8s events may not be available
    } finally {
      setLoading(false)
    }
  }

  const filtered = typeFilter
    ? events.filter((e) => e.type === typeFilter)
    : events

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          K8s Event Timeline
        </Typography.Title>
        <Space>
          <Select
            placeholder="Event Type"
            style={{ width: 130 }}
            allowClear
            value={typeFilter}
            onChange={setTypeFilter}
            options={[
              { label: 'Warning', value: 'Warning' },
              { label: 'Normal', value: 'Normal' },
            ]}
          />
        </Space>
      </div>

      <Card bordered={false}>
        {filtered.length === 0 ? (
          <Empty description="No K8s events found. Connect a K8s cluster to see events here." />
        ) : (
          <Timeline
            items={filtered.map((event) => ({
              color: event.type === 'Warning' ? 'red' : 'green',
              dot: event.type === 'Warning' ? <WarningOutlined /> : <InfoCircleOutlined />,
              children: (
                <div key={event.id}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Space>
                      <Tag color={event.type === 'Warning' ? 'red' : 'green'}>
                        {event.type}
                      </Tag>
                      <Tag>{event.reason}</Tag>
                      <Tag color="blue">{event.namespace}</Tag>
                    </Space>
                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                      {dayjs(event.last_seen).format('HH:mm:ss')}
                      {event.count > 1 && ` (x${event.count})`}
                    </Typography.Text>
                  </div>
                  <Typography.Paragraph style={{ margin: '4px 0 0 0', fontSize: 13 }}>
                    {event.message}
                  </Typography.Paragraph>
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    {event.involved_object?.kind}/{event.involved_object?.name}
                  </Typography.Text>
                </div>
              ),
            }))}
          />
        )}
      </Card>
    </div>
  )
}
