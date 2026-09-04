import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import type { INavItem } from '../components/molecules/NavLink'

const CHORD_TIMEOUT_MS = 1500

const keysFor = (shortcut?: string) =>
  shortcut?.trim().toLowerCase().split(/\s+/).filter(Boolean) ?? []

const editableTarget = (target: EventTarget | null) => {
  const element = target as HTMLElement | null
  return (
    element?.tagName === 'INPUT' ||
    element?.tagName === 'TEXTAREA' ||
    element?.tagName === 'SELECT' ||
    Boolean(element?.isContentEditable)
  )
}

export const useNavShortcuts = (items: INavItem[]) => {
  const navigate = useNavigate()

  useEffect(() => {
    const shortcuts = items
      .filter((item) => !item.external && keysFor(item.shortcut).length)
      .map((item) => ({ item, keys: keysFor(item.shortcut) }))
    let pending: string[] = []
    let timer: ReturnType<typeof setTimeout> | undefined

    const reset = () => {
      pending = []
      if (timer) clearTimeout(timer)
      timer = undefined
    }

    const armReset = () => {
      if (timer) clearTimeout(timer)
      timer = setTimeout(reset, CHORD_TIMEOUT_MS)
    }

    const onKeyDown = (event: KeyboardEvent) => {
      if (
        event.metaKey ||
        event.ctrlKey ||
        event.altKey ||
        editableTarget(event.target)
      ) {
        return
      }

      const key = event.key.toLowerCase()
      const next = [...pending, key]
      const exact = shortcuts.find(
        ({ keys }) =>
          keys.length === next.length &&
          keys.every((value, index) => value === next[index])
      )

      if (exact) {
        event.preventDefault()
        navigate(exact.item.href)
        reset()
        return
      }

      const prefix = shortcuts.some(
        ({ keys }) =>
          keys.length > next.length &&
          next.every((value, index) => value === keys[index])
      )
      if (prefix) {
        pending = next
        armReset()
        return
      }

      const freshPrefix = shortcuts.some(
        ({ keys }) => keys.length > 1 && keys[0] === key
      )
      pending = freshPrefix ? [key] : []
      if (freshPrefix) armReset()
      else reset()
    }

    document.addEventListener('keydown', onKeyDown)
    return () => {
      document.removeEventListener('keydown', onKeyDown)
      reset()
    }
  }, [items, navigate])
}
