import { Navigate, Outlet } from 'react-router'
import { useAuth } from '@/hooks/use-auth'

export const AuthLayout = () => {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return null
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
