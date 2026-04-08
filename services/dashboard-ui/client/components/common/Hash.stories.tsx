export default {
  title: 'Common/Hash',
}

import { Hash } from './Hash'
import { LabeledValue } from './LabeledValue'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Hash Display</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Displays hash values (checksums, git SHAs, etc.) with truncation and
        click-to-copy functionality.
      </p>
    </div>

    <div className="space-y-4">
      <LabeledValue label="Default (12 chars)">
        <Hash hash="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" />
      </LabeledValue>

      <LabeledValue label="Short (8 chars)">
        <Hash hash="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" length={8} />
      </LabeledValue>

      <LabeledValue label="Long (16 chars)">
        <Hash hash="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" length={16} />
      </LabeledValue>

      <LabeledValue label="Empty hash">
        <Hash hash="" />
      </LabeledValue>
    </div>
  </div>
)

export const GitCommitSHAs = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Git Commit SHAs</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Common use case for displaying git commit hashes with standard 7-8
        character truncation.
      </p>
    </div>

    <div className="space-y-4">
      <LabeledValue label="Commit SHA (7 chars)">
        <Hash hash="3a7f2c1b8e4d9f5a6c2b1e8d7f4a9c3b2e1f8d7c" length={7} />
      </LabeledValue>

      <LabeledValue label="Commit SHA (8 chars)">
        <Hash hash="3a7f2c1b8e4d9f5a6c2b1e8d7f4a9c3b2e1f8d7c" length={8} />
      </LabeledValue>

      <LabeledValue label="Short SHA">
        <Hash hash="abc123d" />
      </LabeledValue>
    </div>
  </div>
)

export const Checksums = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Configuration Checksums</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Display component configuration checksums and other hash identifiers.
      </p>
    </div>

    <div className="space-y-4">
      <LabeledValue label="Config checksum">
        <Hash hash="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" />
      </LabeledValue>

      <LabeledValue label="Build artifact hash">
        <Hash hash="sha256:b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2" length={16} />
      </LabeledValue>

      <LabeledValue label="MD5 checksum">
        <Hash hash="d41d8cd98f00b204e9800998ecf8427e" length={10} />
      </LabeledValue>
    </div>
  </div>
)

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Text Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Hash component supports all Text component variants for different
        contexts.
      </p>
    </div>

    <div className="space-y-4">
      <LabeledValue label="Base variant">
        <Hash hash="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" variant="base" />
      </LabeledValue>

      <LabeledValue label="Subtext variant (default)">
        <Hash hash="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" variant="subtext" />
      </LabeledValue>

      <LabeledValue label="Label variant">
        <Hash hash="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" variant="label" />
      </LabeledValue>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Real-World Usage Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Examples showing how Hash is used throughout the application for
        various hash types.
      </p>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div className="space-y-4">
        <h4 className="font-semibold">Component Configs</h4>
        <LabeledValue label="Checksum">
          <Hash hash="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" />
        </LabeledValue>
        <LabeledValue label="Version hash">
          <Hash hash="b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7" />
        </LabeledValue>
      </div>

      <div className="space-y-4">
        <h4 className="font-semibold">Build Artifacts</h4>
        <LabeledValue label="Image digest">
          <Hash hash="sha256:c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8" length={16} />
        </LabeledValue>
        <LabeledValue label="Build hash">
          <Hash hash="d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9" length={10} />
        </LabeledValue>
      </div>

      <div className="space-y-4">
        <h4 className="font-semibold">Git References</h4>
        <LabeledValue label="HEAD">
          <Hash hash="3a7f2c1b8e4d9f5a6c2b1e8d7f4a9c3b" length={7} />
        </LabeledValue>
        <LabeledValue label="Base commit">
          <Hash hash="e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0" length={8} />
        </LabeledValue>
      </div>

      <div className="space-y-4">
        <h4 className="font-semibold">API Keys</h4>
        <LabeledValue label="Token">
          <Hash hash="f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1" length={8} />
        </LabeledValue>
        <LabeledValue label="Secret">
          <Hash hash="g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2" length={10} />
        </LabeledValue>
      </div>
    </div>
  </div>
)
