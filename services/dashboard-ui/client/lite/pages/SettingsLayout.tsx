import { Outlet } from 'react-router'
import { SubNav, type ISubNavItem } from '../components/molecules/SubNav'
import { useModules } from '../providers/modules-provider'
import { useOrg } from '../providers/org-provider'
import {
  ALL_MODULES,
  MODULES,
  moduleHref,
  type TModuleId,
} from '../utils/modules'

export const settingsNavigation = (
  orgId: string,
  enabled: ReadonlySet<TModuleId> = ALL_MODULES
): ISubNavItem[] =>
  MODULES.flatMap((module) =>
    module.settings && enabled.has(module.id)
      ? [
          {
            href: moduleHref(orgId, module),
            label: module.settings.label,
            end: module.settings.path === '',
          },
        ]
      : []
  )

export const SettingsLayout = () => {
  const { orgId } = useOrg()
  const { enabled } = useModules()

  return (
    <div className="flex w-full flex-col gap-6">
      <SubNav
        items={settingsNavigation(orgId ?? '', enabled)}
        label="Settings sections"
      />
      <Outlet />
    </div>
  )
}
