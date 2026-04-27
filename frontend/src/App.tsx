import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from './hooks/useAuth'
import LoginPage from './pages/LoginPage'
import PublicBoxPage from './pages/PublicBoxPage'
import InboxPage from './pages/InboxPage'

function LoadingScreen() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="rounded-2xl border border-ink-100 bg-white px-8 py-6 shadow-sm">
        <p className="text-sm text-ink-500">Loading Postcube...</p>
      </div>
    </div>
  )
}

function HomeGate() {
  const { user, loading } = useAuth()
  if (loading) return <LoadingScreen />
  return <Navigate to={user ? '/inbox' : '/login'} replace />
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <LoadingScreen />
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

function LoginGate() {
  const { user, loading } = useAuth()
  if (loading) return <LoadingScreen />
  if (user) return <Navigate to="/inbox" replace />
  return <LoginPage />
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomeGate />} />
      <Route path="/login" element={<LoginGate />} />
      <Route path="/u/:slug" element={<PublicBoxPage />} />
      <Route
        path="/inbox"
        element={
          <ProtectedRoute>
            <InboxPage />
          </ProtectedRoute>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
