interface DiffEntry {
  delta?: 1 | 2 | 0;
  type?: 1 | 2 | 0;
  payload?: string;
}

// Function to handle diffing of individual lines
export function diffLines(before?: string, after?: string): string {
  if (!before && !after) return '';
  if (!before) return after!.split('\n').map(line => `+ ${line}`).join('\n');
  if (!after) return before.split('\n').map(line => `- ${line}`).join('\n');
  
  // Simple line-by-line diff
  const beforeLines = before.split('\n');
  const afterLines = after.split('\n');
  let result = '';
  
  // This is a simplified diff. For complex cases, consider using a diff library
  const maxLines = Math.max(beforeLines.length, afterLines.length);
  for (let i = 0; i < maxLines; i++) {
    if (i < beforeLines.length && i < afterLines.length) {
      if (beforeLines[i] !== afterLines[i]) {
        result += `- ${beforeLines[i]}\n+ ${afterLines[i]}\n`;
      } else {
        result += `  ${beforeLines[i]}\n`;
      }
    } else if (i < beforeLines.length) {
      result += `- ${beforeLines[i]}\n`;
    } else {
      result += `+ ${afterLines[i]}\n`;
    }
  }
  
  return result;
}

// Function to handle the entries array format
export function diffEntries(entries?: any[]): string {
  if (!entries || entries.length === 0) return '';
  
  return entries
    // Filter out entries with no payload
    .filter(entry => entry.payload !== undefined && entry.payload !== null && entry.payload !== '')
    .map(entry => {
      // Handle both delta and type field formats
      const diffType = entry.delta !== undefined ? entry.delta : entry.type;
      
      switch (diffType) {
        case 0: // Unchanged
          return `  ${entry.payload || ''}`;
        case 1: // Removed
          return `- ${entry.payload || ''}`;
        case 2: // Added
          return `+ ${entry.payload || ''}`;
        default:
          return entry.payload || '';
      }
    }).join('\n');
}

/**
 * Type for Kubernetes diff entries
 */
export interface K8SDiffEntry {
  type: number; // 1 = before, 2 = after
  path?: string;
  payload?: string;
}

/**
 * Formats Kubernetes diff entries into a unified diff string
 * Handles path-based and payload-based entries differently
 */
export function formatK8SDiff(entries: K8SDiffEntry[]): string {
  if (!entries || entries.length === 0) {
    return "No changes";
  }

  // For K8S content diff, we need to process the entries differently
  // If there are path-based entries, format them appropriately
  const pathBasedEntries = entries.filter(entry => entry.path);
  if (pathBasedEntries.length > 0) {
    return pathBasedEntries
      .map(entry => {
        // Make sure we use the correct prefix for diff highlighting
        const prefix = entry.type === 1 ? '- ' : entry.type === 2 ? '+ ' : '  ';
        return `${prefix}${entry.path}: ${entry.payload || ''}`;
      })
      .join('\n');
  }
  
  // Otherwise, these are likely YAML lines
  return entries
    .filter(entry => entry.payload)
    .map(entry => {
      // Make sure we use the correct prefix for diff highlighting
      const prefix = entry.type === 1 ? '- ' : entry.type === 2 ? '+ ' : '  ';
      return `${prefix}${entry.payload}`;
    })
    .join('\n');
}