import { Link } from './Link'
import { Text } from './Text'
import { ComponentDocs } from '../__stories__/ComponentDocs'

export default {
  title: 'lite/atoms/Link',
}

export const Overview = () => (
  <ComponentDocs
    name="Link"
    tier="atom"
    summary="A content link. One job, two destinations, and the destination is inferred from href rather than declared by a prop."
    use={[
      'Relative href — routes through React Router, no full page load.',
      'Absolute href (any scheme, or protocol-relative) — opens in a new tab with rel=noopener, the new-tab icon, and a screen-reader-only "opens in a new tab".',
      'Same-origin but not a router route, like a BFF download — pass reloadDocument and the browser handles the response.',
    ]}
    avoid={[
      'Navigation chrome. Main nav, tab nav and breadcrumbs are their own components on React Router NavLink, which supplies the active state instead of us passing it in.',
      'Anything that runs an action instead of changing the URL — that is a Button.',
    ]}
    rules={[
      'Links inherit surrounding type by default. A link in a sentence should match the sentence; variant is only for a standalone link with no type to inherit.',
      'Links are underlined. Colour alone is not a sufficient affordance, and in high contrast the link yellow and body white are only 1.07:1 apart.',
      'disabled renders a span with aria-disabled. An anchor cannot be disabled and a dead href is worse than no href.',
      'External links always open in a new tab.',
    ]}
    props={[
      { name: 'href', type: 'string', description: 'Destination. Its shape decides internal vs external.' },
      { name: 'variant', type: 'TTextVariant', default: 'inherit', description: 'Sizes a standalone link. Omit inside prose.' },
      { name: 'external', type: 'boolean', description: 'Forces the external treatment for an href that lies about itself.' },
      { name: 'reloadDocument', type: 'boolean', description: 'Same-origin full page load, for BFF endpoints and downloads.' },
      { name: 'disabled', type: 'boolean', default: 'false', description: 'Renders a non-interactive span.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'Delegates to Text, so a link labelled by a resource name shimmers like text.' },
    ]}
    sections={[
      {
        heading: 'Why inference is safe here',
        body: 'It is one mechanical rule on one value. The old dashboard Link had three booleans — isATag, isExternal, and an implicit third case — deciding the element through a nested ternary, with a silent winner when two were set together.',
      },
      {
        heading: 'Colour',
        body: 'Links have their own tokens rather than reusing --text-accent, which is only 4.1:1 in light and would ship a link below AA. Measured 8.00:1 light, 11.90:1 dark, 12.37:1 high contrast.',
      },
    ]}
  />
)

export const Internal = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      A relative href routes through React Router — no full page load.
    </Text>
    <div>
      <Link href="/org123/installs/inst456">acme-production</Link>
    </div>
    <div>
      <Link href="/org123/installs">View all installs</Link>
    </div>
  </div>
)

export const External = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      An absolute href is detected as external: new tab, rel=noopener, the
      new-tab icon, and a screen-reader-only “opens in a new tab”.
    </Text>
    <div>
      <Link href="https://docs.nuon.co">Read the docs</Link>
    </div>
    <div>
      <Link href="https://github.com/nuonco">nuonco on GitHub</Link>
    </div>
    <div>
      <Link href="mailto:support@example.com">Email support</Link>
    </div>
  </div>
)

export const SameOriginDownload = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      A same-origin URL that is not a router route — a BFF endpoint — uses
      reloadDocument so the browser handles the response.
    </Text>
    <div>
      <Link href="/api/orgs/org123/log-streams/ls456/logs/download" reloadDocument>
        Download logs
      </Link>
    </div>
  </div>
)

export const Inline = () => (
  <div className="max-w-lg p-8">
    <Text as="p" variant="body" color="secondary">
      Links inherit the type around them by default, so this{' '}
      <Link href="/org123/installs">install link</Link> matches the sentence it
      sits in. Pass a variant only for a standalone link that has no surrounding
      type to inherit.
    </Text>
    <div className="mt-4 flex flex-col gap-2">
      <Link href="/org123/apps" variant="caption">
        A caption-sized standalone link
      </Link>
      <Link href="/org123/apps" variant="heading">
        A heading-sized standalone link
      </Link>
    </div>
  </div>
)

export const States = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      Hover, active and focus are live — tab through to see the focus ring.
    </Text>
    <div>
      <Link href="/org123/installs">Default — hover and focus me</Link>
    </div>
    <div>
      <Link href="/org123/installs" disabled>
        Disabled — renders a span, not an anchor
      </Link>
    </div>
    <div className="flex items-center gap-2">
      <Text variant="body" color="secondary">
        Loading:
      </Text>
      <Link href="/org123/installs" loading loadingWidth={16}>
        acme-production
      </Link>
    </div>
  </div>
)

export const LoadingInContext = () => (
  <div className="flex max-w-md flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      A link whose label is the resource name shows the same shimmer Text does,
      so the row does not reflow when the name arrives.
    </Text>
    <div className="flex flex-col gap-2 rounded-xl border border-divider bg-surface-01 p-5">
      <div className="flex items-center justify-between gap-4">
        <Text variant="label" color="tertiary">
          Install
        </Text>
        <Link href="/org123/installs/inst456" loading loadingWidth={18} />
      </div>
      <div className="flex items-center justify-between gap-4">
        <Text variant="label" color="tertiary">
          Install
        </Text>
        <Link href="/org123/installs/inst456">acme-production</Link>
      </div>
    </div>
  </div>
)
