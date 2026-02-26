/* eslint-disable react/no-unescaped-entities */
import { Code } from './Code'
import { Text } from './Text'
import { ClickToCopy } from './ClickToCopy'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Code Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Code components display code snippets with syntax highlighting,
        monospace typography, and proper formatting. They support multiple
        variants for different use cases: blocks, preformatted content, and
        inline code within text.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Code Block</h4>
      <Code>This is a default code block with basic styling.</Code>
      <Text variant="subtext" theme="neutral">
        Default variant with padding, shadows, and scrollable content
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Monospace font family for consistent character spacing</li>
        <li>Blue color scheme optimized for light and dark modes</li>
        <li>Automatic scrolling for long content</li>
        <li>Break-all word wrapping for long strings</li>
      </ul>
    </div>
  </div>
)

export const Variants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Code Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>{' '}
        prop controls the rendering mode and styling of code content. Each
        variant is optimized for different content types and contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Variant</h4>
      <Code>const greeting = "Hello, World!"; console.log(greeting);</Code>
      <Text variant="subtext" theme="neutral">
        Standard code block with basic formatting
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Preformatted Variant</h4>
      <Code variant="preformated">
        {`{
  "name": "nuon-dashboard",
  "version": "2.1.0",
  "dependencies": {
    "react": "^18.0.0",
    "next": "^14.0.0"
  }
}`}
      </Code>
      <Text variant="subtext" theme="neutral">
        Preformatted content with preserved whitespace and indentation
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Inline Variant</h4>
      <Text>
        Use the <Code variant="inline">useState</Code> hook to manage component
        state, or call <Code variant="inline">API.get('/users')</Code> to fetch
        user data. The <Code variant="inline">process.env.NODE_ENV</Code>{' '}
        variable determines the current environment.
      </Text>
      <Text variant="subtext" theme="neutral">
        Inline code snippets within text content
      </Text>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>default:</strong> Basic code block with standard styling and
        scrollable overflow
      </div>
      <div>
        <strong>preformated:</strong> Preserves exact formatting, whitespace,
        and indentation using pre element
      </div>
      <div>
        <strong>inline:</strong> Compact inline display for code within text
        paragraphs
      </div>
    </div>
  </div>
)

export const ContentExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Code Content Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Code components work with various types of technical content. Here are
        examples showing different programming languages, configuration files,
        and command-line instructions.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">JavaScript/TypeScript</h4>
      <Code variant="preformated">
        {`interface User {
  id: string;
  name: string;
  email: string;
}

const fetchUser = async (id: string): Promise<User> => {
  const response = await fetch(\`/api/users/\${id}\`);
  if (!response.ok) {
    throw new Error('Failed to fetch user');
  }
  return response.json();
};`}
      </Code>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Configuration Files</h4>
      <div className="space-y-3">
        <div>
          <Text variant="label" weight="strong">
            package.json
          </Text>
          <ClickToCopy>
            <Code variant="preformated">
              {`{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  }
}`}
            </Code>
          </ClickToCopy>
        </div>
        <div>
          <Text variant="label" weight="strong">
            Docker Configuration
          </Text>
          <Code variant="preformated">
            {`FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]`}
          </Code>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Shell Commands</h4>
      <div className="space-y-3">
        <ClickToCopy>
          <Code>curl -X GET https://api.nuon.co/v1/health</Code>
        </ClickToCopy>
        <ClickToCopy>
          <Code>docker run -p 3000:3000 nuon/dashboard:latest</Code>
        </ClickToCopy>
        <Code variant="preformated">
          {`# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build`}
        </Code>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">API Responses</h4>
      <Code variant="preformated">
        {`{
  "id": "user_123456",
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-15T10:30:00Z",
  "permissions": [
    "read:profile",
    "write:profile",
    "read:organizations"
  ]
}`}
      </Code>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Code components are frequently used in documentation, tutorials, and
        technical interfaces. Here are common patterns and recommended
        approaches for different scenarios.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Documentation Examples</h4>
      <div className="space-y-4 p-4 border rounded-lg">
        <div>
          <Text weight="strong">Installation</Text>
          <Text variant="subtext" theme="neutral" className="mt-1">
            Install the Nuon CLI using npm:
          </Text>
          <div className="mt-2">
            <ClickToCopy>
              <Code>npm install -g @nuon/cli</Code>
            </ClickToCopy>
          </div>
        </div>
        <div>
          <Text weight="strong">Authentication</Text>
          <Text variant="subtext" theme="neutral" className="mt-1">
            Set your API token as an environment variable:
          </Text>
          <div className="mt-2">
            <Code>export NUON_API_TOKEN=your_api_token_here</Code>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Error Messages and Logs</h4>
      <div className="space-y-3">
        <div className="p-3 border rounded bg-red-50 dark:bg-red-950/20">
          <Text weight="strong">Build Error</Text>
          <Code className="mt-2">
            Error: Module not found: Can't resolve './components/Button'
          </Code>
        </div>
        <div className="p-3 border rounded bg-yellow-50 dark:bg-yellow-950/20">
          <Text weight="strong">Warning</Text>
          <Code className="mt-2">
            Warning: React Hook useEffect has a missing dependency
          </Code>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Interactive Examples</h4>
      <div className="p-4 border rounded-lg">
        <Text weight="strong">Try it yourself</Text>
        <Text variant="subtext" theme="neutral" className="mt-1 mb-3">
          Click any code block to copy to clipboard:
        </Text>
        <div className="space-y-2">
          <ClickToCopy>
            <Code variant="inline">
              const result = await api.get('/status')
            </Code>
          </ClickToCopy>
          <ClickToCopy>
            <Code variant="inline">console.log('Hello from Nuon!')</Code>
          </ClickToCopy>
          <ClickToCopy>
            <Code variant="inline">npm run deploy --env=production</Code>
          </ClickToCopy>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          Use preformatted variant for structured data and multi-line code
        </li>
        <li>Use inline variant for short code snippets within text</li>
        <li>Combine with ClickToCopy for user-friendly copying experience</li>
        <li>Keep code blocks concise and focused on relevant examples</li>
        <li>
          Ensure proper contrast and readability in both light and dark modes
        </li>
      </ul>
    </div>
  </div>
)
