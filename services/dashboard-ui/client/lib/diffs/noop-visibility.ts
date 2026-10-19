export const terraformActionSelected = (
  action: string,
  selected: ReadonlySet<string>,
  showNoops: boolean
) => (action === 'no-op' ? showNoops : selected.has(action))
