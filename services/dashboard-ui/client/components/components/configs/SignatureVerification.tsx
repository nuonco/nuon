import { Badge } from '@/components/common/Badge'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import type { TSignatureAuthority, TSignatureVerification } from '@/types'

const AUTHORITY_LABELS: Record<string, string> = {
  keyless: 'Keyless',
  public_key: 'Public key',
}

const SignatureAuthority = ({
  authority,
}: {
  authority: TSignatureAuthority
}) => {
  const label = authority?.type
    ? (AUTHORITY_LABELS[authority.type] ?? authority.type)
    : 'Unknown'

  return (
    <div className="flex flex-col gap-3 rounded border p-4">
      <Badge size="sm" theme="default" className="w-fit">
        {label}
      </Badge>

      {authority?.issuer ? (
        <LabeledValue label="Issuer">
          <Text variant="subtext" family="mono" className="break-all">
            {authority.issuer}
          </Text>
        </LabeledValue>
      ) : null}

      {authority?.subject ? (
        <LabeledValue label="Subject">
          <Text variant="subtext" family="mono" className="break-all">
            {authority.subject}
          </Text>
        </LabeledValue>
      ) : null}

      {authority?.subject_regexp ? (
        <LabeledValue label="Subject pattern">
          <Text variant="subtext" family="mono" className="break-all">
            {authority.subject_regexp}
          </Text>
        </LabeledValue>
      ) : null}

      {authority?.public_key ? (
        <LabeledValue
          label={
            <span className="flex items-center justify-between gap-2">
              <Text variant="subtext" theme="neutral">
                Public key
              </Text>
              <ClickToCopyButton textToCopy={authority.public_key} />
            </span>
          }
        >
          <Code variant="preformated" className="text-xs">
            {authority.public_key}
          </Code>
        </LabeledValue>
      ) : null}
    </div>
  )
}

export const SignatureVerification = ({
  verification,
}: {
  verification?: TSignatureVerification
}) => {
  if (!verification) return null

  const authorities = verification.authorities ?? []

  return (
    <div className="flex flex-col gap-3 pt-6 border-t">
      <span className="flex items-center gap-2">
        <Text variant="body" weight="strong" level={5}>
          Image signature verification
        </Text>
        <Badge
          size="sm"
          theme={verification.require_signature ? 'success' : 'neutral'}
        >
          {verification.require_signature ? 'Required' : 'Not required'}
        </Badge>
      </span>

      {authorities.length === 0 ? (
        <Text variant="subtext" theme="neutral">
          No trusted authorities configured
        </Text>
      ) : (
        <>
          {authorities.length > 1 ? (
            <Text variant="subtext" theme="neutral">
              An image is accepted when any one authority verifies it.
            </Text>
          ) : null}
          <div className="grid gap-4 md:grid-cols-2">
            {authorities.map((authority, index) => (
              <SignatureAuthority
                key={`${authority?.type}-${authority?.issuer ?? ''}-${authority?.subject ?? authority?.subject_regexp ?? index}`}
                authority={authority}
              />
            ))}
          </div>
        </>
      )}
    </div>
  )
}
