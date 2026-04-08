import { useNavigate, useSearchParams } from 'react-router'
import { type ChangeEvent, useCallback } from 'react'
import { TriggeredByFilter } from './TriggeredByFilter'

const TRIGGER_PARAM = 'trigger_types'

export const TriggeredByFilterContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const triggerType = searchParams.get(TRIGGER_PARAM) || ''

  const updateTriggerParam = useCallback(
    (type: string) => {
      const params = new URLSearchParams(searchParams.toString())

      if (type) {
        params.set(TRIGGER_PARAM, type)
      } else {
        params.delete(TRIGGER_PARAM)
      }

      navigate(`?${params.toString()}`, { replace: true })
    },
    [navigate, searchParams]
  )

  const handleTriggerChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      updateTriggerParam(e.target.value)
    },
    [updateTriggerParam]
  )

  const handleClearFilter = useCallback(() => {
    updateTriggerParam('')
  }, [updateTriggerParam])

  return (
    <TriggeredByFilter
      triggerType={triggerType}
      onTriggerChange={handleTriggerChange}
      onClearFilter={handleClearFilter}
    />
  )
}
