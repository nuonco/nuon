import type { ReactNode } from 'react'
import { Outlet } from 'react-router'
import { useModules } from '../providers/modules-provider'
import type { TModuleId } from '../utils/modules'
import { NotFound } from './scaffolds'

export interface IModuleLayout {
  module: TModuleId | readonly TModuleId[]
  fallback?: ReactNode
}

export const ModuleLayout = ({
  module,
  fallback = <NotFound />,
}: IModuleLayout) => {
  const { has, ready } = useModules()
  const ids = typeof module === 'string' ? [module] : module

  if (ready && !ids.some(has)) return <>{fallback}</>
  return <Outlet />
}
