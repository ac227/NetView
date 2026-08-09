import { HeartFilled, HeartOutlined, PlayCircleFilled } from '@ant-design/icons'
import { Card, Tag, Tooltip, Typography, theme } from 'antd'
import { useNavigate } from 'react-router-dom'
import { itemApi } from '../api'
import type { Item } from '../types'

function fmtSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

export default function ItemCard({ item, onRefresh }: { item: Item; onRefresh?: () => void }) {
  const navigate = useNavigate()
  const { token } = theme.useToken()
  const hasThumb = !!item.thumbnail_path
  const isVideo = item.type === 'video'

  const toggleFav = async (e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await itemApi.favorite(item.id, !item.favorite)
      onRefresh?.()
    } catch {
      /* ignore */
    }
  }

  return (
    <Card
      hoverable
      cover={
        <div
          onClick={() => navigate(`/items/${item.id}`)}
          style={{ height: 180, overflow: 'hidden', position: 'relative', cursor: 'pointer', background: token.colorFillTertiary, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          {hasThumb ? (
            <img
              src={itemApi.thumbUrl(item.id)}
              alt={item.title || 'item'}
              style={{ width: '100%', height: '100%', objectFit: 'cover' }}
              loading="lazy"
            />
          ) : (
            <Typography.Text type="secondary">{item.type === 'image' ? '🖼' : '🎬'} 暂无预览</Typography.Text>
          )}
          {isVideo && (
            <PlayCircleFilled style={{ position: 'absolute', fontSize: 44, color: 'rgba(255,255,255,0.85)', textShadow: '0 1px 4px rgba(0,0,0,0.4)' }} />
          )}
          <div style={{ position: 'absolute', top: 8, right: 8 }}>
            <Tooltip title={item.favorite ? '取消收藏' : '收藏'}>
              <span onClick={toggleFav} style={{ fontSize: 20, cursor: 'pointer', color: item.favorite ? '#ff4d4f' : 'rgba(255,255,255,0.9)', textShadow: '0 1px 2px rgba(0,0,0,0.5)' }}>
                {item.favorite ? <HeartFilled /> : <HeartOutlined />}
              </span>
            </Tooltip>
          </div>
          {isVideo && item.duration > 0 && (
            <span style={{ position: 'absolute', bottom: 6, right: 8, background: 'rgba(0,0,0,0.6)', color: '#fff', fontSize: 12, padding: '1px 6px', borderRadius: 4 }}>
              {Math.floor(item.duration / 60)}:{String(item.duration % 60).padStart(2, '0')}
            </span>
          )}
        </div>
      }
      onClick={() => navigate(`/items/${item.id}`)}
      styles={{ body: { padding: 12 } }}
    >
      <Card.Meta
        title={<Typography.Text ellipsis style={{ maxWidth: '100%' }}>{item.title || `未命名 ${item.type === 'image' ? '图片' : '视频'}`}</Typography.Text>}
        description={
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ color: '#999', fontSize: 12 }}>{fmtSize(item.size)}</span>
              {item.status === 'downloading' && <Tag color="processing">下载中</Tag>}
              {item.status === 'failed' && <Tag color="error">失败</Tag>}
            </div>
            {item.tags.length > 0 && (
              <div style={{ marginTop: 6 }}>
                {item.tags.slice(0, 3).map(t => <Tag key={t} style={{ fontSize: 11 }}>{t}</Tag>)}
                {item.tags.length > 3 && <Tag style={{ fontSize: 11 }}>+{item.tags.length - 3}</Tag>}
              </div>
            )}
          </div>
        }
      />
    </Card>
  )
}
