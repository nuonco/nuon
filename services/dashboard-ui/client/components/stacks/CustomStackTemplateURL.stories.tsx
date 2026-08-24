import { CustomStackTemplateURL } from './CustomStackTemplateURL'

export default {
  title: 'Stacks/CustomStackTemplateURL',
}

const uploaded = {
  name: 'k8s_namespaces',
  index: 0,
  template_url: './stacks/k8s-namespaces.yaml',
  contents_hash:
    'a685b0702e51ef05cb57e4a80d0183856f2a62d91052c12768d0c26589e617d2',
  template_source_url:
    'https://nuon-install-templates.s3.us-east-1.amazonaws.com/stacks/org123/app456/a685b0702e51ef05cb57e4a80d0183856f2a62d91052c12768d0c26589e617d2.yaml',
  status: 'ready' as const,
}

export const Uploaded = () => <CustomStackTemplateURL stack={uploaded} />

export const NotYetUploaded = () => (
  <CustomStackTemplateURL
    stack={{ ...uploaded, template_source_url: undefined, status: 'pending' }}
  />
)

export const InstallOverride = () => (
  <CustomStackTemplateURL
    stack={{
      name: 'k8s_namespaces',
      index: 0,
      template_url:
        'https://acme-templates.s3.us-west-2.amazonaws.com/k8s-namespaces.yaml',
    }}
  />
)

export const Empty = () => <CustomStackTemplateURL stack={{ name: 'empty' }} />
