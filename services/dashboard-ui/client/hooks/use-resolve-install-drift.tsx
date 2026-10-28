import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { INSTALL_OVERVIEW_QUERY_KEY } from '@/hooks/use-install-overview'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { resolveInstallDrift } from '@/lib'
import type { TAPIError, TInstallAlignmentResource } from '@/types'

export type TResolveInstallDriftInput = {
  resource: TInstallAlignmentResource
  approvedPlanId?: string
}

const INVALIDATED_QUERY_KEYS = [
  INSTALL_OVERVIEW_QUERY_KEY,
  'install-workflows',
  'install-active-workflows',
  'active-workflows',
  'workflow-approvals',
  'install-activity',
  'install-deployments',
  'install-resource-components',
  'install-resource-images',
  'install-components',
  'install-component',
  'install-sandbox-runs',
  'install-stack',
  'install-status',
]

// Request IDs survive remounts so retrying after an uncertain failure (timeout,
// 5xx, conflict) reuses the same ID and the server can deduplicate it.
const requestIds = new Map<string, string>()

const requestKey = (
  installId: string,
  { resource, approvedPlanId }: TResolveInstallDriftInput
) => `${installId}:${resource.id}:${approvedPlanId ?? ''}`

const isDefinitiveRejection = (error: TAPIError) =>
  !!error?.status &&
  error.status >= 400 &&
  error.status < 500 &&
  error.status !== 408 &&
  error.status !== 409

export const resolveDriftRequestId = (
  installId: string,
  input: TResolveInstallDriftInput
) => {
  const key = requestKey(installId, input)
  let requestId = requestIds.get(key)
  if (!requestId) {
    requestId = crypto.randomUUID()
    requestIds.set(key, requestId)
  }
  return requestId
}

export const useResolveInstallDrift = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const installLink = useInstallLink()
  const { addToast } = useToast()

  return useMutation({
    mutationFn: (input: TResolveInstallDriftInput) =>
      resolveInstallDrift({
        orgId: org.id,
        installId: install.id,
        body: {
          resource_id: input.resource.id!,
          request_id: resolveDriftRequestId(install.id, input),
          ...(input.approvedPlanId
            ? { approved_plan_id: input.approvedPlanId }
            : {}),
        },
      }),
    onSuccess: (result, input) => {
      requestIds.delete(requestKey(install.id, input))
      for (const key of INVALIDATED_QUERY_KEYS) {
        void queryClient.invalidateQueries({ queryKey: [key] })
      }
      const workflowId = result?.workflow_id
      addToast(
        <Toast heading="Resolving drift" theme="info">
          <Text>
            Resolving drift for <Text family="mono">{input.resource.name}</Text>
            . This may take a few minutes.
          </Text>
        </Toast>
      )
      navigate(
        installLink({
          installId: install.id,
          appId: install.app_id,
          suffix: workflowId ? `/workflows/${workflowId}` : '/workflows',
        })
      )
    },
    onError: (error, input) => {
      if (isDefinitiveRejection(error)) {
        requestIds.delete(requestKey(install.id, input))
      }
    },
  })
}
