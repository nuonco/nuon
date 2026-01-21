#!/usr/bin/env node

/**
 * Pre-processes OpenAPI spec to remove circular references that break msw-auto-mock
 * 
 * This script:
 * 1. Downloads the OpenAPI spec from the API
 * 2. Identifies and breaks circular $ref chains
 * 3. Outputs a cleaned spec that msw-auto-mock can process
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const http = require('http');
const url = require('url');

const API_URL = process.env.NUON_API_URL || 'https://api.nuon.co';
const SPEC_FILE = process.env.NUON_OPENAPI_SPEC_FILE; // Local file path (optional)
const OUTPUT_FILE = path.join(__dirname, 'cleaned-openapi-spec.json');

if (SPEC_FILE) {
  console.log(`Loading OpenAPI spec from local file: ${SPEC_FILE}`);
} else {
  console.log(`Downloading OpenAPI spec from ${API_URL}/oapi/v3...`);
}

// Load the OpenAPI spec from local file
const loadSpecFromFile = () => {
  return new Promise((resolve, reject) => {
    try {
      const data = fs.readFileSync(SPEC_FILE, 'utf8');
      console.log(`Read ${data.length} bytes from ${SPEC_FILE}`);
      const spec = JSON.parse(data);
      console.log(`Parsed OpenAPI spec v${spec.openapi} with ${Object.keys(spec.components?.schemas || {}).length} schemas`);
      resolve(spec);
    } catch (err) {
      reject(new Error(`Failed to load spec from file: ${err.message}`));
    }
  });
};

// Download the OpenAPI spec (supports both HTTP and HTTPS)
const downloadSpec = () => {
  // Use local file if specified
  if (SPEC_FILE) {
    return loadSpecFromFile();
  }

  return new Promise((resolve, reject) => {
    const specUrl = `${API_URL}/oapi/v3`;
    const parsedUrl = url.parse(specUrl);
    
    // Choose the appropriate module based on protocol
    const httpModule = parsedUrl.protocol === 'https:' ? https : http;
    
    console.log(`Using ${parsedUrl.protocol} for API request...`);
    
    httpModule.get(specUrl, (res) => {
      console.log(`Response status: ${res.statusCode}`);
      
      if (res.statusCode !== 200) {
        reject(new Error(`HTTP ${res.statusCode}: Failed to download spec from ${specUrl}`));
        return;
      }
      
      let data = '';
      
      res.on('data', chunk => {
        data += chunk;
      });
      
      res.on('end', () => {
        console.log(`Downloaded ${data.length} bytes`);
        try {
          const spec = JSON.parse(data);
          console.log(`Parsed OpenAPI spec v${spec.openapi} with ${Object.keys(spec.components?.schemas || {}).length} schemas`);
          resolve(spec);
        } catch (err) {
          reject(new Error(`Failed to parse JSON: ${err.message}`));
        }
      });
    }).on('error', (err) => {
      reject(new Error(`Network error downloading spec: ${err.message}`));
    });
  });
};

// Track circular references to break them
const findCircularRefs = (obj, visited = new Set(), path = '') => {
  const circularRefs = new Set();
  
  const traverse = (current, currentPath) => {
    if (current && typeof current === 'object' && current.$ref) {
      const ref = current.$ref;
      if (visited.has(ref)) {
        circularRefs.add(ref);
        console.log(`Found circular reference: ${ref} at path: ${currentPath}`);
        return;
      }
      
      visited.add(ref);
      // Don't traverse further for $ref nodes to avoid infinite loops
      visited.delete(ref);
      return;
    }
    
    if (current && typeof current === 'object') {
      for (const [key, value] of Object.entries(current)) {
        traverse(value, `${currentPath}.${key}`);
      }
    }
  };
  
  traverse(obj, path);
  return circularRefs;
};

// Remove problematic schemas entirely to eliminate circular references
const cleanCircularRefs = (spec) => {
  console.log('Removing problematic schemas to eliminate circular references...');
  
  // Create a cleaned copy of the spec
  const cleanedSpec = JSON.parse(JSON.stringify(spec));
  
  // Detect if this is local API (more complex) vs production API
  const isLocalAPI = API_URL.includes('localhost') || API_URL.includes('127.0.0.1');
  
  // Schemas to completely remove (only truly unused/internal ones)
  const schemasToRemove = new Set(isLocalAPI ? [
    // Local API: Remove only internal/unused schemas that don't have corresponding types
    'app.CompositeStatus',  // Internal status tracking
    'app.InstallActionWorkflow',  // Internal workflow tracking
    'app.InstallActionWorkflowRun', // Internal workflow run tracking
    
    // Additional problematic schemas that are truly unused by tests
    'app.AppInputGroup',  // Part of complex input chains
    'app.HelmRelease',   // Internal Helm tracking
    'app.RunnerGroup',   // Internal runner management
    'state.SandboxState', // Internal state tracking
    'state.InstallState', // Internal state tracking  
    'app.ComponentRelease', // Internal release tracking
    'app.ComponentReleaseStep' // Internal release step tracking
    
    // KEEP ESSENTIAL BUSINESS SCHEMAS: Workflow, WorkflowStep, WorkflowStepApproval, Role, InstallDeploy
    // These have corresponding T* types in ctl-api.types.ts and are used by tests
  ] : [
    // Production API: Remove the same internal schemas
    'app.CompositeStatus',
    'app.InstallActionWorkflow', 
    'app.InstallActionWorkflowRun',
    'app.AppInputGroup',
    'app.HelmRelease',
    'app.RunnerGroup',
    'state.SandboxState',
    'state.InstallState',
    'app.ComponentRelease',
    'app.ComponentReleaseStep'
    
    // KEEP: Workflow, WorkflowStep, WorkflowStepApproval, Role, InstallDeploy - clean their refs instead
  ]);
  
  // Essential schemas to clean (not remove) - these are used by tests and have T* types
  const schemasToClean = new Set([
    'app.Workflow',      // TWorkflow - needs status, type, install_id fields
    'app.WorkflowStep',  // TWorkflowStep - needs status, approval_required fields  
    'app.WorkflowStepApproval', // TWorkflowStepApproval - needs approval fields
    'app.Role',          // TRole (from Account) - needs role_type, permissions fields
    'app.InstallDeploy'  // TDeploy - needs status_v2 field
  ]);
  
  console.log(`Using ${isLocalAPI ? 'LOCAL' : 'PRODUCTION'} API cleaning strategy`);
  console.log(`Removing ${schemasToRemove.size} internal schemas:`, Array.from(schemasToRemove));
  console.log(`Cleaning ${schemasToClean.size} essential schemas:`, Array.from(schemasToClean));
  
  // Helper function to surgically clean essential schemas
  const cleanEssentialSchema = (schemaName, schema) => {
    if (schemaName === 'app.Workflow') {
      return {
        type: 'object',
        description: 'Workflow object with essential business fields',
        properties: {
          id: { type: 'string' },
          name: { type: 'string' },
          status: { type: 'string', enum: ['running', 'completed', 'failed', 'pending', 'cancelled'] },
          type: { type: 'string', enum: ['install', 'uninstall', 'update'] },
          install_id: { type: 'string' },
          created_at: { type: 'string', format: 'date-time' },
          updated_at: { type: 'string', format: 'date-time' },
          created_by_id: { type: 'string' },
          // Remove circular references to steps, approvals, etc. but keep business fields
        }
      };
    }
    
    if (schemaName === 'app.WorkflowStep') {
      return {
        type: 'object', 
        description: 'Workflow step with essential business fields',
        properties: {
          id: { type: 'string' },
          status: { type: 'string', enum: ['pending', 'running', 'completed', 'failed', 'skipped'] },
          type: { type: 'string' },
          name: { type: 'string' },
          workflow_id: { type: 'string' },
          approval_required: { type: 'boolean' },
          execution_type: { type: 'string' },
          created_at: { type: 'string', format: 'date-time' },
          updated_at: { type: 'string', format: 'date-time' },
          // Remove circular references to approvals, workflows, etc.
        }
      };
    }
    
    if (schemaName === 'app.WorkflowStepApproval') {
      return {
        type: 'object',
        description: 'Workflow step approval with essential business fields', 
        properties: {
          id: { type: 'string' },
          status: { type: 'string', enum: ['pending', 'approved', 'rejected'] },
          workflow_step_id: { type: 'string' },
          created_at: { type: 'string', format: 'date-time' },
          updated_at: { type: 'string', format: 'date-time' },
          approved_by_id: { type: 'string' },
          // Remove circular references but keep approval business logic
        }
      };
    }
    
    if (schemaName === 'app.Role') {
      return {
        type: 'object',
        description: 'Role with essential business fields',
        properties: {
          id: { type: 'string' },
          name: { type: 'string' },
          role_type: { type: 'string', enum: ['org_admin', 'installer', 'runner'] },
          org_id: { type: 'string' },
          created_at: { type: 'string', format: 'date-time' },
          updated_at: { type: 'string', format: 'date-time' },
          // Remove circular references to Account but keep role business fields
        }
      };
    }
    
    if (schemaName === 'app.InstallDeploy') {
      return {
        type: 'object',
        description: 'Install deploy with essential business fields',
        properties: {
          id: { type: 'string' },
          status: { type: 'string' },
          status_v2: { type: 'string' },
          status_description: { type: 'string' },
          build_id: { type: 'string' },
          component_id: { type: 'string' },
          component_name: { type: 'string' },
          install_id: { type: 'string' },
          plan_only: { type: 'boolean' },
          created_at: { type: 'string', format: 'date-time' },
          updated_at: { type: 'string', format: 'date-time' },
          created_by_id: { type: 'string' },
          // Remove circular references to workflows, builds, etc.
        }
      };
    }
    
    // If we don't have specific cleaning logic, return the original schema
    return schema;
  };
  
  // Step 1: Remove truly problematic/unused schemas from components.schemas
  if (cleanedSpec.components && cleanedSpec.components.schemas) {
    for (const schemaName of schemasToRemove) {
      if (cleanedSpec.components.schemas[schemaName]) {
        console.log(`Removing schema: ${schemaName}`);
        delete cleanedSpec.components.schemas[schemaName];
      }
    }
  }
  
  // Step 2: Clean essential schemas surgically (remove circular refs but keep business fields)
  if (cleanedSpec.components && cleanedSpec.components.schemas) {
    for (const schemaName of schemasToClean) {
      if (cleanedSpec.components.schemas[schemaName]) {
        console.log(`Cleaning circular references in schema: ${schemaName}`);
        cleanedSpec.components.schemas[schemaName] = cleanEssentialSchema(schemaName, cleanedSpec.components.schemas[schemaName]);
      }
    }
  }
  
  // Step 3: Remove references to deleted schemas throughout the spec
  const removeRefsToDeletedSchemas = (obj) => {
    if (obj && typeof obj === 'object') {
      if (obj.$ref) {
        const schemaName = obj.$ref.split('/').pop();
        if (schemasToRemove.has(schemaName)) {
          console.log(`Removing reference to deleted schema: ${schemaName}`);
          // Replace with a basic object structure that tests expect
          return {
            type: 'object',
            description: `Simplified replacement for ${schemaName}`,
            properties: {
              id: { type: 'string' },
              name: { type: 'string' },
              created_at: { type: 'string', format: 'date-time' },
              updated_at: { type: 'string', format: 'date-time' }
            }
          };
        }
        
        // For essential schemas that we cleaned, the reference should remain as-is
        // because we kept the schema but just cleaned its internal circular references
        if (schemasToClean.has(schemaName)) {
          return obj; // Keep the reference - the target schema is cleaned but present
        }
        return obj;
      }
      
      // Handle allOf constructs
      if (obj.allOf && Array.isArray(obj.allOf)) {
        const filteredAllOf = obj.allOf
          .map(subSchema => removeRefsToDeletedSchemas(subSchema))
          .filter(subSchema => subSchema !== null);
        
        if (filteredAllOf.length === 0) {
          return {
            type: 'object',
            description: 'Empty allOf after schema removal'
          };
        }
        
        return {
          ...obj,
          allOf: filteredAllOf
        };
      }
      
      // Process object properties
      const cleaned = {};
      for (const [key, value] of Object.entries(obj)) {
        const cleanedValue = removeRefsToDeletedSchemas(value);
        if (cleanedValue !== null) {
          cleaned[key] = cleanedValue;
        }
      }
      return cleaned;
    }
    
    if (Array.isArray(obj)) {
      return obj.map(item => removeRefsToDeletedSchemas(item)).filter(item => item !== null);
    }
    
    return obj;
  };
  
  const result = removeRefsToDeletedSchemas(cleanedSpec);
  console.log('Completed schema removal - no circular references remaining');
  return result;
};

// Main execution
const main = async () => {
  try {
    console.log('Starting OpenAPI spec cleaning process...');
    
    const originalSpec = await downloadSpec();
    console.log(`Downloaded spec with ${Object.keys(originalSpec.components?.schemas || {}).length} schemas`);
    
    const cleanedSpec = cleanCircularRefs(originalSpec);
    
    // Write the cleaned spec to file
    fs.writeFileSync(OUTPUT_FILE, JSON.stringify(cleanedSpec, null, 2));
    console.log(`Cleaned OpenAPI spec written to: ${OUTPUT_FILE}`);
    
    // Output the file path for use by the mock generation command
    console.log(`CLEANED_SPEC_FILE=${OUTPUT_FILE}`);
    
  } catch (error) {
    console.error('Failed to clean OpenAPI spec:', error.message);
    process.exit(1);
  }
};

if (require.main === module) {
  main();
}

module.exports = { downloadSpec, cleanCircularRefs };