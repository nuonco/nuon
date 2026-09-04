import { Outlet } from 'react-router'
import { SurfacesProvider } from '../providers/surfaces-provider'
import { ToastProvider } from '../providers/toast-provider'

export const RootLayout = () => (
  <ToastProvider>
    <SurfacesProvider>
      <Outlet />
    </SurfacesProvider>
  </ToastProvider>
)
