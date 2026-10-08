import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { statusSlices, type TStatusInstall } from './install-status'

export const InstallStatusCounts = ({
  installs,
  showTotal = true,
}: {
  installs: TStatusInstall[]
  showTotal?: boolean
}) => {
  const slices = statusSlices(installs)
  const total = installs.length

  return (
    <>
      {showTotal ? (
        <Text variant="subtext" theme="neutral">
          {total} {total === 1 ? 'install' : 'installs'}
        </Text>
      ) : null}
      {slices.map((slice) => (
        <Status key={slice.status} status={slice.status}>
          <span className="font-normal">
            {slice.count} {slice.label}
          </span>
        </Status>
      ))}
    </>
  )
}
