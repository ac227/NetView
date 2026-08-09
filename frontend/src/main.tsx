import { StrictMode, useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { ConfigProvider, theme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import 'antd/dist/reset.css'
import './index.css'
import App from './App'

function applyBodyTheme(dark: boolean) {
  const el = document.documentElement
  el.style.colorScheme = dark ? 'dark' : 'light'
  el.style.background = dark ? '#141414' : '#f5f5f5'
}

function Root() {
  const [dark, setDark] = useState(
    () => typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches,
  )

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const onChange = (e: MediaQueryListEvent) => {
      setDark(e.matches)
      applyBodyTheme(e.matches)
    }
    mq.addEventListener('change', onChange)
    applyBodyTheme(mq.matches)
    return () => mq.removeEventListener('change', onChange)
  }, [])

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{ algorithm: dark ? theme.darkAlgorithm : theme.defaultAlgorithm }}
    >
      <App />
    </ConfigProvider>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Root />
  </StrictMode>,
)
