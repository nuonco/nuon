import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  deleteCell,
  getCellRun,
  runCell,
  updateCell,
  type TNotebookCell,
} from '@/lib'
import type { TAPIError } from '@/types'
import { getStatusTheme } from '@/utils/status-utils'
import { NotebookCellCard } from './NotebookCellCard'
import { NotebookCellLogs } from '@/components/notebooks/NotebookCellLogs'

interface IDeleteCellModal extends IModal {
  cellName: string
  onConfirm: () => void
}

const DeleteCellModal = ({
  cellName,
  onConfirm,
  ...props
}: IDeleteCellModal) => (
  <Modal
    heading="Delete cell?"
    primaryActionTrigger={{
      children: 'Delete cell',
      onClick: onConfirm,
      variant: 'danger',
    }}
    {...props}
  >
    <Text>
      Deleting {cellName || 'this cell'} will remove its script and run
      history.
    </Text>
  </Modal>
)

interface INotebookCellCardContainer {
  cell: TNotebookCell
  notebookId: string
  index: number
}

export const NotebookCellCardContainer = ({
  cell,
  notebookId,
  index,
}: INotebookCellCardContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const { addModal } = useSurfaces()
  const queryClient = useQueryClient()

  const [name, setName] = useState(cell.name ?? '')
  const [script, setScript] = useState(cell.inline_contents ?? '')
  const [activeRunId, setActiveRunId] = useState<string | undefined>(
    cell.latest_run?.id
  )

  useEffect(() => {
    setName(cell.name ?? '')
    setScript(cell.inline_contents ?? '')
    if (cell.latest_run?.id) setActiveRunId(cell.latest_run.id)
  }, [cell.id, cell.revision])

  const isDirty =
    name !== (cell.name ?? '') || script !== (cell.inline_contents ?? '')

  const invalidateNotebook = () =>
    queryClient.invalidateQueries({
      queryKey: ['notebook', org?.id, install?.id, notebookId],
    })

  const { data: activeRun } = useQuery({
    queryKey: ['notebook-run', org?.id, install?.id, notebookId, activeRunId],
    queryFn: () =>
      getCellRun({
        orgId: org!.id,
        installId: install!.id,
        notebookId,
        runId: activeRunId!,
      }),
    enabled: !!org?.id && !!install?.id && !!activeRunId,
    refetchInterval: (query) => {
      const status =
        query.state.data?.status_v2?.status ?? query.state.data?.status
      const theme = status ? getStatusTheme(status) : 'info'
      return theme === 'success' || theme === 'error' ? false : 2000
    },
  })

  const { mutate: save, isPending: isSaving } = useMutation({
    mutationFn: () =>
      updateCell({
        orgId: org!.id,
        installId: install!.id,
        notebookId,
        cellId: cell.id,
        body: { name, inline_contents: script },
      }),
    onSuccess: () => invalidateNotebook(),
    onError: (err: TAPIError) =>
      addToast(
        <Toast heading="Cell save failed" theme="error">
          <Text>{err?.error || 'Unable to save the cell.'}</Text>
        </Toast>
      ),
  })

  const { mutate: run, isPending: isRunning } = useMutation({
    mutationFn: () =>
      runCell({
        orgId: org!.id,
        installId: install!.id,
        notebookId,
        cellId: cell.id,
      }),
    onSuccess: (newRun) => {
      setActiveRunId(newRun.id)
      addToast(
        <Toast heading="Running cell" theme="info">
          <Text>
            Running {cell.name || 'cell'}. Output will appear below.
          </Text>
        </Toast>
      )
      invalidateNotebook()
    },
    onError: (err: TAPIError) =>
      addToast(
        <Toast heading="Cell run failed" theme="error">
          <Text>{err?.error || 'Unable to run the cell.'}</Text>
        </Toast>
      ),
  })

  const { mutate: remove, isPending: isDeleting } = useMutation({
    mutationFn: () =>
      deleteCell({
        orgId: org!.id,
        installId: install!.id,
        notebookId,
        cellId: cell.id,
      }),
    onSuccess: () => invalidateNotebook(),
    onError: (err: TAPIError) =>
      addToast(
        <Toast heading="Cell deletion failed" theme="error">
          <Text>{err?.error || 'Unable to delete the cell.'}</Text>
        </Toast>
      ),
  })

  const confirmDelete = () => {
    const modal = (
      <DeleteCellModal
        cellName={name}
        onConfirm={() => remove()}
      />
    )
    addModal(modal)
  }

  const run_ = activeRun ?? cell.latest_run
  const logStreamId = run_?.log_stream_id
  const runStatus = run_?.status_v2?.status ?? run_?.status
  const runTheme = runStatus ? getStatusTheme(runStatus) : undefined
  const isRunComplete = runTheme === 'success' || runTheme === 'error'
  const runFailed = runTheme === 'error'

  return (
    <NotebookCellCard
      index={index}
      name={name}
      script={script}
      isDirty={isDirty}
      isSaving={isSaving}
      isRunning={isRunning}
      isDeleting={isDeleting}
      runStatus={runStatus}
      runStatusDescription={run_?.status_v2?.status_human_description}
      runCreatedAt={run_?.created_at}
      onNameChange={setName}
      onScriptChange={setScript}
      onSave={save}
      onRun={run}
      onDelete={confirmDelete}
      logs={
        logStreamId ? (
          <NotebookCellLogs
            logStreamId={logStreamId}
            command={run_?.command || run_?.inline_contents}
            runCreatedAt={run_?.created_at}
            runUpdatedAt={run_?.updated_at}
            isRunComplete={isRunComplete}
            runFailed={runFailed}
          />
        ) : null
      }
    />
  )
}
