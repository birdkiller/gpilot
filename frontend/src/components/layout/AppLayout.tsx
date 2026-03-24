import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Typography, Badge } from 'antd'
import {
  DashboardOutlined,
  AlertOutlined,
  FileSearchOutlined,
  ClusterOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import { useAlertStore } from '../../store/alertStore'

const { Sider, Content, Header } = Layout

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const firingCount = useAlertStore((s) => s.firingCount)

  const menuItems = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: '/alerts',
      icon: <AlertOutlined />,
      label: (
        <span>
          告警中心
          {firingCount > 0 && (
            <Badge count={firingCount} size="small" offset={[8, -2]} />
          )}
        </span>
      ),
    },
    {
      key: '/logs',
      icon: <FileSearchOutlined />,
      label: '日志分析',
    },
    {
      key: '/events',
      icon: <ClusterOutlined />,
      label: 'K8s 事件',
    },
  ]

  const selectedKey = '/' + location.pathname.split('/')[1]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme="dark"
        width={220}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid rgba(255,255,255,0.1)',
          }}
        >
          <ThunderboltOutlined
            style={{ fontSize: 24, color: '#1677ff', marginRight: collapsed ? 0 : 8 }}
          />
          {!collapsed && (
            <Typography.Title level={4} style={{ margin: 0, color: '#fff' }}>
              gPilot
            </Typography.Title>
          )}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid #f0f0f0',
            height: 48,
          }}
        >
          <Typography.Text type="secondary" style={{ fontSize: 13 }}>
            Intelligent Alert & Log Analysis Platform
          </Typography.Text>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <Badge status={firingCount > 0 ? 'error' : 'success'} />
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {firingCount > 0 ? `${firingCount} active alerts` : 'All clear'}
            </Typography.Text>
          </div>
        </Header>
        <Content style={{ margin: 16, overflow: 'auto' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
