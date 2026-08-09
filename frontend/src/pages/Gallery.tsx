import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import { Col, Empty, Input, Pagination, Row, Segmented, Select, Space, Spin } from 'antd'
import { useCallback, useEffect, useRef, useState } from 'react'
import { itemApi, tagApi } from '../api'
import ItemCard from '../components/ItemCard'
import type { ListParams } from '../types'

export default function Gallery() {
  const [items, setItems] = useState([] as any[])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [type, setType] = useState('all')
  const [favOnly, setFavOnly] = useState(false)
  const [tags, setTags] = useState([] as string[])
  const [tag, setTag] = useState<string | undefined>(undefined)
  const [refreshKey, setRefreshKey] = useState(0)
  const pageSize = 24
  const keywordRef = useRef('')

  useEffect(() => {
    tagApi.list().then(setTags).catch(() => {})
  }, [])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const params: ListParams = {
        page,
        page_size: pageSize,
        sort: 'newest',
        keyword: keywordRef.current,
      }
      if (type !== 'all') params.type = type
      if (favOnly) params.favorite = true
      if (tag) params.tag = tag
      const res = await itemApi.list(params)
      setItems(res.items)
      setTotal(res.total)
    } catch {
      /* ignore */
    } finally {
      setLoading(false)
    }
  }, [page, type, favOnly, tag, refreshKey])

  useEffect(() => { load() }, [load])

  return (
    <div>
      <Space wrap style={{ marginBottom: 16, width: '100%', justifyContent: 'space-between' }}>
        <Space wrap>
          <Input.Search
            placeholder="搜索标题 / 描述"
            allowClear
            style={{ width: 260 }}
            onSearch={(v) => { keywordRef.current = v; setPage(1); setRefreshKey(k => k + 1) }}
            prefix={<SearchOutlined />}
          />
          <Select
            placeholder="按标签筛选"
            allowClear
            showSearch
            style={{ width: 180 }}
            options={tags.map(t => ({ value: t, label: t }))}
            value={tag}
            onChange={(v) => { setTag(v); setPage(1) }}
          />
          <Segmented
            value={type}
            onChange={(v) => { setType(v as string); setPage(1) }}
            options={[
              { label: '全部', value: 'all' },
              { label: '图片', value: 'image' },
              { label: '视频', value: 'video' },
            ]}
          />
          <Segmented
            value={favOnly ? 'fav' : 'allfav'}
            onChange={(v) => { setFavOnly(v === 'fav'); setPage(1) }}
            options={[{ label: '全部', value: 'allfav' }, { label: '收藏', value: 'fav' }]}
          />
          <ReloadOutlined style={{ cursor: 'pointer' }} onClick={() => setRefreshKey(k => k + 1)} />
        </Space>
      </Space>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 60 }}><Spin size="large" /></div>
      ) : items.length === 0 ? (
        <Empty description="还没有内容，点击右上角「添加」导入图片或视频" style={{ padding: 60 }} />
      ) : (
        <>
          <Row gutter={[16, 16]}>
            {items.map(item => (
              <Col key={item.id} xs={12} sm={8} md={6} lg={6} xl={4}>
                <ItemCard item={item} onRefresh={() => setRefreshKey(k => k + 1)} />
              </Col>
            ))}
          </Row>
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Pagination current={page} total={total} pageSize={pageSize} onChange={setPage} showTotal={(t) => `共 ${t} 条`} />
          </div>
        </>
      )}
    </div>
  )
}
