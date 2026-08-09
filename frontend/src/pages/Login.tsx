import { LockOutlined, PictureOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Form, Input, Typography, message, theme } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../auth'

export default function Login() {
  const { login, needsSetup } = useAuth()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const { token } = theme.useToken()

  const onFinish = async (values: { password: string }) => {
    setLoading(true)
    try {
      await login(values.password)
      message.success(needsSetup ? '密码已设置，欢迎使用 NetView' : '登录成功')
      navigate('/')
    } catch (e: any) {
      message.error(e.response?.data?.error || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: token.colorBgLayout }}>
      <Card style={{ width: 380 }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <PictureOutlined style={{ fontSize: 40, color: '#1677ff' }} />
          <Typography.Title level={3} style={{ marginTop: 12 }}>NetView</Typography.Title>
          <Typography.Text type="secondary">网络媒体收藏管理</Typography.Text>
        </div>
        {needsSetup && (
          <Alert
            type="info"
            showIcon
            message="首次使用"
            description="请设置一个访问密码，之后局域网内成员用它登录。"
            style={{ marginBottom: 16 }}
          />
        )}
        <Form onFinish={onFinish}>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder={needsSetup ? '设置访问密码' : '输入访问密码'} size="large" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block size="large" loading={loading}>
            {needsSetup ? '设置密码并进入' : '登录'}
          </Button>
        </Form>
      </Card>
    </div>
  )
}
