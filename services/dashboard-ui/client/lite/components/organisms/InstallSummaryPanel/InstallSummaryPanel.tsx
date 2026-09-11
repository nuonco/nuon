import type { TInstall } from '@/types'
import { Badge } from '../../atoms/Badge'
import { Link } from '../../atoms/Link'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { CloudPlatform } from '../../molecules/CloudPlatform'
import { CloudRegion } from '../../molecules/CloudRegion'
import { ID } from '../../molecules/ID'
import {
  installCloudLocation,
  installStatusFacets,
} from '../../../utils/install-details'
import { Panel } from '../surfaces'

export interface IInstallSummaryPanel {
  install?: TInstall
  installHref?: string
  labelColors?: Record<string, string>
  loading?: boolean
  error?: unknown
}

export const InstallSummaryPanel = ({
  install,
  installHref,
  labelColors,
  loading = false,
  error,
}: IInstallSummaryPanel) => {
  const labels = Object.entries(install?.labels ?? {}).sort(([a], [b]) =>
    a.localeCompare(b)
  )
  const location = installCloudLocation(install)
  const statuses = installStatusFacets(install)
  const hasLocation = location.region || location.location

  return (
    <Panel
      heading={
        <span className="flex min-w-0 flex-col gap-1">
          <Text as="h2" variant="heading" lines={1}>
            {loading ? 'Install' : (install?.name ?? 'Install unavailable')}
          </Text>
          {loading ? (
            <ID value="" loading />
          ) : install?.id ? (
            <ID value={install.id} />
          ) : null}
        </span>
      }
      expandable={false}
    >
      {error ? (
        <div className="flex flex-col gap-2">
          <Text as="h3" variant="heading">
            Install failed to load
          </Text>
          <Text color="secondary">
            This install could not be loaded. Try again.
          </Text>
        </div>
      ) : (
        <>
          {loading ? (
            <Text loading loadingWidth={16} />
          ) : (
            <span className="flex flex-wrap items-center gap-2">
              <CloudPlatform
                platform={location.platform}
                display="icon"
                iconSize={20}
              />
              {hasLocation ? <CloudRegion {...location} /> : null}
            </span>
          )}

          <div className="flex flex-col gap-2">
            <Text as="h3" variant="caption" color="secondary" weight="medium">
              Statuses
            </Text>
            {loading ? (
              <span className="flex flex-wrap gap-2">
                <Status loading />
                <Status loading />
                <Status loading />
              </span>
            ) : (
              <span className="flex flex-wrap gap-2">
                {statuses.map((facet) => (
                  <Status
                    key={facet.id}
                    status={facet.status}
                    icon={facet.icon}
                    label={facet.title}
                    description={facet.description}
                  />
                ))}
              </span>
            )}
          </div>

          <div className="flex flex-col gap-2">
            <Text as="h3" variant="caption" color="secondary" weight="medium">
              Labels
            </Text>
            {loading ? (
              <span className="flex flex-wrap gap-1.5">
                <Badge loading loadingWidth={12} variant="code" />
                <Badge loading loadingWidth={10} variant="code" />
              </span>
            ) : labels.length > 0 ? (
              <span className="flex flex-wrap gap-1.5">
                {labels.map(([key, value]) => (
                  <Badge
                    key={key}
                    variant="code"
                    labelKey={key}
                    labelValue={value}
                    color={labelColors?.[key]}
                  />
                ))}
              </span>
            ) : (
              <Text color="tertiary">No labels</Text>
            )}
          </div>

          {installHref ? (
            <Link href={installHref} variant="caption">
              View install
            </Link>
          ) : null}
        </>
      )}
    </Panel>
  )
}
