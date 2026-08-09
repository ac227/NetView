import { LinkOutlined, UploadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Form, Input, Radio, Space, Tabs, Upload, message } from 'antd'
import type { RcFile } from 'antd/es/upload'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { itemApi } from '../api'

export default function AddContent() {
  const navigate = useNavigate()
  const [tab, setTab] = useState('link')
  const [fileType, setFileType] = useState<'auto' | 'image' | 'video'>('auto')
  const [uploading, setUploading] = useState(false)
  const [creating, setCreating] = useState(false)

  const beforeUpload = (file: RcFile) => {
    const type = file.type.startsWith('image/') ? 'image' : 'video'
    setUploading(true)
    itemApi.upload(file, type)
      .then(() => { message.success('上传成功'); navigate('/') })
      .catch((e: any) => message.error(e.response?.data?.error || '上传失败'))
      .finally(() => setUploading(false))
    return false
  }

  const submitLink = async (values: { source_url: string; title?: string; description?: string }) => {
    setCreating(true)
    try {
      let meta = { title: values.title || '', description: values.description || '' }
      try {
        const m = await itemApi.fetchMeta(values.source_url)
        meta.title = meta.title || m.title
        meta.description = meta.description || m.description
      } catch { /* 抓取失败不影响创建 */ }
      const item = await itemApi.create({
        source_url: values.source_url,
        title: meta.title,
        description: meta.description,
        type: fileType === 'auto' ? undefined : fileType,
      })
      message.success('已保存链接')
      navigate(`/items/${item.id}`)
    } catch (e: any) {
      message.error(e.response?.data?.error || '创建失败')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div style={{ maxWidth: 720, margin: '0 auto' }}>
      <Card>
        <Tabs activeKey={tab} onChange={setTab} items={[
          {
            key: 'link',
            label: <span><LinkOutlined /> 粘贴链接</span>,
            children: (
              <>
                <Alert type="info" showIcon style={{ marginBottom: 16 }}
                  message="支持：图片/视频直链（.jpg/.png/.mp4 等），或其他网页链接（视频将可用 yt-dlp 下载）。" />
                <Form layout="vertical" onFinish={submitLink} initialValues={{ source_url: '' }}>
                  <Form.Item name="source_url" label="链接地址" rules={[{ required: true, message: '请输入链接' }]}>
                    <Input placeholder="https://..." />
                  </Form.Item>
                  <Form.Item label="类型">
                    <Radio.Group value={fileType} onChange={(e) => setFileType(e.target.value)}>
                      <Radio.Button value="auto">自动</Radio.Button>
                      <Radio.Button value="image">图片</Radio.Button>
                      <Radio.Button value="video">视频</Radio.Button>
                    </Radio.Group>
                  </Form.Item>
                  <Form.Item name="title" label="标题（可选）"><Input placeholder="留空则自动抓取" /></Form.Item>
                  <Form.Item name="description" label="描述（可选）"><Input.TextArea rows={2} /></Form.Item>
                  <Button type="primary" htmlType="submit" loading={creating} icon={<LinkOutlined />}>保存链接</Button>
                </Form>
              </>
            ),
          },
          {
            key: 'file',
            label: <span><UploadOutlined /> 上传文件</span>,
            children: (
              <>
                <Alert type="info" showIcon style={{ marginBottom: 16 }} message="支持拖拽或点击选择，图片会自动生成缩略图，视频会提取首帧作封面。" />
                <Upload.Dragger
                  multiple
                  beforeUpload={beforeUpload}
                  showUploadList={false}
                  accept="image/*,video/*"
                  disabled={uploading}
                >
                  <p className="ant-upload-drag-icon"><UploadOutlined /></p>
                  <p className="ant-upload-text">{uploading ? '上传中...' : '点击或拖拽文件到此处'}</p>
                  <p className="ant-upload-hint">支持常见图片和视频格式</p>
                </Upload.Dragger>
              </>
            ),
          },
        ]} />
        <Space style={{ marginTop: 16 }}>
          <Button onClick={() => navigate(-1)}>返回</Button>
        </Space>
      </Card>
    </div>
  )
}
