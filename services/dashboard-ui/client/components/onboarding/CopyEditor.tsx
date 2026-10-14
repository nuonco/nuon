import { useEffect, useRef, useState, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Toggle } from '@/components/common/form/Toggle'

// Story-only helper: makes the story's text editable in place and exports the
// edits, so copy can be reviewed in Ladle without touching code.
interface ICopyEdit {
  from: string
  to: string
  tag: string
}

interface ICopyOriginal {
  own: string
  full: string
}

const COPY_EDITOR_SKIP = new Set(['SCRIPT', 'STYLE', 'SVG', 'PATH', 'INPUT', 'TEXTAREA', 'SELECT', 'OPTION'])

const normalizeCopy = (value: string) => value.replace(/\s+/g, ' ').trim()

const ownCopy = (el: Element) =>
  normalizeCopy(
    Array.from(el.childNodes)
      .filter((node) => node.nodeType === Node.TEXT_NODE)
      .map((node) => node.nodeValue ?? '')
      .join('')
  )

const elementOf = (node: Node | null | undefined): Element | null => {
  if (!node) return null
  return node.nodeType === Node.TEXT_NODE ? node.parentElement : (node as Element)
}

export const CopyEditor = ({ children, tools }: { children: ReactNode; tools?: ReactNode }) => {
  const wrapRef = useRef<HTMLDivElement>(null)
  const tracked = useRef(new Set<Element>())
  const originals = useRef(new WeakMap<Element, ICopyOriginal>())
  const live = useRef(new Map<Element, ICopyEdit & { toOwn: string }>())
  const committed = useRef<ICopyEdit[]>([])
  const [editing, setEditing] = useState(false)
  const [count, setCount] = useState(0)
  const [copied, setCopied] = useState(false)

  const snapshot = (el: Element) => {
    originals.current.set(el, { own: ownCopy(el), full: normalizeCopy(el.textContent ?? '') })
    tracked.current.add(el)
  }

  const remember = (el: Element) => {
    if (COPY_EDITOR_SKIP.has(el.tagName) || originals.current.has(el)) return
    if (ownCopy(el)) snapshot(el)
  }

  const resync = (el: Element, root: HTMLElement) => {
    const original = originals.current.get(el)
    if (!original) return
    const edit = live.current.get(el)
    const detached = !root.contains(el)
    const rewritten = ownCopy(el) !== (edit?.toOwn ?? original.own)
    if (!detached && !rewritten) return
    if (edit) {
      const { toOwn: _toOwn, ...done } = edit
      committed.current.push(done)
      live.current.delete(el)
    }
    if (detached) {
      tracked.current.delete(el)
      originals.current.delete(el)
    } else {
      snapshot(el)
    }
  }

  const resyncAll = (root: HTMLElement) => {
    Array.from(tracked.current).forEach((el) => resync(el, root))
    setCount(committed.current.length + live.current.size)
  }

  useEffect(() => {
    const root = wrapRef.current?.firstElementChild as HTMLElement | null
    if (!root || !editing) return

    const lineage = (node: Node | null | undefined) => {
      const out: Element[] = []
      let el = elementOf(node)
      while (el && el !== root) {
        out.push(el)
        el = el.parentElement
      }
      return out
    }

    const swallowClicks = (event: MouseEvent) => {
      if ((event.target as Element).closest('button, a, summary, [role="button"]')) {
        event.preventDefault()
        event.stopPropagation()
      }
    }

    const guardInput = (event: InputEvent) => {
      const selection = document.getSelection()
      if (
        event.inputType.startsWith('format') ||
        event.inputType === 'insertParagraph' ||
        event.inputType === 'insertLineBreak'
      ) {
        event.preventDefault()
        return
      }
      if (
        selection &&
        !selection.isCollapsed &&
        elementOf(selection.anchorNode) !== elementOf(selection.focusNode)
      ) {
        event.preventDefault()
        return
      }
      lineage(selection?.anchorNode).forEach((el) => {
        remember(el)
        resync(el, root)
      })
    }

    const pastePlain = (event: ClipboardEvent) => {
      event.preventDefault()
      const text = event.clipboardData?.getData('text/plain') ?? ''
      document.execCommand('insertText', false, normalizeCopy(text))
    }

    const record = () => {
      const selection = document.getSelection()
      const touched = new Set<Element>([
        ...lineage(selection?.anchorNode),
        ...lineage(selection?.focusNode),
      ])
      touched.forEach((el) => {
        const original = originals.current.get(el)
        if (!original) return
        const toOwn = ownCopy(el)
        if (toOwn === original.own) {
          live.current.delete(el)
        } else {
          live.current.set(el, {
            from: original.full,
            to: normalizeCopy(el.textContent ?? ''),
            tag: el.tagName.toLowerCase(),
            toOwn,
          })
        }
      })
      setCount(committed.current.length + live.current.size)
    }

    resyncAll(root)
    root.querySelectorAll('*').forEach(remember)
    root.contentEditable = 'true'
    root.spellcheck = false
    root.style.cursor = 'text'
    root.addEventListener('click', swallowClicks, true)
    root.addEventListener('beforeinput', guardInput)
    root.addEventListener('paste', pastePlain)
    root.addEventListener('input', record)

    return () => {
      root.removeEventListener('click', swallowClicks, true)
      root.removeEventListener('beforeinput', guardInput)
      root.removeEventListener('paste', pastePlain)
      root.removeEventListener('input', record)
      root.removeAttribute('contenteditable')
      root.removeAttribute('spellcheck')
      root.style.cursor = ''
    }
  }, [editing])

  const collect = () => {
    const root = wrapRef.current?.firstElementChild as HTMLElement | null
    if (root) resyncAll(root)
    const current = Array.from(live.current.values()).map(({ toOwn: _toOwn, ...edit }) => edit)
    return JSON.stringify([...committed.current, ...current], null, 2)
  }

  const copyEdits = () => {
    navigator.clipboard?.writeText(collect()).catch(() => {})
    setCopied(true)
    window.setTimeout(() => setCopied(false), 2000)
  }

  const downloadEdits = () => {
    const url = URL.createObjectURL(new Blob([collect()], { type: 'application/json' }))
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = 'nuon-copy-edits.json'
    anchor.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 1000)
  }

  const clearEdits = () => {
    live.current.clear()
    committed.current = []
    tracked.current.clear()
    originals.current = new WeakMap()
    setCount(0)
  }

  const nothingYet = count === 0
  const editsLabel = `${count} ${count === 1 ? 'edit' : 'edits'}`

  return (
    <>
      <div ref={wrapRef} className="contents">
        {children}
      </div>
      <Card className="fixed bottom-4 left-1/2 z-50 max-w-[calc(100vw-2rem)] -translate-x-1/2 !flex-row flex-wrap items-center justify-center !gap-3 !p-3 bg-background">
        {tools ? (
          <>
            {tools}
            <span className="h-5 w-px bg-neutral-200 dark:bg-neutral-600" aria-hidden />
          </>
        ) : null}
        <Toggle checked={editing} onChange={setEditing} label={editing ? 'Editing copy' : 'Edit copy'} />
        <Button
          variant="secondary"
          size="sm"
          onClick={copyEdits}
          disabled={nothingYet}
          tooltipProps={nothingYet ? { tipContent: 'Cannot copy edits — no text changed yet' } : undefined}
        >
          <Icon variant={copied ? 'CheckIcon' : 'CopyIcon'} weight="bold" />
          {copied ? 'Copied' : `Copy ${editsLabel}`}
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={downloadEdits}
          disabled={nothingYet}
          tooltipProps={nothingYet ? { tipContent: 'Cannot download edits — no text changed yet' } : undefined}
        >
          <Icon variant="DownloadSimpleIcon" weight="bold" />
          Download
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={clearEdits}
          disabled={nothingYet}
          tooltipProps={nothingYet ? { tipContent: 'Cannot clear edits — nothing recorded yet' } : undefined}
        >
          Clear
        </Button>
      </Card>
    </>
  )
}
