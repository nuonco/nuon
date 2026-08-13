import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponentBuild, TComponentType } from '@/types'

export interface IBranchRunBuilds {
  builds: TComponentBuild[]
  orgId: string
  appId: string
}

export const BranchRunBuilds = ({ builds, orgId, appId }: IBranchRunBuilds) => {
  if (builds.length === 0) return null

  return (
    <Expand
      id="run-builds"
      isOpen
      className="border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden"
      headerClassName="px-5 py-4"
      heading={
        <div className="flex items-center gap-3 w-full">
          <Text variant="h3" weight="strong">
            Builds
          </Text>
          <Text variant="subtext" theme="neutral">
            {builds.length} {builds.length === 1 ? 'component' : 'components'}
          </Text>
        </div>
      }
    >
      <div className="border-t divide-y">
        {builds.map((build) => {
          const componentType = build.component_config_connection?.type as
            | TComponentType
            | undefined
          const buildUrl =
            build.component_id && build.id
              ? `/${orgId}/apps/${appId}/components/${build.component_id}/builds/${build.id}`
              : undefined

          return (
            <div
              key={build.id}
              className="flex items-center justify-between gap-3 px-5 py-3"
            >
              <div className="flex items-center gap-3 min-w-0">
                {componentType && componentType !== 'unknown' ? (
                  <ComponentType
                    type={componentType}
                    displayVariant="icon-only"
                    colorVariant="color"
                    iconSize="16"
                  />
                ) : (
                  <Icon
                    variant="PackageIcon"
                    size={16}
                    className="shrink-0 text-cool-grey-400"
                  />
                )}
                <div className="flex flex-col gap-0.5 min-w-0">
                  <div className="flex items-center gap-2">
                    <Text variant="body" weight="strong" className="truncate">
                      {build.component_name || build.component_id}
                    </Text>
                    {componentType && componentType !== 'unknown' && (
                      <Text variant="subtext" theme="neutral">
                        {componentType.replace(/_/g, ' ')}
                      </Text>
                    )}
                  </div>
                  {build.id && (
                    <Text variant="subtext" theme="neutral" family="mono">
                      {build.id}
                    </Text>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-3 shrink-0">
                <Status
                  status={build.status_v2?.status || 'unknown'}
                  variant="badge"
                />
                {buildUrl && (
                  <Link href={buildUrl} className="text-xs">
                    View
                  </Link>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </Expand>
  )
}
