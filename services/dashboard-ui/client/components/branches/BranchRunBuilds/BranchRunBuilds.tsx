import { useMemo } from 'react'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import {
  BuildTypeFilter,
  SANDBOX_FILTER,
  uniqueBuildFilterTypes,
  useBuildTypeFilter,
} from '@/components/branches/BuildTypeFilter'
import type { TAppSandboxBuild, TComponentBuild, TComponentType } from '@/types'

export interface IBranchRunBuilds {
  builds: TComponentBuild[]
  sandboxBuild?: TAppSandboxBuild | null
  orgId: string
  appId: string
}

const componentTypeOf = (build: TComponentBuild): TComponentType | undefined =>
  build.component_config_connection?.type as TComponentType | undefined

export const BranchRunBuilds = ({
  builds,
  sandboxBuild,
  orgId,
  appId,
}: IBranchRunBuilds) => {
  const types = useMemo(
    () =>
      uniqueBuildFilterTypes([
        ...builds.map(componentTypeOf),
        sandboxBuild ? SANDBOX_FILTER : undefined,
      ]),
    [builds, sandboxBuild]
  )
  const filter = useBuildTypeFilter(types)

  const visibleBuilds = builds.filter((build) =>
    filter.matches(componentTypeOf(build))
  )
  const showSandbox = !!sandboxBuild && filter.matches(SANDBOX_FILTER)
  const visibleCount = visibleBuilds.length + (showSandbox ? 1 : 0)
  const totalCount = builds.length + (sandboxBuild ? 1 : 0)

  if (totalCount === 0) return null

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
            {totalCount} {totalCount === 1 ? 'build' : 'builds'}
          </Text>
        </div>
      }
    >
      <div className="border-t">
        {types.length > 1 && (
          <div className="px-5 py-2">
            <BuildTypeFilter
              types={filter.types}
              deselected={filter.deselected}
              onToggle={filter.toggle}
            />
          </div>
        )}
        <div className={types.length > 1 ? 'divide-y border-t' : 'divide-y'}>
          {visibleCount === 0 ? (
            <div className="px-5 py-3">
              <Text variant="subtext" theme="neutral">
                No builds match filters
              </Text>
            </div>
          ) : (
            <>
              {visibleBuilds.map((build) => {
                const componentType = componentTypeOf(build)
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
                          <Text
                            variant="body"
                            weight="strong"
                            className="truncate"
                          >
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
                      {buildUrl && <Link href={buildUrl}>View build</Link>}
                    </div>
                  </div>
                )
              })}
              {showSandbox && sandboxBuild && (
                <div className="flex items-center justify-between gap-3 px-5 py-3">
                  <div className="flex items-center gap-3 min-w-0">
                    <Icon
                      variant="TerminalWindowIcon"
                      size={16}
                      className="shrink-0 text-cool-grey-400"
                    />
                    <div className="flex flex-col gap-0.5 min-w-0">
                      <div className="flex items-center gap-2">
                        <Text
                          variant="body"
                          weight="strong"
                          className="truncate"
                        >
                          Sandbox
                        </Text>
                      </div>
                      {sandboxBuild.id && (
                        <Text variant="subtext" theme="neutral" family="mono">
                          {sandboxBuild.id}
                        </Text>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-3 shrink-0">
                    <Status
                      status={
                        sandboxBuild.status_v2?.status ||
                        sandboxBuild.status ||
                        'unknown'
                      }
                      variant="badge"
                    />
                    {sandboxBuild.id && (
                      <Link
                        href={`/${orgId}/apps/${appId}/sandbox/builds/${sandboxBuild.id}`}
                      >
                        View build
                      </Link>
                    )}
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </Expand>
  )
}
