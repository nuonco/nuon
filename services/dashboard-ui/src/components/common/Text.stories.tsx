import { Text } from "./Text";

export const Themes = () => (
  <div className="flex flex-col gap-4">
    <Text>Default text</Text>
    <Text theme="neutral">Neutral text</Text>
    <Text theme="info">Info text</Text>
    <Text theme="success">Success text</Text>
    <Text theme="warn">Warning text</Text>
    <Text theme="error">Error text</Text>
    <Text theme="brand">Brand text</Text>
  </div>
);

export const Families = () => (
  <div className="flex flex-col gap-4">
    <Text family="sans">Sans serif family</Text>
    <Text family="mono">Monospace family</Text>
    <Text family="mono" variant="h2">Mono heading with adjusted tracking</Text>
  </div>
);

export const Weights = () => (
  <div className="flex flex-col gap-4">
    <Text weight="normal">Normal weight</Text>
    <Text weight="strong">Strong weight</Text>
    <Text weight="stronger">Stronger weight</Text>
  </div>
);

export const Variants = () => (
  <div className="flex flex-col gap-4">
    <Text variant="h1">H1 variant</Text>
    <Text variant="h2">H2 variant</Text>
    <Text variant="h3">H3 variant</Text>
    <Text variant="base">Base variant</Text>
    <Text variant="body">Body variant</Text>
    <Text variant="subtext">Subtext variant</Text>
    <Text variant="label">Label variant</Text>
  </div>
);

export const SemanticRoles = () => (
  <div className="flex flex-col gap-4">
    <Text role="heading" level={1} variant="h1">Semantic heading level 1</Text>
    <Text role="heading" level={2} variant="h2">Semantic heading level 2</Text>
    <Text role="heading" level={3} variant="h3">Semantic heading level 3</Text>
    <Text role="paragraph" variant="body">This is a semantic paragraph element</Text>
    <Text role="code" family="mono" variant="body">const array = []</Text>
    <Text role="time" variant="subtext">2024-01-15T10:30:00Z</Text>
  </div>
);

export const Combinations = () => (
  <div className="flex flex-col gap-4">
    <Text variant="h1" weight="stronger" theme="brand">Strong brand heading</Text>
    <Text variant="body" theme="neutral" weight="normal">Neutral body text</Text>
    <Text variant="label" weight="strong" theme="info" family="mono">MONO INFO LABEL</Text>
    <Text variant="subtext" theme="error">Error subtext message</Text>
    <Text 
      role="heading" 
      level={2} 
      variant="h2" 
      theme="success" 
      weight="strong"
    >
      Success heading with semantic markup
    </Text>
  </div>
);

export const CustomStyling = () => (
  <div className="flex flex-col gap-4">
    <Text className="underline decoration-2">Custom underlined text</Text>
    <Text className="bg-yellow-100 px-2 py-1 rounded">Highlighted text</Text>
    <Text variant="body" className="max-w-xs truncate">
      This is very long text that will be truncated when it exceeds the maximum width
    </Text>
  </div>
);
