import { Badge } from '@/components/common/Badge'
import { Expand } from '@/components/common/Expand'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

export interface IRuntimeChangeRow {
  buildId: string
  componentName?: string
  componentHref?: string
  image?: string
  resolvedTag?: string
  digest?: string
  noOp?: boolean
  status?: string
  buildHref?: string
}

export interface IRuntimeChanges {
  rows: IRuntimeChangeRow[]
}

const HEADER_CLASSES =
  'px-5 py-2.5 text-left text-[12px] font-medium text-cool-grey-500 dark:text-cool-grey-400'
const CELL_CLASSES = 'px-5 py-3 border-t border-cool-grey-100 dark:border-dark-grey-800'

export const RuntimeChanges = ({ rows }: IRuntimeChanges) => {
  if (!rows.length) return null

  return (
    <Expand
      id="runtime-changes"
      isOpen
      className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden"
      headerClassName="px-5 py-4"
      heading={
        <div className="flex items-center gap-3 w-full">
          <Text variant="h3" weight="strong">
            Runtime changes
          </Text>
          <Text variant="subtext" theme="neutral">
            {rows.length} image {rows.length === 1 ? 'component' : 'components'}
          </Text>
        </div>
      }
    >
      <div className="border-t border-cool-grey-100 dark:border-dark-grey-800 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr>
              <th className={HEADER_CLASSES}>Component</th>
              <th className={HEADER_CLASSES}>Image</th>
              <th className={HEADER_CLASSES}>Resolved tag</th>
              <th className={HEADER_CLASSES}>Digest</th>
              <th className={HEADER_CLASSES}>Outcome</th>
              <th className={HEADER_CLASSES} />
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.buildId}>
                <td className={CELL_CLASSES}>
                  {row.componentHref ? (
                    <Link href={row.componentHref}>{row.componentName || row.buildId}</Link>
                  ) : (
                    <Text variant="subtext">{row.componentName || row.buildId}</Text>
                  )}
                </td>
                <td className={`${CELL_CLASSES} max-w-[280px]`}>
                  {row.image ? (
                    <Text
                      variant="subtext"
                      family="mono"
                      className="block truncate"
                      title={row.image}
                    >
                      {row.image}
                    </Text>
                  ) : (
                    <Text variant="subtext" theme="neutral">
                      —
                    </Text>
                  )}
                </td>
                <td className={CELL_CLASSES}>
                  {row.resolvedTag ? (
                    <Text variant="subtext" family="mono">
                      {row.resolvedTag}
                    </Text>
                  ) : (
                    <Text variant="subtext" theme="neutral">
                      —
                    </Text>
                  )}
                </td>
                <td className={CELL_CLASSES}>
                  {row.digest ? (
                    <ID className="text-[12px] font-mono">{row.digest}</ID>
                  ) : (
                    <Text variant="subtext" theme="neutral">
                      —
                    </Text>
                  )}
                </td>
                <td className={CELL_CLASSES}>
                  {row.noOp ? (
                    <Badge size="sm" title="Resolved digest matches the previous build; no new artifact was pushed">
                      No change
                    </Badge>
                  ) : (
                    <Status status={row.status || 'unknown'} variant="badge" />
                  )}
                </td>
                <td className={`${CELL_CLASSES} text-right`}>
                  {row.buildHref && (
                    <Link href={row.buildHref} className="text-sm whitespace-nowrap">
                      View build
                      <Icon variant="ArrowRightIcon" size={14} />
                    </Link>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Expand>
  )
}
