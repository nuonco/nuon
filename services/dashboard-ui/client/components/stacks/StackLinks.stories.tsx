export default {
  title: 'Stacks/StackLinks',
}

import { StackLinks } from './StackLinks'

export const Both = () => (
  <div className="max-w-2xl p-4">
    <StackLinks
      template_url="https://s3.amazonaws.com/nuon-stacks/template.json"
      quick_link_url="https://console.aws.amazon.com/cloudformation/home?#/stacks/create/review?templateURL=https://s3.amazonaws.com/nuon-stacks/template.json"
    />
  </div>
)

export const TemplateOnly = () => (
  <div className="max-w-2xl p-4">
    <StackLinks
      template_url="https://s3.amazonaws.com/nuon-stacks/template.json"
      quick_link_url=""
    />
  </div>
)

export const QuickLinkOnly = () => (
  <div className="max-w-2xl p-4">
    <StackLinks
      template_url=""
      quick_link_url="https://console.aws.amazon.com/cloudformation/home?#/stacks/create/review"
    />
  </div>
)
