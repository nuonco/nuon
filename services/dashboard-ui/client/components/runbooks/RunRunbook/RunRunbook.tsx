import { useRef, useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { CodeInput } from '@/components/common/form/CodeInput'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { runRunbook } from '@/lib'
import type { TRunbookInput } from '@/lib/ctl-api/apps/runbooks'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'

interface IRunRunbookModal extends IModal {
  installRunbook: TInstallRunbook
}

const RunbookInputField = ({ input }: { input: TRunbookInput }) => {
  const name = `inputs:${input.name}`
  const label = input.display_name || input.name
  const isBoolean =
    input.type === 'bool' ||
    input.default === 'true' ||
    input.default === 'false'

  if (isBoolean) {
    return (
      <div className="flex flex-col gap-1">
        <input type="hidden" name={name} value="off" />
        <CheckboxInput
          name={name}
          defaultChecked={input.default === 'true'}
          labelProps={{ labelText: label }}
        />
        {input.description ? (
          <Text variant="subtext">{input.description}</Text>
        ) : null}
      </div>
    )
  }

  return (
    <label className="flex flex-col gap-1">
      <span className="flex flex-col gap-0">
        <Text variant="body" weight="strong">
          {label}{' '}
          <Text as="span" variant="subtext" theme={input.required ? 'error' : 'neutral'}>
            {input.required ? '*' : '(optional)'}
          </Text>
        </Text>
        {input.description ? (
          <Text variant="subtext">{input.description}</Text>
        ) : null}
      </span>
      {input.type === 'json' ? (
        <CodeInput
          name={name}
          language="json"
          defaultValue={input.default ?? ''}
          required={input.required}
        />
      ) : (
        <Input
          name={name}
          type={
            input.sensitive
              ? 'password'
              : input.type === 'number'
                ? 'number'
                : 'text'
          }
          defaultValue={input.default ?? ''}
          required={input.required}
        />
      )}
    </label>
  )
}

export const RunRunbookModal = ({
  installRunbook,
  ...props
}: IRunRunbookModal) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const formRef = useRef<HTMLFormElement>(null)
  const [page, setPage] = useState<0 | 1>(0)

  const runbookName = installRunbook.runbook?.name ?? 'runbook'
  const runbookId = installRunbook.runbook_id ?? installRunbook.id
  const config = installRunbook.runbook?.configs?.[0]
  const steps = config?.steps ?? []
  const inputs = (config?.inputs ?? [])
    .slice()
    .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0))

  const { mutate, isPending } = useMutation({
    mutationFn: (body?: { inputs?: Record<string, string> }) =>
      runRunbook({
        installId: install!.id,
        runbookId,
        orgId: org!.id,
        body,
      }),
    onSuccess: (result) => {
      addToast(
        <Toast heading="Runbook run started" theme="info">
          <Text>
            Running <Badge variant="code" size="md">{runbookName}</Badge>.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
      queryClient.invalidateQueries({ queryKey: ['install-runbook'] })
      const workflowId = result?.install_workflow_id
      if (workflowId) {
        navigate(`/${org!.id}/installs/${install!.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org!.id}/installs/${install!.id}/runbooks/${runbookId}`)
      }
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Runbook run failed" theme="error">
          <Text>{err?.error || `Unable to run ${runbookName}.`}</Text>
        </Toast>
      )
    },
  })

  const hasInputs = inputs.length > 0
  const onReview = !hasInputs || page === 1

  const collectInputs = (): Record<string, string> => {
    const form = formRef.current
    if (!form) return {}
    const formData = Object.fromEntries(new FormData(form))
    return Object.keys(formData).reduce(
      (acc, key) => {
        if (key.startsWith('inputs:')) {
          let value = formData[key] as string
          if (value === 'on' || value === 'off') {
            value = String(value === 'on')
          }
          acc[key.replace('inputs:', '')] = value
        }
        return acc
      },
      {} as Record<string, string>
    )
  }

  const handleNext = () => {
    const form = formRef.current
    if (!form) return

    const firstInvalid = form.querySelector<HTMLElement>(
      ':invalid:not(fieldset):not(form)'
    )
    if (firstInvalid) {
      firstInvalid.scrollIntoView({ behavior: 'smooth', block: 'center' })
      firstInvalid.focus()
      form.reportValidity()
      return
    }

    setPage(1)
  }

  const handleRun = () => {
    mutate(hasInputs ? { inputs: collectInputs() } : undefined)
  }

  const primaryActionTrigger: IButtonAsButton = onReview
    ? {
        children: isPending ? (
          <>
            <Icon variant="Loading" className="animate-spin" />
            Running...
          </>
        ) : (
          <>
            Run runbook
            <Icon variant="PlayIcon" />
          </>
        ),
        disabled: isPending,
        onClick: handleRun,
        variant: 'primary',
      }
    : {
        children: 'Next',
        onClick: handleNext,
        variant: 'primary',
      }

  return (
    <Modal
      heading={`Run ${runbookName}${onReview ? '?' : ''}`}
      primaryActionTrigger={primaryActionTrigger}
      secondaryActionTrigger={
        hasInputs && onReview
          ? {
              children: 'Back',
              onClick: () => setPage(0),
              disabled: isPending,
              variant: 'secondary',
            }
          : undefined
      }
      {...props}
    >
      <div className="flex flex-col gap-4">
        {/* Inputs page — kept mounted (hidden on review) so form values persist. */}
        {hasInputs ? (
          <form
            ref={formRef}
            className={onReview ? 'hidden' : 'flex flex-col gap-4'}
          >
            <Text>Provide inputs for {runbookName}:</Text>
            {inputs.map((input) => (
              <RunbookInputField key={input.id ?? input.name} input={input} />
            ))}
          </form>
        ) : null}

        {onReview ? (
          <>
            <Text>
              This will execute {steps.length} step
              {steps.length !== 1 ? 's' : ''} in order:
            </Text>
            <ol className="flex flex-col gap-1">
              {steps
                .slice()
                .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0))
                .map((step, i) => (
                  <li key={step.id ?? i} className="flex items-center gap-2">
                    <Text as="span" variant="body">
                      {i + 1}. {step.name}
                    </Text>
                    <Badge variant="code" size="sm" theme="neutral">
                      {step.type}
                    </Badge>
                  </li>
                ))}
            </ol>
          </>
        ) : null}
      </div>
    </Modal>
  )
}

export const RunRunbookButton = ({
  installRunbook,
  children = 'Run runbook',
  ...props
}: {
  installRunbook: TInstallRunbook
} & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RunRunbookModal installRunbook={installRunbook} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      {children} <Icon variant="PlayIcon" />
    </Button>
  )
}
