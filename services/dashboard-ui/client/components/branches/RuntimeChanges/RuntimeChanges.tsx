import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Expand } from '@/components/common/Expand'
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

const imageName = (image: string) =>
  image.split('/').filter(Boolean).at(-1) ?? image

const imageRegistry = (image: string) => {
  const parts = image.split('/').filter(Boolean)
  return parts.length > 1 ? parts.slice(0, -1).join('/') : ''
}

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
        <table className="w-full text-sm table-fixed">
          <colgroup>
            <col className="w-[22%]" />
            <col className="w-[30%]" />
            <col className="w-[12%]" />
            <col className="w-[22%]" />
            <col className="w-[14%]" />
          </colgroup>
          <thead>
            <tr>
              <th className={HEADER_CLASSES}>Component</th>
              <th className={HEADER_CLASSES}>Image</th>
              <th className={HEADER_CLASSES}>Resolved tag</th>
              <th className={HEADER_CLASSES}>Digest</th>
              <th className={HEADER_CLASSES}>Outcome</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => {
              const label = row.componentName || row.buildId
              return (
                <tr key={row.buildId}>
                  <td className={CELL_CLASSES}>
                    {row.componentHref ? (
                      <Link
                        href={row.componentHref}
                        className="block truncate"
                        title={label}
                      >
                        {label}
                      </Link>
                    ) : (
                      <Text
                        variant="subtext"
                        className="block truncate"
                        title={label}
                      >
                        {label}
                      </Text>
                    )}
                  </td>
                  <td className={CELL_CLASSES}>
                    {row.image ? (
                      <div className="flex flex-col gap-0.5 min-w-0" title={row.image}>
                        <span className="truncate font-mono text-[12px] text-cool-grey-900 dark:text-cool-grey-100">
                          {imageName(row.image)}
                        </span>
                        {imageRegistry(row.image) && (
                          <span className="truncate font-mono text-[11px] text-cool-grey-500 dark:text-cool-grey-500">
                            {imageRegistry(row.image)}
                          </span>
                        )}
                      </div>
                    ) : (
                      <Text variant="subtext" theme="neutral">
                        —
                      </Text>
                    )}
                  </td>
                  <td className={CELL_CLASSES}>
                    {row.resolvedTag ? (
                      <Text
                        variant="subtext"
                        family="mono"
                        className="block truncate"
                        title={row.resolvedTag}
                      >
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
                      <span className="flex items-center gap-1 min-w-0">
                        <Text
                          variant="subtext"
                          family="mono"
                          theme="neutral"
                          className="truncate"
                          title={row.digest}
                        >
                          {row.digest}
                        </Text>
                        <ClickToCopyButton
                          textToCopy={row.digest}
                          className="shrink-0 !border-0 !p-1 text-cool-grey-400"
                        />
                      </span>
                    ) : (
                      <Text variant="subtext" theme="neutral">
                        —
                      </Text>
                    )}
                  </td>
                  <td className={CELL_CLASSES}>
                    {row.noOp ? (
                      <span
                        className="flex items-center gap-1.5 whitespace-nowrap"
                        title="Resolved digest matches the previous build; no new artifact was pushed"
                      >
                        <Icon
                          variant="MinusCircleIcon"
                          size={14}
                          className="shrink-0 text-cool-grey-400"
                        />
                        <Text variant="subtext" theme="neutral">
                          No change
                        </Text>
                      </span>
                    ) : (
                      <Status status={row.status || 'unknown'} variant="badge" />
                    )}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </Expand>
  )
}
