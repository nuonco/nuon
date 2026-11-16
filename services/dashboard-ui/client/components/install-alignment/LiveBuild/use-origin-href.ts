import { useCallback } from 'react'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import type { TInstallAlignmentOrigin } from '@/types'
import { branchRunHref } from '../alignment-utils'

export const useOriginHref = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  return useCallback(
    (origin?: TInstallAlignmentOrigin) =>
      branchRunHref({
        orgId: org?.id,
        appId: install?.app_id,
        commit: origin?.run,
      }) ??
      (origin?.workflow_id
        ? installLink({
            installId: install?.id,
            appId: install?.app_id,
            suffix: `/workflows/${origin.workflow_id}`,
          })
        : undefined),
    [org?.id, install?.id, install?.app_id, installLink]
  )
}
