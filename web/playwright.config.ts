import { defineConfig, devices } from '@playwright/test'

// Real-browser smoke tests against a running GateCHA binary. Unit tests mock
// the ALTCHA widget, which is how a widget release that rejected every
// challenge (altcha 3.4.0) shipped with a green CI.
//
//   GATECHA_URL             instance to test (default http://127.0.0.1:8080)
//   GATECHA_ADMIN_PASSWORD  its admin password (default changeme)
export default defineConfig({
  testDir: './e2e',
  testMatch: '*.e2e.ts',
  timeout: 60_000,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: process.env.GATECHA_URL ?? 'http://127.0.0.1:8080',
    trace: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
