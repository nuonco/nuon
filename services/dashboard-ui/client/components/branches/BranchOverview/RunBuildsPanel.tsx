import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TChangedBuildRow } from './changed-builds'

export const RunBuildsPanel = ({
  rows,
  ...props
}: IPanel & { rows: TChangedBuildRow[] }) => (
  <Panel {...props} size="half" heading="Builds">
    {rows.length === 0 ? (
      <Text variant="subtext" theme="neutral">
        No builds changed.
      </Text>
    ) : (
      <ul className="flex flex-col divide-y">
        {rows.map((row) => (
          <li
            key={row.id}
            className="flex items-center justify-between gap-3 py-3"
          >
            <span className="flex min-w-0 flex-col gap-1">
              <Text variant="body" weight="strong" className="truncate">
                {row.name}
              </Text>
              <Status status={row.status} />
            </span>
            {row.href ? <Link href={row.href}>View build</Link> : null}
          </li>
        ))}
      </ul>
    )}
  </Panel>
)
