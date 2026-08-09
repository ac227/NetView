import {
  CloudUploadOutlined, DownloadOutlined, LogoutOutlined, PictureOutlined, SettingOutlined,
} from '@ant-design/icons'
import { Layout, Menu, Space, Typography } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth'

const { Header, Content } = Layout

export default function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { logout } = useAuth()

  const selected = location.pathname.startsWith('/upload') ? 'upload'
    : location.pathname.startsWith('/jobs') ? 'jobs'
      : location.pathname.startsWith('/settings') ? 'settings'
        : 'gallery'

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center', gap: 16, paddingInline: 24 }}>
        <Space>
          <PictureOutlined style={{ fontSize: 22, color: '#fff' }} />
          <Typography.Text strong style={{ color: '#fff', fontSize: 17 }}>NetView</Typography.Text>
        </Space>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={[selected]}
          onClick={(e) => navigate('/' + e.key)}
          items={[
            { key: '', label: '媒体库', icon: <PictureOutlined /> },
            { key: 'upload', label: '添加', icon: <CloudUploadOutlined /> },
            { key: 'jobs', label: '下载任务', icon: <DownloadOutlined /> },
            { key: 'settings', label: '设置', icon: <SettingOutlined /> },
          ]}
          style={{ flex: 1, minWidth: 0 }}
        />
        <LogoutOutlined
          style={{ color: '#fff', fontSize: 18, cursor: 'pointer' }}
          onClick={() => { logout(); navigate('/login') }}
        />
      </Header>
      <Content style={{ padding: 24, maxWidth: 1400, width: '100%', margin: '0 auto' }}>
        <Outlet />
      </Content>
    </Layout>
  )
}
