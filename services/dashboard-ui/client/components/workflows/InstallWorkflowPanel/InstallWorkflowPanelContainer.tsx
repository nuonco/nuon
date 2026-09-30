import {
  useCallback,
  useEffect,
  useRef,
  type ComponentProps,
  type ReactNode,
} from 'react'
import { useLocation, useSearchParams } from 'react-router'
import { Link } from '@/components/common/Link'
import type { IPanel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { WorkflowProvider } from '@/providers/workflow-provider'
import {
  InstallWorkflowPanel,
  InstallWorkflowPanelContent,
} from './InstallWorkflowPanel'

const WORKFLOW_PARAM = 'workflow'

const WorkflowPanel = ({
  workflowId,
  onClose,
  ...props
}: {
  workflowId: string
  onClose: () => void
} & Partial<IPanel>) => (
  <InstallWorkflowPanel onClose={onClose} {...props}>
    <WorkflowProvider workflowId={workflowId} shouldPoll>
      <InstallWorkflowPanelContent />
    </WorkflowProvider>
  </InstallWorkflowPanel>
)

export const InstallWorkflowPanelController = () => {
  const { addPanel, removePanel, panels } = useSurfaces()
  const [searchParams, setSearchParams] = useSearchParams()
  const workflowId = searchParams.get(WORKFLOW_PARAM)
  const panelIdRef = useRef<string | null>(null)
  const workflowIdRef = useRef<string | null>(null)
  const panelIsOpen = panels.some(
    (panel) => panel.id === panelIdRef.current && panel.isVisible
  )

  const clearWorkflow = useCallback(() => {
    panelIdRef.current = null
    workflowIdRef.current = null
    setSearchParams(
      (current) => {
        const next = new URLSearchParams(current)
        next.delete(WORKFLOW_PARAM)
        return next
      },
      { replace: true }
    )
  }, [setSearchParams])

  useEffect(() => {
    if (!workflowId) {
      if (panelIdRef.current && panelIsOpen) {
        removePanel(panelIdRef.current)
      }
      panelIdRef.current = null
      workflowIdRef.current = null
      return
    }

    if (panelIsOpen && workflowIdRef.current === workflowId) return

    if (panelIdRef.current && panelIsOpen) {
      removePanel(panelIdRef.current)
    }

    workflowIdRef.current = workflowId
    panelIdRef.current = addPanel(
      <WorkflowPanel workflowId={workflowId} onClose={clearWorkflow} />
    )
  }, [workflowId, panelIsOpen, addPanel, removePanel, clearWorkflow])

  return null
}

export const WorkflowPanelLink = ({
  workflowId,
  children,
  ...props
}: {
  workflowId: string
  children: ReactNode
} & Omit<ComponentProps<typeof Link>, 'href'>) => {
  const { pathname } = useLocation()
  const [searchParams] = useSearchParams()
  const next = new URLSearchParams(searchParams)
  next.set(WORKFLOW_PARAM, workflowId)

  return (
    <Link href={`${pathname}?${next.toString()}`} {...props}>
      {children}
    </Link>
  )
}
