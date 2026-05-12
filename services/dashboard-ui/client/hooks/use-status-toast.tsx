import { useEffect, useRef, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { getStatusTheme } from '@/utils/status-utils'

export function useStatusToast({
  status,
  label,
  resourceType,
}: {
  status: string | undefined
  label?: string
  resourceType: string
}) {
  const { addToast } = useToast()
  const firedRef = useRef(false)
  const hasSeenNonTerminalRef = useRef(false)

  useEffect(() => {
    if (firedRef.current || !status) return

    const theme = getStatusTheme(status)
    if (theme !== 'success' && theme !== 'error') {
      hasSeenNonTerminalRef.current = true
      return
    }

    if (!hasSeenNonTerminalRef.current) return

    firedRef.current = true
    const outcome = theme === 'success' ? 'succeeded' : 'failed'

    const heading: ReactNode = label ? (
      <span className="inline-flex items-center gap-1.5">
        <Badge variant="code" size="md">{label}</Badge> {resourceType} {outcome}
      </span>
    ) : (
      `${resourceType} ${outcome}`
    )

    addToast(<Toast heading={heading} theme={theme} />)
  }, [status])
}
