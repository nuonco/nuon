import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Card } from '../../atoms/Card'
import { Text } from '../../atoms/Text'
import { PageTransition } from './PageTransition'

export default {
  title: 'lite/templates/PageTransition',
}

export const Overview = () => (
  <ComponentDocs
    name="PageTransition"
    tier="template"
    summary="The routed-page boundary that carries the view transition between navigations."
    use={[
      'Wraps every leaf route element, applied for you by the route table.',
      'Rely on it to keep the sidebar, header, sub-navigation, and status bar still while the page swaps.',
    ]}
    avoid={[
      'Do not wrap a layout route, or the shell it renders will animate with the page.',
      'Do not render two boundaries at once, since the view transition name has to stay unique.',
      'Do not use it for in-page state such as tabs, filters, or surface query parameters.',
    ]}
    rules={[
      'Navigation opts in through Link, Button, NavLink, menu items, and the navigation shortcuts.',
      'Motion is a 180ms crossfade with a 4px opposing offset, and reduced motion swaps instantly.',
      'Browsers without same-document view transitions fall back to an instant swap.',
    ]}
    props={[
      {
        name: 'children',
        type: 'ReactNode',
        description: 'The routed page content.',
      },
      {
        name: 'className',
        type: 'string',
        description: 'Extra layout classes for the boundary element.',
      },
    ]}
  />
)

export const Default = () => (
  <PageTransition className="gap-6">
    <Text as="h1" variant="title">
      Install overview
    </Text>
    <Card className="min-h-40">
      <Text variant="caption" color="secondary">
        Page content animates as one boundary, so nothing inside needs its own
        motion.
      </Text>
    </Card>
  </PageTransition>
)
