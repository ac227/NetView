import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { authApi } from './api'

interface AuthState {
  authed: boolean
  needsSetup: boolean
  login: (password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthState>({ authed: false, needsSetup: false, login: async () => {}, logout: () => {} })

export function AuthProvider({ children }: { children: ReactNode }) {
  const [authed, setAuthed] = useState(!!localStorage.getItem('nv_token'))
  const [needsSetup, setNeedsSetup] = useState(false)

  useEffect(() => {
    authApi.status().then((s) => setNeedsSetup(s.needs_setup)).catch(() => {})
  }, [])

  const login = async (password: string) => {
    const res = await authApi.login(password)
    localStorage.setItem('nv_token', res.token)
    setAuthed(true)
  }

  const logout = () => {
    localStorage.removeItem('nv_token')
    setAuthed(false)
  }

  return (
    <AuthContext.Provider value={{ authed, needsSetup, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
