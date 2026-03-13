import { Navigate, Outlet } from 'react-router'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from "@/hooks/use-config"

export const AuthLayout = () => {
  const { authServiceUrl, appUrl } = useConfig()
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return null
  }

  if (!isAuthenticated) {
    return <Navigate to={`${authServiceUrl}/?url=${appUrl}`} replace />
  }

  return <Outlet />
}
