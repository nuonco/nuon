import { useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from '../types'

interface StackEnvelope {
  inputs: string
  secrets: string
  spaceliftAdminTf: string
  spaceliftBlueprintYaml: string
}

function parseEnvelope(contents: unknown): StackEnvelope {
  const empty: StackEnvelope = {
    inputs: '',
    secrets: '',
    spaceliftAdminTf: '',
    spaceliftBlueprintYaml: '',
  }
  if (!contents) return empty

  let raw = contents
  if (typeof raw === 'string') {
    try {
      raw = JSON.parse(raw)
    } catch {
      try {
        raw = JSON.parse(atob(raw as string))
      } catch {
        return empty
      }
    }
  }

  if (typeof raw === 'object' && raw !== null) {
    const rec = raw as Record<string, unknown>
    return {
      inputs: String(rec.inputs_tfvars ?? ''),
      secrets: String(rec.secrets_tfvars ?? ''),
      spaceliftAdminTf: String(rec.spacelift_admin_tf ?? ''),
      spaceliftBlueprintYaml: String(rec.spacelift_blueprint_yaml ?? ''),
    }
  }

  return empty
}

interface IAwaitGCPDetails extends IStackDetails {
  installId?: string
}

export const AwaitGCPDetails = ({ stack, installId }: IAwaitGCPDetails) => {
  const version = stack?.versions?.at(0)
  const envelope = useMemo(
    () => parseEnvelope(version?.contents),
    [version?.contents]
  )
  const hasSpacelift =
    envelope.spaceliftAdminTf.length > 0 ||
    envelope.spaceliftBlueprintYaml.length > 0

  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Setup your install stack
      </Text>

      {hasSpacelift ? (
        <Tabs
          initActiveTab="terraform"
          tabs={{
            terraform: (
              <TerraformTab
                inputsTfvars={envelope.inputs}
                secretsTfvars={envelope.secrets}
                installId={installId}
              />
            ),
            spacelift: (
              <SpaceliftTab
                adminTf={envelope.spaceliftAdminTf}
                blueprintYaml={envelope.spaceliftBlueprintYaml}
                inputsTfvars={envelope.inputs}
                secretsTfvars={envelope.secrets}
              />
            ),
          }}
        />
      ) : (
        <TerraformTab
          inputsTfvars={envelope.inputs}
          secretsTfvars={envelope.secrets}
          installId={installId}
        />
      )}
    </div>
  )
}

interface ITerraformTab {
  inputsTfvars: string
  secretsTfvars: string
  installId?: string
}

