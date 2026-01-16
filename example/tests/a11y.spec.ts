import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import type { AxeResults, Result } from 'axe-core';

// WCAG 2.1 AA compliance tags
const WCAG_21_AA_TAGS = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

/**
 * Format axe violations into a readable string for test failure messages.
 * Makes test failures actionable by providing rule id, impact, description, and affected elements.
 */
function formatViolations(violations: Result[]): string {
  if (violations.length === 0) {
    return 'No violations found';
  }

  return violations.map((violation, index) => {
    const nodes = violation.nodes.map(node => {
      const target = node.target.join(', ');
      const html = node.html.substring(0, 100) + (node.html.length > 100 ? '...' : '');
      return `    - Target: ${target}\n      HTML: ${html}`;
    }).join('\n');

    return `
${index + 1}. ${violation.id} (${violation.impact})
   Description: ${violation.description}
   Help: ${violation.help}
   Help URL: ${violation.helpUrl}
   Affected elements:
${nodes}`;
  }).join('\n');
}

/**
 * Wait for WASM application to fully load.
 * The #app element initially contains "Loading..." and is replaced when WASM mounts.
 */
async function waitForWasmLoad(page: any) {
  await page.goto('/');
  await expect(page.locator('#app')).not.toContainText('Loading...', { timeout: 15000 });
}

test.describe('Accessibility Tests', () => {
  test.describe.configure({ retries: 1 }); // Handle WASM loading flakiness

  test('home page should have no WCAG 2.1 AA violations', async ({ page }) => {
    await waitForWasmLoad(page);

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(WCAG_21_AA_TAGS)
      .analyze();

    const violations = accessibilityScanResults.violations;

    expect(
      violations,
      `Found ${violations.length} accessibility violation(s):\n${formatViolations(violations)}`
    ).toHaveLength(0);
  });

  test('interactive components have no critical violations', async ({ page }) => {
    await waitForWasmLoad(page);

    // Focus on high-impact rules critical for interactive components
    const criticalRules = [
      'button-name',
      'label',
      'aria-required-attr',
      'aria-valid-attr',
      'aria-valid-attr-value',
      'color-contrast',
      'focus-order-semantics',
      'interactive-supports-focus',
    ];

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(WCAG_21_AA_TAGS)
      .analyze();

    // Filter for critical violations only
    const criticalViolations = accessibilityScanResults.violations.filter(
      v => criticalRules.includes(v.id) || v.impact === 'critical' || v.impact === 'serious'
    );

    expect(
      criticalViolations,
      `Found ${criticalViolations.length} critical accessibility violation(s):\n${formatViolations(criticalViolations)}`
    ).toHaveLength(0);
  });
});
