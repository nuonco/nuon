import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallAction } from '@/lib'
import { ActionRunHeader } from './ActionRunHeader'
import type { IActionRunHeader } from '../types'

export const ActionRunHeaderContainer = (props: IActionRunHeader) => {
  const { org } = useOrg()
  const installId = props.step?.owner_id
  const actionWorkflowId = props.actionRun?.config?.action_workflow_id

  const { data: installAction } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['action', org?.id, actionWorkflowId],
    queryFn: () =>
      getInstallAction({
        orgId: org!.id,
        installId: installId!,
        actionId: actionWorkflowId!,
      }),
    enabled: !!org?.id && !!installId && !!actionWorkflowId && !props.isAdhoc,
  })

  return (
    <ActionRunHeader
      {...props}
      orgId={org?.id}
      actionName={installAction?.action_workflow?.name}
    />
  )
}
