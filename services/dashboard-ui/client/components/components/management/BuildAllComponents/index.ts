export { BuildAllComponentsButtonContainer as BuildAllComponentsButton } from './BuildAllComponentsContainer'
export {
  BuildAllComponentsButton as BuildAllComponentsButtonComponent,
  BuildAllComponentsModal as BuildAllComponentsModalComponent,
} from './BuildAllComponents'

// Re-export container modal for consumers that need the hooked version (e.g., spotlight)
export { BuildAllComponentsModalContainer as BuildAllComponentsModal } from './BuildAllComponentsContainer'
