import { useCallback } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { WorkflowTypeFilter } from './WorkflowTypeFilter'

const WORKFLOW_TYPE_PARAM = 'type'

export const WorkflowTypeFilterContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const workflowType = searchParams.get(WORKFLOW_TYPE_PARAM) || ''

  const updateTypeParam = useCallback(
    (type: string) => {
      const params = new URLSearchParams(searchParams.toString())
      if (type) {
        params.set(WORKFLOW_TYPE_PARAM, type)
      } else {
        params.delete(WORKFLOW_TYPE_PARAM)
      }
      params.delete('offset')
      navigate(`?${params.toString()}`, { replace: true })
    },
    [navigate, searchParams]
  )

  const handleTypeChange = useCallback(
    (type: string) => updateTypeParam(type),
    [updateTypeParam]
  )

  const handleClearFilter = useCallback(
    () => updateTypeParam(''),
    [updateTypeParam]
  )

  return (
    <WorkflowTypeFilter
      workflowType={workflowType}
      onTypeChange={handleTypeChange}
      onClearFilter={handleClearFilter}
    />
  )
}
