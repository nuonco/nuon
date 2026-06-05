import { useEffect, useRef } from 'react'
import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { getStatusTheme } from '@/utils/status-utils'
import { toSentenceCase } from '@/utils/string-utils'

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
    const heading = `${toSentenceCase(resourceType)} ${outcome}`

    addToast(
      <Toast heading={heading} theme={theme}>
        {label ? <Text><Badge variant="code" size="md">{label}</Badge></Text> : null}
      </Toast>
    )
  }, [status])
}
