'use server'

import { redirect } from 'next/navigation'
import { cookies } from 'next/headers'

export async function createOrgAction(prevState: any, formData: FormData) {
  const name = formData.get('name') as string
  
  if (!name?.trim()) {
    return { error: 'Organization name is required' }
  }

  try {
    // For now, let's just test the basic functionality
    // Simulate org creation - we'll add real API call after we confirm this works
    const mockOrg = {
      id: 'test-org-' + Date.now(),
      name: name.trim()
    }
    
    // Set the org session cookie directly
    const cookieStore = await cookies()
    cookieStore.set('org-session', mockOrg.id)
    
    // Redirect to a placeholder for now
    redirect('/?test=success')
  } catch (error) {
    console.error('Failed to create organization:', error)
    return { 
      error: error instanceof Error 
        ? error.message 
        : 'Failed to create organization. Please try again.' 
    }
  }
}