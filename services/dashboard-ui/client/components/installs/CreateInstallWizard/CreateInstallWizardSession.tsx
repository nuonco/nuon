import { useEffect, useRef, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useStore } from '@tanstack/react-form'
import {
  appBranchGroupFromSelection,
  labelsFromGroupSelection,
  type TGroupStepSelection,
} from '@/components/installs/CreateInstall/GroupStep'
import {
  buildCreateInstallBody,
  normalizeInstallPlatform,
  useInstallForm,
  type InstallFormValues,
} from '@/components/installs/forms/InstallForm'
import { ResumeDraftModal } from '@/components/installs/forms/shared/ResumeDraftModal'
import { useAuth } from '@/hooks/use-auth'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createAppInstall, installNameTaken } from '@/lib'
import { trackEvent } from '@/lib/posthog-analytics'
import type {
  TAPIError,
  TApp,
  TAppBranch,
  TAppBranchConfig,
  TAppInputConfig,
  TAction,
  TCloudConnectionSummary,
  TComponent,
} from '@/types'
import { CreateInstallWizard } from './CreateInstallWizard'
import { clearGroupDraft, loadGroupDraft, saveGroupDraft } from './group-draft'
import { AppStep } from './steps/AppStep'
import { DetailsStep } from './steps/DetailsStep'
import { InputsStep } from './steps/InputsStep'
import { ProvisionStep } from './steps/ProvisionStep'
import { ReviewStep } from './steps/ReviewStep'
import {
  assignableGroupLabels,
  assignmentBlocker,
  enrollmentBlocker,
  mergeCarriedLabels,
} from './readiness'
import {
  CREATE_WIZARD_STEPS,
  createWizardStatuses,
} from './wizard-path'

interface ICreatedInstall {
  id: string
  name: string
  workflowId?: string
  onboarding: boolean
}

const toApiError = (err: unknown): TAPIError => {
  const error = (err ?? {}) as Partial<TAPIError>
  return {
    error: error.error || 'Unable to create install.',
    description: error.description || '',
    user_error: !!error.user_error,
    status: error.status,
  }
}

interface ICreateInstallWizardSession {
  app: TApp
  branch: TAppBranch
  groupConfig?: TAppBranchConfig
  inputConfig: TAppInputConfig
  actions: TAction[]
  components: TComponent[]
  cloudConnections?: TCloudConnectionSummary[]
  requireTargetAccount: boolean
  installsPath: string
}

