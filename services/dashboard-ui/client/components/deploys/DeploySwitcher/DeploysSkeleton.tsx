import { Text } from '@/components/common/Text'

export const DeploysSkeleton = ({ limit = 5 }: { limit?: number }) => {
  return Array.from({ length: limit }).map((_, idx) => (
    <span
      key={`deploy-skeleton-${idx}`}
      className="flex flex-col w-full gap-1 rounded-lg border p-2"
    >
      <span className="flex items-center justify-between">
        <Text variant="subtext" loading loadingWidth={26} />
        <Text variant="label" loading loadingWidth={8} />
      </span>
      <span className="flex items-center gap-4 w-full">
        <Text variant="subtext" loading loadingWidth={8} />
        <Text variant="label" loading loadingWidth={11} />
      </span>
    </span>
  ))
}
