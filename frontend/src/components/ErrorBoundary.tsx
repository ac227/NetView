import { Alert, Button } from 'antd'
import { Component, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  render() {
    if (this.state.error) {
      return (
        <div style={{ padding: 48, display: 'flex', justifyContent: 'center' }}>
          <Alert
            type="error"
            showIcon
            message="页面出错了"
            description={String(this.state.error?.message || this.state.error)}
            action={
              <Button size="small" danger onClick={() => { this.setState({ error: null }) }}>
                重试
              </Button>
            }
          />
        </div>
      )
    }
    return this.props.children
  }
}
