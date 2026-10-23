import { expect, test } from '@playwright/test'

// The dashboard login with its captcha on, end to end: the bundled ALTCHA
// widget fetches a challenge from this instance, solves it in a real browser,
// and the server accepts the solution. This is the path altcha 3.4.0 broke
// ("Challenge validation failed.") while every unit test stayed green.

const password = process.env.GATECHA_ADMIN_PASSWORD ?? 'changeme'

test.beforeAll(async ({ request }) => {
  const login = await request.post('/api/admin/login', { data: { username: 'admin', password } })
  expect(login.ok(), 'admin login to enable the captcha').toBeTruthy()
  const { token } = await login.json()

  const settings = await request.put('/api/admin/settings', {
    headers: { Authorization: `Bearer ${token}` },
    data: { login_captcha_enabled: true },
  })
  expect(settings.ok(), 'enable the login captcha').toBeTruthy()
})

test('signs in through the login captcha', async ({ page }) => {
  await page.goto('/login')

  const widget = page.locator('altcha-widget')
  await expect(widget).toBeVisible()
  // Click the label like a visitor: the widget draws its own box over the input.
  await widget.getByText("I'm not a robot").click()
  await expect(widget).toContainText('Verified', { timeout: 30_000 })
  await expect(widget).not.toContainText('Verification failed')

  await page.locator('#username').fill('admin')
  await page.locator('#password').fill(password)
  await page.getByRole('button', { name: 'Sign in' }).click()

  await expect(page).toHaveURL(/\/$/)
  await expect(page.getByRole('heading', { level: 1, name: 'Dashboard' })).toBeVisible()
})
