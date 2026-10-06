import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { splitChangedBuilds, type TChangedBuildRow } from './changed-builds'

const BuildRows = ({ rows }: { rows: TChangedBuildRow[] }) => (
  <ul className="flex flex-col divide-y">
    {rows.map((row) => (
      <li
        key={row.id}
        className="flex items-center justify-between gap-3 px-4 py-3"
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
)

const ChangeCard = ({
  title,
  empty,
  rows,
}: {
  title: string
  empty: string
  rows: TChangedBuildRow[]
}) => (
  <div className="border rounded-lg overflow-hidden">
    <div className="px-4 py-3 border-b">
      <Text variant="body" weight="strong">
        {title}
      </Text>
    </div>
    {rows.length === 0 ? (
      <div className="px-4 py-3">
        <Text variant="subtext" theme="neutral">
          {empty}
        </Text>
      </div>
    ) : (
      <BuildRows rows={rows} />
    )}
  </div>
)

export const BuildChangeCards = ({ rows }: { rows: TChangedBuildRow[] }) => {
  const { config, source } = splitChangedBuilds(rows)
  if (config.length === 0 && source.length === 0) {
    if (rows.length === 0) {
      return (
        <Text variant="subtext" theme="neutral">
          No builds changed.
        </Text>
      )
    }
    return <BuildRows rows={rows} />
  }

  return (
    <div className="flex flex-col gap-4">
      <ChangeCard title="Config" empty="No config changes." rows={config} />
      <ChangeCard title="Source" empty="No source changes." rows={source} />
    </div>
  )
}

export const RunBuildsPanel = ({
  rows,
  ...props
}: IPanel & { rows: TChangedBuildRow[] }) => (
  <Panel {...props} size="half" heading="Builds">
    <BuildChangeCards rows={rows} />
  </Panel>
)
