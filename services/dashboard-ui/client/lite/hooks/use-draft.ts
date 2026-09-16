import { useCallback, useMemo, useState } from 'react'
import { clearDraft, loadDraft, migrateDraft, saveDraft } from '../utils/draft'

export interface IUseDraftOptions {
  orgId?: string
  wizard: string
  resourceId?: string
}

export const useDraft = <T>({
  orgId,
  wizard,
  resourceId,
}: IUseDraftOptions) => {
  const [generation, setGeneration] = useState(0)
  const [decision, setDecision] = useState<'pending' | 'resume' | 'fresh'>(
    'pending'
  )

  const stored = useMemo(() => {
    void generation
    if (!orgId) return undefined
    if (resourceId) migrateDraft(orgId, wizard, resourceId)
    return loadDraft<T>(orgId, wizard, resourceId)
  }, [generation, orgId, resourceId, wizard])

  const bump = () => setGeneration((value) => value + 1)

  const resume = useCallback(() => {
    setDecision('resume')
  }, [])

  const startFresh = useCallback(() => {
    if (orgId) clearDraft(orgId, wizard, resourceId)
    setDecision('fresh')
    bump()
  }, [orgId, resourceId, wizard])

  const save = useCallback(
    (values: T) => {
      if (!orgId) return
      saveDraft(orgId, wizard, values, resourceId)
      setDecision('resume')
      bump()
    },
    [orgId, resourceId, wizard]
  )

  const clear = useCallback(() => {
    if (!orgId) return
    clearDraft(orgId, wizard, resourceId)
    setDecision('fresh')
    bump()
  }, [orgId, resourceId, wizard])

  const resumeOpen = Boolean(stored) && decision === 'pending'

  return {
    values: decision === 'resume' ? stored?.values : undefined,
    updatedAt: stored?.updatedAt,
    resumeOpen,
    resume,
    startFresh,
    save,
    clear,
  }
}
