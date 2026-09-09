import { Outlet } from 'react-router'
import { AppProvider } from '../providers/app-provider'

export const AppLayout = () => (
  <AppProvider>
    <Outlet />
  </AppProvider>
)