const TerraformTab = ({
  inputsTfvars,
  secretsTfvars,
  installId,
}: ITerraformTab) => {
  const cloneCmd = `git clone https://github.com/nuonco/install-stacks.git
cd install-stacks/gcp`

  const backendSnippet = `terraform {
  backend "gcs" {
    bucket = "<your-state-bucket>"
    prefix = "nuon/${installId}"
  }
}`

  const applyCmd = `terraform init && terraform apply`

  return (
    <div className="flex flex-col gap-4 pt-4">
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Clone the install stack module
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Clone and enter the GCP module directory</Text>
            <ClickToCopyButton textToCopy={cloneCmd} />
          </span>
          <Code variant="preformated">{cloneCmd}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          2. Configure remote state (recommended)
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Create a <code>backend.tf</code> file to store Terraform state in
              GCS
            </Text>
            <ClickToCopyButton textToCopy={backendSnippet} />
          </span>
          <Code variant="preformated">{backendSnippet}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Save the install configuration
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>inputs.auto.tfvars</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={inputsTfvars} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(inputsTfvars, 'inputs.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{inputsTfvars}</Code>
        </Card>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>secrets.auto.tfvars</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={secretsTfvars} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(secretsTfvars, 'secrets.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{secretsTfvars}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Apply with Terraform
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Run Terraform</Text>
            <ClickToCopyButton textToCopy={applyCmd} />
          </span>
          <Code variant="preformated">{applyCmd}</Code>
        </Card>
      </div>
    </div>
  )
}

interface ISpaceliftTab {
  adminTf: string
  blueprintYaml: string
  inputsTfvars: string
  secretsTfvars: string
}

const SpaceliftTab = ({
  adminTf,
  blueprintYaml,
  inputsTfvars,
  secretsTfvars,
}: ISpaceliftTab) => {
  return (
    <div className="flex flex-col gap-4 pt-4">
      <Text variant="subtext" theme="neutral">
        Run the install stack in Spacelift instead of applying Terraform
        yourself. Both options run the same install-stacks module and mount your
        generated tfvars — pick whichever fits how you manage Spacelift.
      </Text>

      <Tabs
        initActiveTab="blueprint"
        tabs={{
          blueprint: <BlueprintSubTab blueprintYaml={blueprintYaml} />,
          'administrative stack': (
            <AdminStackSubTab
              adminTf={adminTf}
              inputsTfvars={inputsTfvars}
              secretsTfvars={secretsTfvars}
            />
          ),
        }}
      />
    </div>
  )
}

interface IAdminStackSubTab {
  adminTf: string
  inputsTfvars: string
  secretsTfvars: string
}

const AdminStackSubTab = ({
  adminTf,
  inputsTfvars,
  secretsTfvars,
}: IAdminStackSubTab) => {
  return (
    <div className="flex flex-col gap-4 pt-4">
      <Text variant="subtext" theme="neutral">
        The stack that runs this <code>spacelift.tf</code> is the administrative
        (parent) stack — its run uses the Spacelift Terraform provider to create
        the install stack and mount your tfvars.
      </Text>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Save these files together in a repo
        </Text>
        <Text variant="subtext" theme="neutral">
          Commit all three to one directory of a Git repo connected to
          Spacelift. The <code>spacelift.tf</code> reads the tfvars from its
          sibling files, so you can edit <code>inputs.auto.tfvars</code> and
          replace <code>secrets.auto.tfvars</code> with your real secret values
          before applying. Use a private repo — the secrets file is plaintext at
          rest here.
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>spacelift.tf</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={adminTf} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() => createFileDownload(adminTf, 'spacelift.tf')}
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{adminTf}</Code>
        </Card>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>inputs.auto.tfvars</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={inputsTfvars} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(inputsTfvars, 'inputs.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{inputsTfvars}</Code>
        </Card>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>secrets.auto.tfvars</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={secretsTfvars} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(secretsTfvars, 'secrets.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{secretsTfvars}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          2. Create the parent stack
        </Text>
        <Text variant="subtext" theme="neutral">
          In Spacelift, create a Terraform stack pointed at your repo. Set the{' '}
          <strong>project root</strong> to the directory holding the three
          files, pick your branch, and choose a Terraform version at or above
          the one pinned in <code>spacelift.tf</code>. See{' '}
          <Link
            href="https://docs.spacelift.io/concepts/stack/creating-a-stack"
            target="_blank"
            rel="noopener noreferrer"
          >
            Creating a stack
          </Link>
          .
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Grant it permission to manage Spacelift
        </Text>
        <Text variant="subtext" theme="neutral">
          The provider authenticates automatically — Spacelift injects{' '}
          <code>SPACELIFT_API_TOKEN</code> into runs of stacks that have a role
          attached. Open the stack&apos;s <strong>Settings → Roles</strong>,
          click <strong>Manage roles</strong>, and add the{' '}
          <strong>Space Admin</strong> role for its space. This replaces the
          deprecated Administrative flag. See{' '}
          <Link
            href="https://docs.spacelift.io/concepts/authorization/assigning-roles-stacks"
            target="_blank"
            rel="noopener noreferrer"
          >
            Assigning roles to stacks
          </Link>{' '}
          and the{' '}
          <Link
            href="https://docs.spacelift.io/vendors/terraform/terraform-provider"
            target="_blank"
            rel="noopener noreferrer"
          >
            Terraform provider
          </Link>{' '}
          docs.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Run it
        </Text>
        <Text variant="subtext" theme="neutral">
          Trigger a run on the parent stack. On apply it creates the install
          stack and mounts your tfvars. A newly created stack doesn&apos;t run
          on its own, so trigger the install stack&apos;s first run — it&apos;s
          set to auto-deploy, so it plans and applies your runner without
          further approval.
        </Text>
      </div>
    </div>
  )
}

interface IBlueprintSubTab {
  blueprintYaml: string
}

const BlueprintSubTab = ({ blueprintYaml }: IBlueprintSubTab) => {
  return (
    <div className="flex flex-col gap-4 pt-4">
      <Text variant="subtext" theme="neutral">
        A blueprint is a template Spacelift uses to create a stack. This one
        embeds your tfvars, so publishing it and creating a stack provisions the
        install with no extra configuration. See{' '}
        <Link
          href="https://docs.spacelift.io/concepts/blueprint/"
          target="_blank"
          rel="noopener noreferrer"
        >
          Blueprints
        </Link>
        .
      </Text>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Create the blueprint
        </Text>
        <Text variant="subtext" theme="neutral">
          In Spacelift, go to <strong>Blueprints → Create blueprint</strong> and
          paste this YAML as the template body. It starts as a draft you can
          edit freely.
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Blueprint template body</Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={blueprintYaml} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(blueprintYaml, 'blueprint.yaml')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{blueprintYaml}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          2. Publish it
        </Text>
        <Text variant="subtext" theme="neutral">
          Click <strong>Publish</strong> to move the blueprint from draft to
          published. The template clones the public{' '}
          <code>install-stacks</code> repository over raw Git, so no VCS
          integration setup is required. Publishing is one-way — to change a
          published blueprint you clone it, edit, and publish again.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Create a stack and fill in the inputs
        </Text>
        <Text variant="subtext" theme="neutral">
          On the published blueprint, click <strong>Create stack</strong> and
          fill in the GCP project, region, and any install inputs and secrets.
          This creates the stack but doesn&apos;t run it yet.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Attach GCP credentials, then trigger the run
        </Text>
        <Text variant="subtext" theme="neutral">
          Open the new stack&apos;s{' '}
          <strong>Settings → Integrations</strong> and attach your{' '}
          <Link
            href="https://docs.spacelift.io/integrations/cloud-providers/gcp"
            target="_blank"
            rel="noopener noreferrer"
          >
            GCP integration
          </Link>{' '}
          (its service account must already have IAM access on the target
          project). Then trigger the stack&apos;s first run — it provisions the
          runner. The run isn&apos;t triggered automatically because GCP
          credentials can&apos;t be attached from the blueprint itself.
        </Text>
      </div>
    </div>
  )
}

export const AwaitGCPDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="275px" />

      <Card>
        <Skeleton height="17px" width="250px" />
        <Skeleton height="52px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="300px" />

      <Card>
        <Skeleton height="17px" width="300px" />
        <Skeleton height="72px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="250px" />

      <Card>
        <Skeleton height="17px" width="200px" />
        <Skeleton height="100px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="200px" />

      <Card>
        <Skeleton height="17px" width="150px" />
        <Skeleton height="52px" width="100%" />
      </Card>
    </>
  )
}
