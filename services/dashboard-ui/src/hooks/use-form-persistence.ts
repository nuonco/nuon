import { useCallback, useEffect, useRef, useState } from 'react'

interface IFormDraft {
  values: Record<string, string>
  timestamp: string
  version: number
  configId?: string
}

interface IUseFormPersistence {
  storageKey: string
  formRef: React.RefObject<HTMLFormElement>
  enabled?: boolean
  ttlHours?: number
  configId?: string
}

interface IUseFormPersistenceReturn {
  hasDraft: boolean
  draftTimestamp: string | null
  draftValues: Record<string, string> | null
  clearDraft: () => void
  restoreDraft: () => void
  formKey: string
}

const DRAFT_VERSION = 1
const DEFAULT_TTL_HOURS = 24

export function useFormPersistence({
  storageKey,
  formRef,
  enabled = true,
  ttlHours = DEFAULT_TTL_HOURS,
  configId,
}: IUseFormPersistence): IUseFormPersistenceReturn {
  const [hasDraft, setHasDraft] = useState(false)
  const [draftTimestamp, setDraftTimestamp] = useState<string | null>(null)
  const [draftValues, setDraftValues] = useState<Record<string, string> | null>(
    null
  )
  const [shouldUseDraft, setShouldUseDraft] = useState(false)
  const [formKey, setFormKey] = useState('form-0')
  const saveTimeoutRef = useRef<NodeJS.Timeout>()

  const isLocalStorageAvailable = useCallback(() => {
    if (typeof window === 'undefined') return false
    try {
      const test = '__localStorage_test__'
      localStorage.setItem(test, test)
      localStorage.removeItem(test)
      return true
    } catch {
      return false
    }
  }, [])

  const loadDraft = useCallback(() => {
    if (!enabled || !isLocalStorageAvailable()) return null

    try {
      const stored = localStorage.getItem(storageKey)
      if (!stored) return null

      const draft: IFormDraft = JSON.parse(stored)

      if (draft.version !== DRAFT_VERSION) {
        localStorage.removeItem(storageKey)
        return null
      }

      if (configId && draft.configId !== configId) {
        localStorage.removeItem(storageKey)
        return null
      }

      const age = Date.now() - new Date(draft.timestamp).getTime()
      const maxAge = ttlHours * 60 * 60 * 1000

      if (age > maxAge) {
        localStorage.removeItem(storageKey)
        return null
      }

      return draft
    } catch (error) {
      console.warn('Failed to load form draft:', error)
      return null
    }
  }, [enabled, isLocalStorageAvailable, storageKey, configId, ttlHours])

  const saveDraft = useCallback(
    (values: Record<string, string>) => {
      if (!enabled || !isLocalStorageAvailable()) return

      try {
        const draft: IFormDraft = {
          values,
          timestamp: new Date().toISOString(),
          version: DRAFT_VERSION,
          ...(configId && { configId }),
        }

        localStorage.setItem(storageKey, JSON.stringify(draft))
      } catch (error) {
        if (error instanceof Error && error.name === 'QuotaExceededError') {
          console.warn('localStorage quota exceeded, draft not saved')
        } else {
          console.warn('Failed to save form draft:', error)
        }
      }
    },
    [enabled, isLocalStorageAvailable, storageKey, configId]
  )

  const clearDraft = useCallback(() => {
    if (!isLocalStorageAvailable()) return

    try {
      localStorage.removeItem(storageKey)
      setHasDraft(false)
      setDraftTimestamp(null)
      setDraftValues(null)
      setShouldUseDraft(false)
      setFormKey(`form-${Date.now()}`)
    } catch (error) {
      console.warn('Failed to clear form draft:', error)
    }
  }, [isLocalStorageAvailable, storageKey])

  const restoreDraft = useCallback(() => {
    if (!draftValues) return

    setShouldUseDraft(true)
    setFormKey(`form-${Date.now()}`)
  }, [draftValues])

  const getFormValues = useCallback((form: HTMLFormElement) => {
    const values: Record<string, string> = {}

    const inputs = form.querySelectorAll<
      HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
    >('input, select, textarea')

    inputs.forEach((input) => {
      if (!input.name) return

      if (input instanceof HTMLInputElement) {
        if (input.type === 'checkbox') {
          values[input.name] = input.checked.toString()
        } else if (input.type === 'radio') {
          if (input.checked) {
            values[input.name] = input.value
          }
        } else {
          values[input.name] = input.value
        }
      } else if (
        input instanceof HTMLSelectElement ||
        input instanceof HTMLTextAreaElement
      ) {
        values[input.name] = input.value
      }
    })

    return values
  }, [])

  const handleFormChange = useCallback(() => {
    if (!enabled || !formRef.current) return

    if (saveTimeoutRef.current) {
      clearTimeout(saveTimeoutRef.current)
    }

    saveTimeoutRef.current = setTimeout(() => {
      if (formRef.current) {
        const values = getFormValues(formRef.current)
        saveDraft(values)
      }
    }, 500)
  }, [enabled, formRef, getFormValues, saveDraft])

  useEffect(() => {
    const draft = loadDraft()

    if (draft) {
      setHasDraft(true)
      setDraftTimestamp(draft.timestamp)
      setDraftValues(draft.values)
    }
  }, [loadDraft])

  useEffect(() => {
    const form = formRef.current
    if (!enabled || !form) return

    const handleChange = () => {
      handleFormChange()
    }

    form.addEventListener('input', handleChange, true)
    form.addEventListener('change', handleChange, true)

    // Observe for structural changes and value attribute changes
    const observer = new MutationObserver((mutations) => {
      let shouldSave = false

      for (const mutation of mutations) {
        // Check if it's a value attribute change on an input element
        if (
          mutation.type === 'attributes' &&
          mutation.attributeName === 'value' &&
          mutation.target instanceof HTMLInputElement
        ) {
          shouldSave = true
          break
        }
        // Check if children were added/removed (Expand opening/closing)
        if (mutation.type === 'childList') {
          shouldSave = true
          break
        }
      }

      if (shouldSave) {
        handleFormChange()
      }
    })

    // Observe entire form for changes
    observer.observe(form, {
      attributes: true,
      attributeFilter: ['value'],
      childList: true,
      subtree: true,
    })

    return () => {
      form.removeEventListener('input', handleChange, true)
      form.removeEventListener('change', handleChange, true)
      observer.disconnect()
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current)
      }
    }
  }, [enabled, formRef, handleFormChange, formKey])

  return {
    hasDraft,
    draftTimestamp,
    draftValues: shouldUseDraft ? draftValues : null,
    clearDraft,
    restoreDraft,
    formKey,
  }
}
