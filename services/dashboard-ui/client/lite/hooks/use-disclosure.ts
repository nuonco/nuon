import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from 'react'

export interface IDisclosureGroup {
  defaultOpen: boolean
  register: (id: string, setOpen: (open: boolean) => void) => () => void
  report: (id: string, open: boolean) => void
  openAll: () => void
  closeAll: () => void
  allOpen: boolean
  anyOpen: boolean
  count: number
}

export const DisclosureGroupContext = createContext<IDisclosureGroup | null>(
  null
)

export const useDisclosureGroup = () => useContext(DisclosureGroupContext)

export const useDisclosureGroupState = (
  defaultOpen: boolean
): IDisclosureGroup => {
  const members = useRef(new Map<string, (open: boolean) => void>())
  const [states, setStates] = useState<Record<string, boolean>>({})

  const register = useCallback(
    (id: string, setOpen: (open: boolean) => void) => {
      members.current.set(id, setOpen)
      return () => {
        members.current.delete(id)
        setStates((current) => {
          if (!(id in current)) return current
          const next = { ...current }
          delete next[id]
          return next
        })
      }
    },
    []
  )

  const report = useCallback((id: string, open: boolean) => {
    setStates((current) =>
      current[id] === open ? current : { ...current, [id]: open }
    )
  }, [])

  const setAll = useCallback((open: boolean) => {
    for (const setOpen of members.current.values()) setOpen(open)
  }, [])

  return useMemo(() => {
    const values = Object.values(states)
    return {
      defaultOpen,
      register,
      report,
      openAll: () => setAll(true),
      closeAll: () => setAll(false),
      allOpen: values.length > 0 && values.every(Boolean),
      anyOpen: values.some(Boolean),
      count: values.length,
    }
  }, [defaultOpen, register, report, setAll, states])
}

export interface IUseDisclosure {
  id?: string
  open?: boolean
  defaultOpen?: boolean
  onOpenChange?: (open: boolean) => void
}

export const useDisclosure = ({
  id: idProp,
  open: controlledOpen,
  defaultOpen,
  onOpenChange,
}: IUseDisclosure = {}) => {
  const group = useDisclosureGroup()
  const generatedId = useId()
  const id = idProp ?? generatedId

  const isControlled = controlledOpen !== undefined
  const [uncontrolledOpen, setUncontrolledOpen] = useState(
    defaultOpen ?? group?.defaultOpen ?? false
  )
  const open = isControlled ? controlledOpen : uncontrolledOpen

  const setOpen = useCallback(
    (next: boolean) => {
      if (!isControlled) setUncontrolledOpen(next)
      onOpenChange?.(next)
    },
    [isControlled, onOpenChange]
  )

  const openRef = useRef(open)
  openRef.current = open

  const toggle = useCallback(() => setOpen(!openRef.current), [setOpen])

  const register = group?.register
  const report = group?.report

  useEffect(() => register?.(id, setOpen), [register, id, setOpen])
  useEffect(() => report?.(id, open), [report, id, open])

  const triggerId = `${id}-trigger`
  const contentId = `${id}-content`

  return {
    open,
    setOpen,
    toggle,
    triggerProps: {
      id: triggerId,
      'aria-expanded': open,
      'aria-controls': contentId,
      onClick: toggle,
    },
    contentProps: {
      id: contentId,
      role: 'region' as const,
      'aria-labelledby': triggerId,
    },
  }
}
