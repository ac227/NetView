import {
  DeleteOutlined, DownloadOutlined, HeartFilled, HeartOutlined, RobotOutlined, SaveOutlined,
} from '@ant-design/icons'
import {
  Button, Card, Descriptions, Input, Popconfirm, Space, Spin, Tag, message,
} from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { itemApi } from '../api'
import type { Item } from '../types'

export default function ItemDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [item, setItem] = useState<Item | null>(null)
  const [loading, setLoading] = useState(true)
  const [title, setTitle] = useState('')
  const [desc, setDesc] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [aiLoading, setAiLoading] = useState(false)

  const load = async () => {
    if (!id) return
    setLoading(true)
    try {
      const it = await itemApi.get(Number(id))
      setItem(it)
      setTitle(it.title)
      setDesc(it.description)
      setTags(it.tags || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [id])

  if (loading) return <div style={{ textAlign: 'center', padding: 80 }}><Spin size="large" /></div>
  if (!item) return <div>条目不存在</div>

  const save = async () => {
    setSaving(true)
    try {
      await itemApi.update(item.id, { title, description: desc, tags })
      message.success('已保存')
      load()
    } catch (e: any) {
      message.error(e.response?.data?.error || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const toggleFav = async () => {
    await itemApi.favorite(item.id, !item.favorite)
    load()
  }

  const triggerDownload = async () => {
    try {
      const res = await itemApi.download(item.id)
      message.success(`已创建下载任务 #${res.job_id}`)
    } catch (e: any) {
      message.error(e.response?.data?.error || '触发失败')
    }
  }

  const doAiTag = async () => {
    setAiLoading(true)
    try {
      await itemApi.aiTag(item.id)
      message.success('AI 打标签完成')
      load()
    } catch (e: any) {
      message.error(e.response?.data?.error || 'AI 打标签失败')
    } finally {
      setAiLoading(false)
    }
  }

  const remove = async () => {
    await itemApi.remove(item.id)
    message.success('已删除')
    navigate('/')
  }

  const fileUrl = itemApi.fileUrl(item.id)

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto' }}>
      <Space style={{ marginBottom: 16 }}>
        <Button onClick={() => navigate(-1)}>返回</Button>
        <Button icon={item.favorite ? <HeartFilled /> : <HeartOutlined />} onClick={toggleFav}>
          {item.favorite ? '已收藏' : '收藏'}
        </Button>
        {item.type === 'video' && item.local_path === '' && (
          <Button type="primary" icon={<DownloadOutlined />} onClick={triggerDownload}>下载到本地</Button>
        )}
        <Button icon={<RobotOutlined />} loading={aiLoading} onClick={doAiTag}>AI 打标签</Button>
        <Popconfirm title="确定删除这条内容吗？" description="本地文件也会一并删除" onConfirm={remove} okText="删除" cancelText="取消">
          <Button danger icon={<DeleteOutlined />}>删除</Button>
        </Popconfirm>
      </Space>

      <Card>
        <div style={{ textAlign: 'center', background: '#000', borderRadius: 8, overflow: 'hidden', marginBottom: 16 }}>
          {item.type === 'video' && item.local_path !== '' ? (
            <video src={fileUrl} controls style={{ maxWidth: '100%', maxHeight: 560 }} poster={item.thumbnail_path ? itemApi.thumbUrl(item.id) : undefined} />
          ) : item.local_path !== '' ? (
            <img src={fileUrl} alt={item.title} style={{ maxWidth: '100%', maxHeight: 560 }} />
          ) : item.thumbnail_path ? (
            <img src={itemApi.thumbUrl(item.id)} alt={item.title} style={{ maxWidth: '100%', maxHeight: 560 }} />
          ) : (
            <div style={{ padding: 80, color: '#999' }}>暂无本地文件{item.source_url && `（源链接：${item.source_url}）`}</div>
          )}
        </div>

        <Input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="标题"
          style={{ fontSize: 18, fontWeight: 600, marginBottom: 8 }}
        />
        <Input.TextArea value={desc} onChange={(e) => setDesc(e.target.value)} placeholder="描述" rows={3} style={{ marginBottom: 8 }} />
        <Space wrap>
          {tags.map(t => <Tag key={t} closable onClose={() => setTags(tags.filter(x => x !== t))}>{t}</Tag>)}
          <Input
            placeholder="+ 新标签"
            style={{ width: 120 }}
            onPressEnter={(e) => {
              const v = (e.target as HTMLInputElement).value.trim()
              if (v && !tags.includes(v)) setTags([...tags, v])
              ;(e.target as HTMLInputElement).value = ''
            }}
          />
        </Space>
        <div style={{ marginTop: 16 }}>
          <Button type="primary" icon={<SaveOutlined />} loading={saving} onClick={save}>保存</Button>
        </div>

        <Descriptions size="small" column={2} style={{ marginTop: 16 }}>
          <Descriptions.Item label="类型">{item.type === 'image' ? '图片' : '视频'}</Descriptions.Item>
          <Descriptions.Item label="大小">{(item.size / 1024).toFixed(1)} KB</Descriptions.Item>
          {item.width > 0 && <Descriptions.Item label="尺寸">{item.width} × {item.height}</Descriptions.Item>}
          {item.duration > 0 && <Descriptions.Item label="时长">{Math.floor(item.duration / 60)}:{String(item.duration % 60).padStart(2, '0')}</Descriptions.Item>}
          <Descriptions.Item label="状态">{item.status}</Descriptions.Item>
          <Descriptions.Item label="添加时间">{new Date(item.created_at).toLocaleString()}</Descriptions.Item>
          {item.source_url && <Descriptions.Item label="来源" span={2}><a href={item.source_url} target="_blank" rel="noreferrer">{item.source_url}</a></Descriptions.Item>}
        </Descriptions>
      </Card>
    </div>
  )
}
