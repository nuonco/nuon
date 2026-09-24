import { Outlet } from 'react-router'
import { PageTitleProvider } from '../providers/page-title-provider'
import { SurfacesProvider } from '../providers/surfaces-provider'
import { ToastProvider } from '../providers/toast-provider'

export const RootLayout = () => (
  <PageTitleProvider>
    <ToastProvider>
      <SurfacesProvider>
        <Outlet />
      </SurfacesProvider>
    </ToastProvider>
  </PageTitleProvider>
)
