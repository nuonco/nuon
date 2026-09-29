import { useEffect, useRef } from 'react'
import { useBlocker } from 'react-router'

// A browser Back is a router POP, which would leave /onboarding. While the
// wizard still has somewhere to go, cancel that POP and step back instead.
export function useWizardHistory(
  stepIndex: number,
  onBack: () => void,
  onExit?: () => void
) {
  const holdRef = useRef(false)
  holdRef.current = stepIndex > 0 || onExit != null
  const stepRef = useRef(stepIndex)
  stepRef.current = stepIndex
  const onBackRef = useRef(onBack)
  onBackRef.current = onBack
  const onExitRef = useRef(onExit)
  onExitRef.current = onExit

  const blocker = useBlocker(
    ({ historyAction }) => historyAction === 'POP' && holdRef.current
  )

  useEffect(() => {
    if (blocker.state !== 'blocked') return
    if (stepRef.current > 0) onBackRef.current()
    else onExitRef.current?.()
    blocker.reset()
  }, [blocker])
}
