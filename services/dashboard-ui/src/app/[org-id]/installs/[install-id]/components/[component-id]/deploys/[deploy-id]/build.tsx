import { Duration, StatusBadge, Text, Time } from '@/components'
import { getComponentBuildById } from '@/lib'

export const Build = async ({
  buildId,
  componentId,
  orgId,
}: {
  buildId: string
  componentId: string
  orgId: string
}) => {
  const { data: build, error } = await getComponentBuildById({
    buildId,
    componentId,
    orgId,
  })

  return build && !error ? (
    <div className="flex items-start justify-start gap-6">
      <span className="flex flex-col gap-2">
        <StatusBadge
          description={build.status_description}
          status={build.status}
          label="Status"
        />
      </span>

      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Build date
        </Text>
        <Time time={build.created_at} />
      </span>

      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Build duration
        </Text>
        <Duration beginTime={build.created_at} endTime={build.updated_at} />
      </span>
    </div>
  ) : (
    <Text>No component build found.</Text>
  )
}
