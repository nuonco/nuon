import { Outlet } from 'react-router'

export const AppLayout = () => {
  return (
    <div className="flex flex-col gap-4">
      <span>App layout</span>
      <Outlet />
    </div>
  )
}
