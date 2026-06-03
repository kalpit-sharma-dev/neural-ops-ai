import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

const PAGES = ['/', '/logs', '/incidents', '/design-system', '/nfr-certification'];

test.describe('Accessibility (axe)', () => {
  for (const path of PAGES) {
    test(`no critical violations on ${path}`, async ({ page }) => {
      await page.goto(path);
      await page.waitForLoadState('networkidle');
      const results = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa']).analyze();
      const critical = results.violations.filter((v) => v.impact === 'critical' || v.impact === 'serious');
      expect(critical, JSON.stringify(critical, null, 2)).toEqual([]);
    });
  }
});
