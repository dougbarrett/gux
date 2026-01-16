import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'list',

  use: {
    baseURL: 'http://localhost:8093',
    trace: 'on-first-retry',
  },

  // Timeout for each test (30s for WASM load)
  timeout: 30000,

  // Output directory for test artifacts
  outputDir: 'test-results/',

  // Single project targeting chromium (sufficient for a11y testing)
  projects: [
    {
      name: 'chromium',
      use: {
        browserName: 'chromium',
      },
    },
  ],

  // Start dev server before running tests
  webServer: {
    command: 'make dev',
    url: 'http://localhost:8093',
    reuseExistingServer: !process.env.CI,
    timeout: 60000,
  },
});
