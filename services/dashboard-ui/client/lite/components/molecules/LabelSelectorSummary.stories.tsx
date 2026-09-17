import { ComponentDocs } from '../__stories__/ComponentDocs'
import { LabelSelectorSummary } from './LabelSelectorSummary'

export default {
  title: 'lite/molecules/LabelSelectorSummary',
}

const LABEL_COLORS = {
  env: '#4cc9f0',
  tier: '#4aa578',
  region: '#8b5cf6',
}

export const Overview = () => (
  <ComponentDocs
    name="LabelSelectorSummary"
    tier="molecule"
    summary="A labels.Selector rendered as chips, with negation and wildcards visible."
    use={[
      'Show how an install group, match rule, or other selector chooses members.',
      'Pass a selector with match_labels, not_match_labels, or both.',
      'Pass labelColors to colour the value half from the app\'s label colour map.',
    ]}
    avoid={[
      'Do not paraphrase the selector as a sentence. The chips are the rule.',
      'Do not import the production matcher or LabelBadge. This molecule is the Lite surface.',
      'Do not use it for an install\'s own labels. That is a row of Badge chips.',
    ]}
    rules={[
      'match_labels render as key/value Badge chips.',
      'not_match_labels render the same chips with a visible not marker.',
      'A * value is the key alone plus an any marker, not key=*.',
      'labelColors is keyed by label key and passed to Badge; high contrast drops it.',
      'Keys are sorted so the same selector always renders in the same order.',
      'A nil or empty selector renders nothing. Loading still shows skeleton chips.',
    ]}
    props={[
      {
        name: 'selector',
        type: 'TLabelSelector | null',
        description:
          'The selector to render. Nil or empty produces no chips unless loading.',
      },
      {
        name: 'labelColors',
        type: 'Record<string, string>',
        description:
          'Per-key CSS colours from the API. Applied to the value half of key/value chips.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Skeleton chips in the same wrap layout.',
      },
      {
        name: 'loadingWidth',
        type: 'number',
        description: 'Width in ch of the first skeleton chip.',
      },
      {
        name: 'className',
        type: 'string',
        description: 'Extra classes for the wrap row.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="p-8">
    <LabelSelectorSummary
      selector={{ match_labels: { env: 'prod', tier: 'a' } }}
      labelColors={LABEL_COLORS}
    />
  </div>
)

export const Negation = () => (
  <div className="p-8">
    <LabelSelectorSummary
      selector={{
        match_labels: { env: '*' },
        not_match_labels: { env: 'stage', region: 'us-east-1' },
      }}
      labelColors={LABEL_COLORS}
    />
  </div>
)

export const Wildcard = () => (
  <div className="p-8">
    <LabelSelectorSummary
      selector={{ match_labels: { env: '*' } }}
      labelColors={LABEL_COLORS}
    />
  </div>
)

export const Empty = () => (
  <div className="p-8">
    <LabelSelectorSummary selector={{}} />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <LabelSelectorSummary loading />
  </div>
)
