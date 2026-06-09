import { useRef, useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { CodeInput } from '@/components/common/form/CodeInput'
import { Input } from '@/components/common/form/Input'
import { WizardNavComponent } from '@/components/onboarding/WizardNav'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { runRunbook } from '@/lib'
import type { TRunbookInput } from '@/lib/ctl-api/apps/runbooks'
import type {
  TInstallRunbook,
  TRunRunbookBody,
} from '@/lib/ctl-api/installs/runbooks'

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

  const labelText = `${label}${input.required ? ' *' : ' (optional)'}`

  if (input.type === 'json') {
    return (
      <CodeInput
        name={name}
        language="json"
        defaultValue={input.default ?? ''}
        required={input.required}
        labelProps={{ labelText }}
        helperText={input.description}
      />
    )
  }

  return (
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
      labelProps={{ labelText }}
      helperText={input.description}
    />
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
  const [page, setPage] = useState<0 | 1 | 2>(0)
  const [reviewValues, setReviewValues] = useState<Record<string, string>>({})

  const runbookName = installRunbook.runbook?.name ?? 'runbook'
  const runbookId = installRunbook.runbook_id ?? installRunbook.id
  const config = installRunbook.runbook?.configs?.[0]
  const steps = (config?.steps ?? [])
    .slice()
    .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0))
  const inputs = (config?.inputs ?? [])
    .slice()
    .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0))

  const [stepEnabled, setStepEnabled] = useState<Record<string, boolean>>(() =>
    Object.fromEntries(steps.map((s) => [s.id ?? '', true]))
  )
  const isStepEnabled = (id?: string) => stepEnabled[id ?? ''] ?? true
  const enabledCount = steps.filter((s) => isStepEnabled(s.id)).length

  const { mutate, isPending } = useMutation({
    mutationFn: (body?: TRunRunbookBody) =>
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

  // View per wizard page. With inputs: 0=inputs form, 1=steps summary, 2=inputs summary (+submit).
  // Without inputs: a single steps summary (+submit).
  const showInputsForm = hasInputs && page === 0
  const showStepsSummary = !hasInputs || page === 1
  const showInputsSummary = hasInputs && page === 2
  const isSubmitView = !hasInputs || page === 2

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

    setReviewValues(collectInputs())
    setPage(1)
  }

  const handleRun = () => {
    mutate({
      ...(hasInputs ? { inputs: collectInputs() } : {}),
      steps: steps.map((s) => ({
        step_id: s.id ?? '',
        enabled: isStepEnabled(s.id),
      })),
    })
  }

  const noStepsEnabled = enabledCount === 0

  const primaryActionTrigger: IButtonAsButton = isSubmitView
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
        disabled: isPending || noStepsEnabled,
        onClick: handleRun,
        variant: 'primary',
      }
    : {
        children: 'Next',
        onClick: showInputsForm ? handleNext : () => setPage(2),
        disabled: showInputsForm ? false : noStepsEnabled,
        variant: 'primary',
      }

  const secondaryActionTrigger: IButtonAsButton | undefined =
    hasInputs && page > 0
      ? {
          children: 'Back',
          onClick: () => setPage(page === 2 ? 1 : 0),
          disabled: isPending,
          variant: 'secondary',
        }
      : undefined

  return (
    <Modal
      size="lg"
      heading={`Run ${runbookName}${isSubmitView ? '?' : ''}`}
      primaryActionTrigger={primaryActionTrigger}
      secondaryActionTrigger={secondaryActionTrigger}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {hasInputs ? (
          <WizardNavComponent
            steps={[
              { id: 'inputs', title: 'Inputs' },
              { id: 'steps', title: 'Steps' },
              { id: 'confirm', title: 'Confirm' },
            ]}
            currentStepIndex={page}
            completedSteps={
              new Set(
                ['inputs', 'steps'].slice(0, page) as string[]
              )
            }
            onboardingV2={false}
            skipHref={null}
            onGoToStep={(index) => {
              if (index <= page) setPage(index as 0 | 1 | 2)
            }}
          />
        ) : null}

        {/* Inputs form — kept mounted (hidden off the inputs page) so values persist. */}
        {hasInputs ? (
          <form
            ref={formRef}
            className={showInputsForm ? 'flex flex-col gap-4' : 'hidden'}
          >
            <Text>Provide inputs for {runbookName}:</Text>
            {inputs.map((input) => (
              <RunbookInputField key={input.id ?? input.name} input={input} />
            ))}
          </form>
        ) : null}

        {/* Steps selection page — toggle which steps run. */}
        {showStepsSummary ? (
          <Expand
            id="runbook-select-steps"
            isOpen
            heading={
              <Text weight="strong">
                Steps ({enabledCount}/{steps.length})
              </Text>
            }
          >
            <div className="flex flex-col gap-1 p-2">
              {steps.map((step, i) => (
                <div
                  key={step.id ?? i}
                  className="flex items-center justify-between gap-2"
                >
                  <CheckboxInput
                    checked={isStepEnabled(step.id)}
                    onChange={(e) =>
                      setStepEnabled((prev) => ({
                        ...prev,
                        [step.id ?? '']: e.target.checked,
                      }))
                    }
                    labelProps={{ labelText: `${i + 1}. ${step.name}` }}
                  />
                  <Badge variant="code" size="sm" theme="neutral">
                    {step.type}
                  </Badge>
                </div>
              ))}
              {noStepsEnabled ? (
                <Text variant="subtext" theme="error">
                  Enable at least one step to run the runbook.
                </Text>
              ) : null}
            </div>
          </Expand>
        ) : null}

        {/* Confirm page — inputs summary + steps summary. */}
        {showInputsSummary ? (
          <>
            <Expand
              id="runbook-review-inputs"
              isOpen
              heading={<Text weight="strong">Inputs ({inputs.length})</Text>}
            >
              <dl className="flex flex-col gap-2 p-2">
                {inputs.map((input) => {
                  const value = reviewValues[input.name] ?? ''
                  return (
                    <div
                      key={input.id ?? input.name}
                      className="grid grid-cols-2 gap-2"
                    >
                      <Text as="dt" variant="subtext">
                        {input.display_name || input.name}
                      </Text>
                      <Text as="dd" variant="body">
                        {input.sensitive
                          ? '••••••••'
                          : value !== ''
                            ? value
                            : '—'}
                      </Text>
                    </div>
                  )
                })}
              </dl>
            </Expand>

            <Expand
              id="runbook-review-steps"
              isOpen
              heading={
                <Text weight="strong">
                  Steps ({enabledCount}/{steps.length})
                </Text>
              }
            >
              <ol className="flex flex-col gap-1 p-2">
                {steps.map((step, i) => {
                  const on = isStepEnabled(step.id)
                  return (
                    <li key={step.id ?? i} className="flex items-center gap-2">
                      <Text
                        as="span"
                        variant="body"
                        className={on ? undefined : 'line-through opacity-60'}
                      >
                        {i + 1}. {step.name}
                      </Text>
                      <Badge variant="code" size="sm" theme="neutral">
                        {step.type}
                      </Badge>
                      {!on ? (
                        <Badge size="sm" theme="neutral">
                          skipped
                        </Badge>
                      ) : null}
                    </li>
                  )
                })}
              </ol>
            </Expand>
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
