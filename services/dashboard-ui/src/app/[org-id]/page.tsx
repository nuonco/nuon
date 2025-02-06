import { redirect } from 'next/navigation'

// TODO(nnnat): future org level dashboard
export default async function OrgDashboard({ params }) {
  const orgId = params?.['org-id'] as string

  redirect(`/${orgId}/apps`)  
}
