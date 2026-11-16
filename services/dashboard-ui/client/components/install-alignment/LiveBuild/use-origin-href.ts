import { useInstallPage } from '@/hooks/use-install-path'
import type { TInstallAlignmentOrigin } from '@/types'
import { branchRunHref } from '../alignment-utils'

export const useOriginHref = () => {
  const { org, appId, href } = useInstallPage()
  return (origin?: TInstallAlignmentOrigin) =>
    branchRunHref({ orgId: org?.id, appId, commit: origin?.run }) ??
    (origin?.workflow_id ? href(`/workflows/${origin.workflow_id}`) : undefined)
}
