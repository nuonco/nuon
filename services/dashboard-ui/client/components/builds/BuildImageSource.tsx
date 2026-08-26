import { ClickToCopy } from '@/components/common/ClickToCopy'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TBuild } from '@/types'

export const BuildImageSource = ({ build }: { build: TBuild }) => (
  <div className="flex flex-col gap-4">
    <SectionHeader title="Image source" />
    <div className="grid gap-4 md:grid-cols-2">
      {build?.source_ref ? (
        <LabeledValue label="Source ref">
          <Text variant="subtext" family="mono" className="break-all">
            {build.source_ref}
          </Text>
        </LabeledValue>
      ) : null}
      {build?.resolved_tag ? (
        <LabeledValue label="Resolved tag">
          <Text variant="subtext" family="mono">
            {build.resolved_tag}
          </Text>
        </LabeledValue>
      ) : null}
      {build?.source_digest ? (
        <LabeledValue label="Digest" className="md:col-span-2">
          <ClickToCopy>
            <Text variant="subtext" family="mono" className="break-all">
              {build.source_digest}
            </Text>
          </ClickToCopy>
        </LabeledValue>
      ) : null}
      {build?.source_media_type ? (
        <LabeledValue label="Media type" className="md:col-span-2">
          <Text variant="subtext" family="mono" className="break-all">
            {build.source_media_type}
          </Text>
        </LabeledValue>
      ) : null}
      {build?.resolved_at ? (
        <LabeledValue label="Resolved">
          <Time variant="subtext" time={build.resolved_at} />
        </LabeledValue>
      ) : null}
    </div>
  </div>
)
