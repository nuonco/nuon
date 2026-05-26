import { useState, useRef, type KeyboardEvent } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

interface IMessageInput {
  onSend: (content: string) => void
  disabled?: boolean
}

export function MessageInput({ onSend, disabled }: IMessageInput) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleSend = () => {
    const trimmed = value.trim()
    if (!trimmed || disabled) return
    onSend(trimmed)
    setValue('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleInput = () => {
    const el = textareaRef.current
    if (el) {
      el.style.height = 'auto'
      el.style.height = Math.min(el.scrollHeight, 120) + 'px'
    }
  }

  return (
    <div className="flex items-end gap-2 border-t border-cool-grey-300 px-4 md:px-6 py-4 dark:border-dark-grey-500">
      <textarea
        ref={textareaRef}
        value={value}
        onChange={(e) => {
          setValue(e.target.value)
          handleInput()
        }}
        onKeyDown={handleKeyDown}
        placeholder="Ask agent..."
        disabled={disabled}
        rows={1}
        className="flex-1 resize-none rounded-md border border-cool-grey-300 bg-white px-3 py-2 font-sans text-sm text-cool-grey-900 outline-none shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)] placeholder:text-cool-grey-500 focus:border-primary-500 focus:ring-1 focus:ring-primary-500 disabled:opacity-50 dark:border-dark-grey-500 dark:bg-dark-grey-900 dark:text-cool-grey-100 dark:placeholder:text-cool-grey-600"
      />
      <Button
        variant="primary"
        size="lg"
        onClick={handleSend}
        disabled={disabled || !value.trim()}
        className="!p-2 justify-center"
      >
        <Icon variant="PaperPlaneTiltIcon" size={16} />
      </Button>
    </div>
  )
}
