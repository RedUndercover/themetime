import assert from 'node:assert/strict';
import test from 'node:test';

import { renderApp, renderLocationPreviewHTML } from './views.js';

function context() {
  return {
    appState: {
      now: '2026-07-22T12:00:00-04:00',
      today: 'Wed, Jul 22',
      plan: { active: { phaseId: 'morning', phaseName: 'Morning' }, next: { phaseName: 'Evening', clock: '20:00' } },
      config: {
        location: { label: 'New York', timezone: 'America/New_York' },
        phases: [{
          id: 'morning', name: 'Morning', color: '#f2b84b', start: { kind: 'sunrise', offsetMinutes: 20 },
          actions: [{ type: 'colorScheme', value: 'BreezeLight' }],
        }],
      },
      triggerOptions: [
        { kind: 'clock', label: 'Clock', shortLabel: 'Clock' },
        { kind: 'sunrise', label: 'Sunrise', shortLabel: 'Sunrise' },
        { kind: 'sunset', label: 'Sunset', shortLabel: 'Sunset' },
      ],
      solarEvents: [
        { kind: 'sunrise', label: 'Sunrise', shortLabel: 'Sunrise', at: '2026-07-22T05:43:00-04:00', clock: '05:43' },
        { kind: 'sunset', label: 'Sunset', shortLabel: 'Sunset', at: '2026-07-22T20:20:00-04:00', clock: '20:20' },
      ],
      transitions: [{ phaseId: 'morning', phaseName: 'Morning', color: '#f2b84b', at: '2026-07-22T06:03:00-04:00', clock: '06:03' }],
      actionOptions: [{ type: 'colorScheme', label: 'Color scheme', placeholder: 'installed value', choices: [] }],
      inventory: { smartVideoPlugin: false },
      checks: [{ name: 'KDE Plasma', detail: 'available', ok: true }],
    },
    selectedID: 'morning',
    openSheetName: '',
    settingsTab: 'location',
    locationDraft: null,
    locationPreview: null,
  };
}

test('renderApp uses backend trigger metadata', () => {
  const html = renderApp(context());
  assert.match(html, /<option value="clock"[^>]*>Clock<\/option>/);
  assert.match(html, /<option value="sunrise" selected>Sunrise<\/option>/);
  assert.match(html, /Morning/);
});

test('renderLocationPreviewHTML renders preview data', () => {
  const value = context();
  value.locationPreview = {
    today: 'Wed, Jul 22',
    solarEvents: value.appState.solarEvents,
    transitions: value.appState.transitions,
  };
  const html = renderLocationPreviewHTML(value);
  assert.match(html, /05:43/);
  assert.match(html, /20:20/);
  assert.match(html, /Morning/);
});
