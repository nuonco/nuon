import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteInstallActionRunLogs } from '@/lib'
import { getAccount } from '@/lib/ctl-api/accounts'
import type { TAPIError, TInstallActionRun } from '@/types'
import {
  DeleteActionRunLogsButton as DeleteActionRunLogsButtonComponent,
  DeleteActionRunLogsModal as DeleteActionRunLogsModalComponent,
} from './DeleteActionRunLogs'

const deletableStatuses = [
  'finished',
  'error',
  'timed-out',
  'cancelled',
  'retried',
]

interface IDeleteActionRunLogs {
  actionName: string
  installActionRun: TInstallActionRun
}

export const DeleteActionRunLogsModal = ({
  actionName,
  installActionRun,
  ...props
}: IDeleteActionRunLogs & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending, error } = useMutation<unknown, TAPIError>({
    mutationFn: () =>
      deleteInstallActionRunLogs({
        orgId: org.id,
        installId: installActionRun.install_id!,
        runId: installActionRun.id!,
      }),
    onSuccess: () => {
      queryClient.removeQueries({
        queryKey: [
          'log-stream-spans',
          org?.id,
          installActionRun.log_stream?.id,
        ],
      })
      queryClient.invalidateQueries({
        queryKey: [
          'install-action-run',
          org?.id,
          installActionRun.install_id,
          installActionRun.id,
        ],
      })
      queryClient.invalidateQueries({ queryKey: ['action-run', org?.id] })
      addToast(
        <Toast heading="Logs deleted" theme="success">
          <Text>Logs and trace spans for this run were deleted.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <DeleteActionRunLogsModalComponent
      actionName={actionName}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeleteActionRunLogsButton = ({
  actionName,
  installActionRun,
}: IDeleteActionRunLogs) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const { data: account } = useQuery({
    queryKey: ['account'],
    queryFn: getAccount,
    staleTime: 60_000,
    enabled: !!org?.id,
  })
  const isOrgAdmin = account?.roles?.some(
    (role) => role.org_id === org?.id && role.role_type === 'org_admin'
  )

  const logStream = installActionRun?.log_stream
  if (!isOrgAdmin || !logStream?.id || logStream.logs_deleted_at) return null

  const isRunning = !deletableStatuses.includes(installActionRun?.status ?? '')

  return (
    <DeleteActionRunLogsButtonComponent
      disabled={isRunning}
      tooltipProps={
        isRunning
          ? { tipContent: 'Logs can be deleted once the run finishes' }
          : undefined
      }
      onOpenModal={() =>
        addModal(
          <DeleteActionRunLogsModal
            actionName={actionName}
            installActionRun={installActionRun}
          />
        )
      }
    />
  )
}
