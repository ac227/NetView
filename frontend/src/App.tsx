import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from './auth'
import AppLayout from './components/AppLayout'
import ErrorBoundary from './components/ErrorBoundary'
import Login from './pages/Login'
import Gallery from './pages/Gallery'
import ItemDetail from './pages/ItemDetail'
import Upload from './pages/Upload'
import Jobs from './pages/Jobs'
import Settings from './pages/Settings'

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { authed } = useAuth()
  if (!authed) return <Navigate to="/login" replace />
  return <>{children}</>
}

function AppRoutes() {
  const { authed } = useAuth()
  return (
    <Routes>
      <Route path="/login" element={authed ? <Navigate to="/" replace /> : <Login />} />
      <Route element={<RequireAuth><AppLayout /></RequireAuth>}>
        <Route path="/" element={<Gallery />} />
        <Route path="/items/:id" element={<ItemDetail />} />
        <Route path="/upload" element={<Upload />} />
        <Route path="/jobs" element={<Jobs />} />
        <Route path="/settings" element={<Settings />} />
      </Route>
    </Routes>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <ErrorBoundary>
          <AppRoutes />
        </ErrorBoundary>
      </BrowserRouter>
    </AuthProvider>
  )
}
