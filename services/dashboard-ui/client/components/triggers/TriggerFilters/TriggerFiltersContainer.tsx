import { type ChangeEvent, useCallback } from 'react'
import { useSearchParams } from 'react-router'
import { TriggerFilters } from './TriggerFilters'

export const SOURCE_PARAM = 'trigger'
export const AUTH_TYPE_PARAM = 'auth_type'
export const ENVELOPE_PARAM = 'envelope'

export const TriggerFiltersContainer = () => {
  const [searchParams, setSearchParams] = useSearchParams()

  const updateParam = useCallback(
    (param: string, value: string) => {
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev)
          if (value) {
            params.set(param, value)
          } else {
            params.delete(param)
          }
          return params
        },
        { replace: true }
      )
    },
    [setSearchParams]
  )

  const handleAuthTypeChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) =>
      updateParam(AUTH_TYPE_PARAM, e.target.value),
    [updateParam]
  )

  const handleSourceChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) =>
      updateParam(SOURCE_PARAM, e.target.value),
    [updateParam]
  )

  const handleEnvelopeChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) =>
      updateParam(ENVELOPE_PARAM, e.target.value),
    [updateParam]
  )

  return (
    <TriggerFilters
      trigger={searchParams.get(SOURCE_PARAM) || ''}
      authType={searchParams.get(AUTH_TYPE_PARAM) || ''}
      envelope={searchParams.get(ENVELOPE_PARAM) || ''}
      onSourceChange={handleSourceChange}
      onAuthTypeChange={handleAuthTypeChange}
      onEnvelopeChange={handleEnvelopeChange}
      onClearSource={() => updateParam(SOURCE_PARAM, '')}
      onClearAuthType={() => updateParam(AUTH_TYPE_PARAM, '')}
      onClearEnvelope={() => updateParam(ENVELOPE_PARAM, '')}
    />
  )
}
