import { Outlet } from 'react-router'
import { SurfacesProvider } from '../providers/surfaces-provider'

export const RootLayout = () => (
  <SurfacesProvider>
    <Outlet />
  </SurfacesProvider>
)