export const CreateInstallWizardSession = ({
  app,
  branch,
  groupConfig,
  inputConfig,
  actions,
  components,
  cloudConnections,
  requireTargetAccount,
  installsPath,
}: ICreateInstallWizardSession) => {
  const { org } = useOrg()
  const { user } = useAuth()
  const { addModal, removeModal } = useSurfaces()
  const queryClient = useQueryClient()
  const installLink = useInstallLink()
  const platform = normalizeInstallPlatform(app.runner_config?.app_runner_type)
  const draftShownRef = useRef(false)
  const carriedLabelsRef = useRef<Record<string, string>>({})
  const [farthest, setFarthest] = useState(0)
  const [showing, setShowing] = useState(0)
  const [created, setCreated] = useState<ICreatedInstall>()
  const [selection, setSelection] = useState<TGroupStepSelection>({
    mode: 'none',
    group: null,
  })

  const provisionIndex = CREATE_WIZARD_STEPS.length - 1
  const locked = farthest === provisionIndex
  const storageKey = `install-draft:${app.id}`

  const { mutateAsync, isPending, error } = useMutation({
    mutationFn: (body: ReturnType<typeof buildCreateInstallBody>) =>
      createAppInstall({ appId: app.id, body, orgId: org.id }),
    onSuccess: (result) => {
      trackEvent({
        event: 'install_create',
        status: 'ok',
        user,
        props: { appId: app.id, installId: result.data.id },
      })
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
    },
    onError: (err: TAPIError) => {
      trackEvent({
        event: 'install_create',
        status: 'error',
        user,
        props: { appId: app.id, err: err?.error },
      })
    },
  })

  const { form, canSubmit, hasDraft, draftTimestamp, clearDraft, restoreDraft } =
    useInstallForm({
      mode: 'create',
      platform,
      inputConfig,
      requireTargetAccount,
      storageKey,
      onSubmit: async (values) => {
        const base = buildCreateInstallBody(values, platform)
        const result = await mutateAsync({
          ...base,
          labels: {
            ...base.labels,
            ...labelsFromGroupSelection(selection),
          },
          app_branch_id: branch.id,
          app_branch_group: appBranchGroupFromSelection(selection),
        })
        clearDraft()
        clearGroupDraft(storageKey)
        try {
          localStorage.removeItem(storageKey)
        } catch (storageError) {
          console.warn('Failed to clear form draft:', storageError)
        }
        setCreated({
          id: result.data.id,
          name: result.data.name ?? values.name.trim(),
          workflowId: result.data.workflow_id,
          onboarding: result.data.install_number === 1,
        })
        setFarthest(provisionIndex)
        setShowing(provisionIndex)
      },
    })

  useEffect(() => {
    if (!hasDraft || draftShownRef.current || !draftTimestamp) return
    draftShownRef.current = true

    let modalId: string
    const modal = (
      <ResumeDraftModal
        draftTimestamp={draftTimestamp}
        onResume={() => {
          restoreDraft()
          const restored = loadGroupDraft(storageKey, branch.id, groupConfig)
          if (restored) {
            carriedLabelsRef.current =
              restored.mode === 'labels'
                ? assignableGroupLabels(restored.group)
                : {}
            setSelection(restored)
          }
          removeModal(modalId)
        }}
        onStartFresh={() => {
          clearDraft()
          clearGroupDraft(storageKey)
          draftShownRef.current = false
          removeModal(modalId)
        }}
        onClose={() => removeModal(modalId)}
      />
    )
    modalId = addModal(modal)
  }, [
    hasDraft,
    draftTimestamp,
    restoreDraft,
    clearDraft,
    addModal,
    removeModal,
  ])

  const validateName = async (name: string) => {
    const trimmed = name.trim()
    if (!trimmed || !org.id) return undefined
    try {
      const taken = await installNameTaken({
        appId: app.id,
        orgId: org.id,
        name: trimmed,
      })
      return taken ? `An install named "${trimmed}" already exists` : undefined
    } catch (err) {
      console.warn('Unable to check install name availability:', err)
      return undefined
    }
  }

  const advance = () => {
    if (showing < farthest) {
      setShowing(farthest)
      return
    }
    if (showing >= provisionIndex) return
    const next = showing + 1
    setFarthest(next)
    setShowing(next)
  }

  const onGoBack =
    showing > 0 && !locked ? () => setShowing((current) => current - 1) : undefined

  const selectGroup = (next: TGroupStepSelection) => {
    const carried =
      next.mode === 'labels' ? assignableGroupLabels(next.group) : {}
    form.setFieldValue(
      'labels',
      mergeCarriedLabels({
        labels: form.state.values.labels,
        previous: carriedLabelsRef.current,
        next: carried,
      })
    )
    carriedLabelsRef.current = carried
    setSelection(next)
    saveGroupDraft(storageKey, branch.id, next)
  }

  const step = CREATE_WIZARD_STEPS[showing]
  const statuses = createWizardStatuses(farthest)
  const values = useStore(form.store, (state) => state.values) as InstallFormValues
  const installHref = created
    ? installLink({
        orgId: org.id,
        appId: app.id,
        installId: created.id,
        suffix: created.onboarding ? '?onboardingComplete=true' : '',
      })
    : ''

  return (
    <form
      autoComplete="off"
      noValidate
      onSubmit={(event) => event.preventDefault()}
      className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-6 pt-8 pb-8"
    >
      <CreateInstallWizard
        steps={CREATE_WIZARD_STEPS.map((item, index) => ({
          id: item.id,
          label: item.navLabel,
          status: statuses[index],
        }))}
        selectedId={step.id}
        title={step.title}
        description={step.description}
        onSelect={(id) => {
          if (locked) return
          const index = CREATE_WIZARD_STEPS.findIndex((item) => item.id === id)
          if (index !== -1 && index <= farthest) setShowing(index)
        }}
      >
        {step.id === 'app' ? (
          <AppStep
            appName={app.name}
            branchName={branch.name}
            config={groupConfig}
            selected={selection}
            labels={values.labels}
            onSelect={selectGroup}
            onAdvance={advance}
            cancelHref={installsPath}
          />
        ) : null}
        {step.id === 'details' ? (
          <DetailsStep
            form={form}
            platform={platform}
            cloudConnections={cloudConnections}
            groupConfig={groupConfig}
            selection={selection}
            requireTargetAccount={requireTargetAccount}
            validateName={validateName}
            onAdvance={advance}
            onGoBack={onGoBack}
            cancelHref={installsPath}
          />
        ) : null}
        {step.id === 'inputs' ? (
          <InputsStep
            form={form}
            inputConfig={inputConfig}
            onAdvance={advance}
            onGoBack={onGoBack}
            cancelHref={installsPath}
          />
        ) : null}
        {step.id === 'review' ? (
          <ReviewStep
            form={form}
            platform={platform}
            appName={app.name ?? ''}
            branchName={branch.name ?? ''}
            selection={selection}
            inputConfig={inputConfig}
            actions={actions}
            components={components}
            cloudConnections={cloudConnections}
            submitError={error ? toApiError(error) : null}
            submitting={isPending}
            canSubmit={
              canSubmit &&
              !isPending &&
              !enrollmentBlocker(groupConfig?.install_groups?.length ?? 0, selection) &&
              !assignmentBlocker(selection, values.labels, groupConfig)
            }
            disabledReason={
              enrollmentBlocker(
                groupConfig?.install_groups?.length ?? 0,
                selection
              ) ?? assignmentBlocker(selection, values.labels, groupConfig)
            }
            onAdvance={() => {
              void form.handleSubmit()
            }}
            onGoBack={onGoBack}
            cancelHref={installsPath}
          />
        ) : null}
        {step.id === 'provision' && created ? (
          <ProvisionStep
            installName={created.name || values.name}
            workflowId={created.workflowId}
            installHref={installHref}
          />
        ) : null}
      </CreateInstallWizard>
    </form>
  )
}
