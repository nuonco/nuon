import { Outlet } from 'react-router'
import { useRoutePrefetch } from '../hooks/use-route-prefetch'
import { PageTitleProvider } from '../providers/page-title-provider'
import { SurfacesProvider } from '../providers/surfaces-provider'
import { ToastProvider } from '../providers/toast-provider'

export const RootLayout = () => {
  useRoutePrefetch()

  return (
    <PageTitleProvider>
      <ToastProvider>
        <SurfacesProvider>
          <Outlet />
        </SurfacesProvider>
      </ToastProvider>
    </PageTitleProvider>
  )
}
