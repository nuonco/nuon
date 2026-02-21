export function useUser() {
  return { user: null, error: null, isLoading: false }
}

export const UserProvider = ({ children }: { children: React.ReactNode }) => children
