import { RobotOutlined, SafetyOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, message } from 'antd'
import { useEffect, useState } from 'react'
import { settingsApi } from '../api'
import type { Settings } from '../types'

export default function Settings() {
  const [settings, setSettings] = useState<Settings | null>(null)
  const [saving, setSaving] = useState(false)
  const [aiForm] = Form.useForm()
  const [pwForm] = Form.useForm()

  useEffect(() => {
    settingsApi.get().then(setSettings).catch(() => {})
  }, [])

  useEffect(() => {
    if (settings) {
      aiForm.setFieldsValue({ ...settings.ai, api_key: '' })
    }
  }, [settings, aiForm])

  const saveAI = async (values: any) => {
    setSaving(true)
    try {
      const payload: any = { ai: { base_url: values.base_url, model: values.model } }
      if (values.api_key && values.api_key !== '••••••••') payload.ai.api_key = values.api_key
      await settingsApi.update(payload)
      message.success('AI 配置已保存')
      settingsApi.get().then(setSettings)
    } catch (e: any) {
      message.error(e.response?.data?.error || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const changePassword = async (values: any) => {
    setSaving(true)
    try {
      await settingsApi.update({ password: values.password })
      message.success('密码已修改')
      pwForm.resetFields()
    } catch (e: any) {
      message.error(e.response?.data?.error || '修改失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div style={{ maxWidth: 720, margin: '0 auto' }}>
      <Card title={<span><RobotOutlined /> AI 自动打标签</span>} style={{ marginBottom: 24 }}>
        <Form form={aiForm} layout="vertical" onFinish={saveAI}>
          <Form.Item name="base_url" label="API Base URL" extra="任意 OpenAI 兼容接口，如 https://api.openai.com/v1 或 DeepSeek/通义等地址">
            <Input placeholder="https://api.openai.com/v1" />
          </Form.Item>
          <Form.Item name="api_key" label="API Key" extra={settings?.ai.configured ? '已配置（如需更换请输入新 Key，留空保持不变）' : '尚未配置'}>
            <Input.Password placeholder="sk-..." />
          </Form.Item>
          <Form.Item name="model" label="模型名">
            <Input placeholder="gpt-4o-mini / deepseek-chat 等" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={saving}>保存 AI 配置</Button>
        </Form>
      </Card>

      <Card title={<span><SafetyOutlined /> 修改访问密码</span>}>
        <Form form={pwForm} layout="vertical" onFinish={changePassword}>
          <Form.Item name="password" label="新密码" rules={[{ required: true, min: 6, message: '至少 6 位' }]}>
            <Input.Password />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={saving}>修改密码</Button>
        </Form>
      </Card>
    </div>
  )
}
