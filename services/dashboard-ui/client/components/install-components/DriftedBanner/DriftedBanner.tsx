import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import type { TDriftedObject } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

const DRIFTED_KIND = {
  install_deploy: 'component',
  install_sandbox_run: 'sandbox',
}

interface IDriftedBanner {
  drifted: TDriftedObject
  orgId: string
  installId: string
}

export const DriftedBanner = ({ drifted, orgId, installId }: IDriftedBanner) => {
  return (
    <Banner theme="warn">
      <div className="flex items-center gap-8">
        <div className="flex flex-col max-w-86">
          <Text weight="strong" variant="base">
            {toSentenceCase(DRIFTED_KIND[drifted?.target_type])} drift detected
          </Text>
          <Text className="text-pretty" theme="neutral">
            This {DRIFTED_KIND[drifted?.target_type]} has drifted from the
            desired state. Review the changes to understand what's changed.
          </Text>
        </div>
        <Button
          className="ml-auto"
          href={`/${orgId}/installs/${installId}/workflows/${drifted?.install_workflow_id}`}
          variant="primary"
        >
          View details
        </Button>
      </div>
    </Banner>
  )
}
