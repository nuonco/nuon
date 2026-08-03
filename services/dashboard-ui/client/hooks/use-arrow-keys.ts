import { useEffect } from 'react'

interface IUseArrowKeysProps {
  onUpArrow: () => void
  onDownArrow: () => void
  enabled?: boolean
}

export const useArrowKeys = ({
  onUpArrow,
  onDownArrow,
  enabled = true,
}: IUseArrowKeysProps) => {
  useEffect(() => {
    if (!enabled) return

    const handleKeyDown = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement
      const isInputField =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.contentEditable === 'true'

      if (isInputField) return

      switch (event.key) {
        case 'ArrowUp':
        case 'k':
          event.preventDefault()
          onUpArrow()
          break
        case 'ArrowDown':
        case 'j':
          event.preventDefault()
          onDownArrow()
          break
        default:
          break
      }
    }

    document.addEventListener('keydown', handleKeyDown)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [onUpArrow, onDownArrow, enabled])
}
