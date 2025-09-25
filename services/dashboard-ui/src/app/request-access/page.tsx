import { redirect } from 'next/navigation'

export default async function RequestAccessRedirect() {
  // Redirect to homepage - this page is deprecated
  redirect('/')
}
