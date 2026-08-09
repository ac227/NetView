import { StopOutlined } from '@ant-design/icons'
import { Button, Card, Progress, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { jobApi } from '../api'
import type { Job } from '../types'

const statusColor: Record<string, string> = {
  pending: 'default',
  running: 'processing',
  done: 'success',
  failed: 'error',
  cancelled: 'warning',
}

const statusText: Record<string, string> = {
  pending: '排队中',
  running: '下载中',
  done: '完成',
  failed: '失败',
  cancelled: '已取消',
}

export default function Jobs() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      setJobs(await jobApi.list())
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    const timer = setInterval(load, 5000)
    return () => clearInterval(timer)
  }, [])

  const cancel = async (job: Job) => {
    await jobApi.cancel(job.id)
    message.success('已取消')
    load()
  }

  return (
    <Card title="下载任务">
      <Table
        rowKey="id"
        loading={loading}
        dataSource={jobs}
        size="small"
        pagination={false}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 70 },
          {
            title: '地址',
            dataIndex: 'url',
            ellipsis: true,
            render: (u: string) => <span title={u}>{u}</span>,
          },
          { title: '方式', dataIndex: 'adapter', width: 100, render: (a: string) => <Tag>{a === 'yt-dlp' ? 'yt-dlp' : '直链'}</Tag> },
          {
            title: '状态',
            dataIndex: 'status',
            width: 110,
            render: (s: string) => (
              <Tag color={statusColor[s] || 'default'}>{statusText[s] || s}</Tag>
            ),
          },
          {
            title: '进度',
            dataIndex: 'progress',
            width: 180,
            render: (p: number, j: Job) =>
              j.status === 'running' || j.status === 'done'
                ? <Progress percent={Math.round(p * 100)} size="small" />
                : <span style={{ color: '#999' }}>—</span>,
          },
          {
            title: '',
            width: 90,
            render: (_, job) =>
              (job.status === 'pending' || job.status === 'running') && (
                <Button size="small" danger icon={<StopOutlined />} onClick={() => cancel(job)}>取消</Button>
              ),
          },
        ]}
      />
    </Card>
  )
}
