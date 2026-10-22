import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Input } from '@/components/common/form/Input'
import type { TAPIError } from '@/types'

interface IConfirmStep {
  installName: string
  installHref: string
  isPending: boolean
  error?: TAPIError | null
  onGoBack: () => void
  onSubmit: () => void
}

export const ConfirmStep = ({
  installName,
  installHref,
  isPending,
  error,
  onGoBack,
  onSubmit,
}: IConfirmStep) => {
  const [name, setName] = useState('')
  const matches = name === installName

  return (
    <div className="flex flex-col gap-8">
      <FormErrorBanner error={error} fallback="Unable to start teardown" />
      <Text>
        Type <Code variant="inline">{installName}</Code> to confirm teardown.
      </Text>
      <Input
        id="confirm-install-name"
        labelProps={{ labelText: 'Install name' }}
        value={name}
        autoComplete="off"
        error={name.length > 0 && !matches}
        errorMessage={
          name.length > 0 && !matches ? "Install name doesn't match" : undefined
        }
        onChange={(event) => setName(event.target.value)}
      />
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Button type="button" variant="secondary" onClick={onGoBack}>
          <Icon variant="CaretLeftIcon" weight="bold" /> Back
        </Button>
        <div className="flex flex-wrap items-center gap-3">
          <Button type="button" variant="secondary" href={installHref}>
            Cancel teardown
          </Button>
          <Button
            type="button"
            variant="danger"
            disabled={!matches || isPending}
            onClick={onSubmit}
            tooltipProps={
              !matches
                ? { tipContent: "Install name doesn't match" }
                : undefined
            }
          >
            {isPending ? (
              <>
                <Icon variant="Loading" /> Starting teardown...
              </>
            ) : (
              'Start teardown'
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}
