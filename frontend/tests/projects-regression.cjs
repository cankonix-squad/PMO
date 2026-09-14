// Browser regression with fixture-only API responses. No real account or database writes.
const assert = require('node:assert/strict');
const { chromium } = require('playwright');
const baseURL = process.env.TEST_BASE_URL || 'http://127.0.0.1:3317';
const id = '11111111-1111-4111-8111-111111111111';
const project = {
  id, organization_id: id, code: 'REF-001', name: 'Proyek Uji Refactor',
  status: 'ACTIVE', priority: 'MEDIUM', category: 'BND', progress_pct: 25,
  budget_total: 1000000, currency: 'IDR', description: '<p>Deskripsi proyek</p>',
  objectives: '<p>Tujuan proyek</p>', start_date: '2026-01-01', end_date: '2026-12-31',
  org_unit_id: id, org_unit_name: 'Balai Uji', program_id: id, sector_id: id,
  region_id: id, river_basin_id: id, manager_id: null,
  created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z',
};
const master = { id, code: 'BND', name: 'Master Uji', is_active: true };
const state = JSON.stringify({ state: {
  user: { id, organization_id: id, first_name: 'Uji', last_name: 'Refactor', roles: ['SUPER_ADMIN'], permissions: [] },
  accessToken: 'fixture-access-token', refreshToken: 'fixture-refresh-token',
  isAuthenticated: true, rememberMe: true,
}, version: 0 });

(async () => {
  const browser = await chromium.launch({ headless: true, executablePath: process.env.TEST_CHROME_PATH || undefined });
  try {
    const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
    await context.addCookies([{ name: 'cankora-auth', value: encodeURIComponent(state), url: baseURL }]);
    await context.addInitScript(value => localStorage.setItem('cankora-auth', value), state);
    const tasks = [], budgets = [], writes = [], errors = [];
    await context.route('**/api/v1/**', async route => {
      const request = route.request(), url = new URL(request.url()), endpoint = url.pathname.replace('/api/v1', '');
      let data = [];
      if (request.method() === 'POST') {
        const payload = request.postDataJSON(); writes.push({ endpoint, payload });
        if (endpoint === `/projects/${id}/tasks`) {
          data = { ...payload, id: '22222222-2222-4222-8222-222222222222', project_id: id, organization_id: id, status: 'TODO', progress_pct: 0, created_at: project.created_at };
          tasks.push(data);
        } else if (endpoint === `/projects/${id}/budgets`) {
          data = { ...payload, id: '33333333-3333-4333-8333-333333333333', project_id: id, variance: payload.planned - payload.actual, usage_pct: payload.actual / payload.planned * 100, status: 'NORMAL' };
          budgets.push(data);
        } else throw new Error(`Unexpected write to ${endpoint}`);
      } else if (request.method() !== 'GET') {
        throw new Error(`Unexpected ${request.method()} ${endpoint}`);
      } else if (endpoint === '/projects') data = [project];
      else if (endpoint === `/projects/${id}`) data = project;
      else if (endpoint === `/projects/${id}/tasks`) data = tasks;
      else if (endpoint === `/projects/${id}/budgets`) data = budgets;
      else if (endpoint === `/projects/${id}/control`) data = {
        as_of: project.created_at, project,
        contract: { count: 0, total_value: 0, active_count: 0, currency: 'IDR' },
        evidence: { inspections: 0, verified_inspections: 0, evidence_files: 0 },
        issues: [], risks: [], actions: [],
      };
      else if (['/org-units', '/programs', '/sectors', '/regions', '/river-basins', '/project-categories'].includes(endpoint)) data = [master];
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, message: 'Fixture', data, meta: { total: Array.isArray(data) ? data.length : 1, page: 1, page_size: 100 } }) });
    });
    const page = await context.newPage();
    page.on('pageerror', error => errors.push(error.message));
    await page.goto(`${baseURL}/projects`);
    await page.getByRole('heading', { name: project.name, exact: true }).waitFor();
    await page.getByRole('button', { name: 'Proyek Baru', exact: true }).click();
    await page.locator('input[name="code"]').waitFor();
    await page.getByRole('button', { name: 'Batal', exact: true }).click();
    assert.equal(await page.locator('input[name="code"]').count(), 0, 'project form closes');

    await page.goto(`${baseURL}/projects/${id}`);
    await page.getByRole('heading', { name: project.name, exact: true }).waitFor();
    await page.getByRole('button', { name: 'Task', exact: true }).click();
    const taskForm = page.locator('form').filter({ has: page.getByRole('heading', { name: 'Tambah Task', exact: true }) });
    await taskForm.getByRole('button', { name: 'Simpan', exact: true }).click();
    await page.getByText('Task title is required', { exact: true }).waitFor();
    assert.equal(writes.length, 0, 'invalid task does not send an API write');
    await taskForm.locator('input[name="title"]').fill('  Task hasil refactor  ');
    await taskForm.getByRole('button', { name: 'Simpan', exact: true }).click();
    await page.getByRole('heading', { name: 'Task hasil refactor', exact: true }).waitFor();
    assert.equal(writes[0].endpoint, `/projects/${id}/tasks`);
    assert.equal(writes[0].payload.title, 'Task hasil refactor');
    assert.equal(writes[0].payload.est_hours, 0);
    assert.equal(await taskForm.count(), 0, 'task form closes after successful mutation');

    await page.getByRole('button', { name: 'Budget', exact: true }).click();
    const budgetForm = page.locator('form').filter({ has: page.getByRole('heading', { name: 'Tambah Baris Anggaran', exact: true }) });
    await budgetForm.locator('input[name="category"]').fill('Pekerjaan Uji');
    await budgetForm.locator('input[name="planned"]').fill('1000');
    await budgetForm.locator('input[name="actual"]').fill('250');
    await budgetForm.getByRole('button', { name: 'Simpan', exact: true }).click();
    await page.getByText('Pekerjaan Uji', { exact: true }).waitFor();
    assert.equal(writes[1].payload.planned, 1000);
    assert.equal(writes[1].payload.actual, 250);
    assert.equal(await budgetForm.count(), 0, 'budget form closes after mutation');
    await page.setViewportSize({ width: 390, height: 844 });
    await page.getByRole('button', { name: 'Task', exact: true }).click();
    await taskForm.waitFor();
    await taskForm.getByRole('button', { name: 'Batal', exact: true }).click();
    assert.deepEqual(errors, [], 'no browser runtime errors');
    console.log('PASS: list/detail rendering, project form, validation, task/budget payloads, query refresh, mobile form, no runtime errors');
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
