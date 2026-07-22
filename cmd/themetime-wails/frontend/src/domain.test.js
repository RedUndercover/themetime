import assert from 'node:assert/strict';
import test from 'node:test';

import {
  addAction,
  escapeHTML,
  humanTrigger,
  nextID,
  normaliseConfig,
  parseLocationDraft,
  removeAction,
  removeRule,
  solarKinds,
  triggerLabel,
  triggerText,
} from './domain.js';

const triggerOptions = [
  { kind: 'clock', label: 'Clock', shortLabel: 'Clock' },
  { kind: 'sunrise', label: 'Sunrise', shortLabel: 'Sunrise' },
  { kind: 'sunset', label: 'Sunset', shortLabel: 'Sunset' },
];

test('trigger helpers preserve backend order and labels', () => {
  const state = { triggerOptions };
  assert.deepEqual(solarKinds(state), ['sunrise', 'sunset']);
  assert.equal(triggerLabel(state, 'sunrise'), 'Sunrise');
  assert.equal(triggerLabel(state, 'unknown'), 'unknown');
});

test('clock and solar triggers have stable human-readable formatting', () => {
  const state = { triggerOptions };
  assert.equal(triggerText(state, { kind: 'clock', clock: '07:05' }), '07:05');
  assert.equal(humanTrigger(state, { kind: 'clock', clock: '07:05' }), 'At 07:05');
  assert.equal(triggerText(state, { kind: 'sunrise', offsetMinutes: -15 }), 'Sunrise -15m');
  assert.equal(humanTrigger(state, { kind: 'sunrise', offsetMinutes: -15 }), '15 minutes before sunrise');
  assert.equal(humanTrigger(state, { kind: 'sunset', offsetMinutes: 20 }), '20 minutes after sunset');
});

test('normaliseConfig canonicalizes clock and solar phases', () => {
  const config = {
    phases: [
      { name: '  ', color: '', start: { kind: 'clock', clock: '', offsetMinutes: 20 } },
      { name: ' Dawn ', color: ' #123456 ', start: { kind: 'sunrise', clock: '10:00', offsetMinutes: '15' } },
    ],
  };
  normaliseConfig(config);
  assert.deepEqual(config.phases[0], { name: 'Untitled rule', color: '#4e79a7', start: { kind: 'clock', clock: '12:00' } });
  assert.deepEqual(config.phases[1], { name: 'Dawn', color: '#123456', start: { kind: 'sunrise', clock: '', offsetMinutes: 15 } });
});

test('editor mutations add and remove actions and select a neighboring rule', () => {
  const phase = { actions: [] };
  addAction(phase, 'colorScheme', 'BreezeDark');
  assert.deepEqual(phase.actions, [{ type: 'colorScheme', value: 'BreezeDark' }]);
  removeAction(phase, 0);
  assert.deepEqual(phase.actions, []);

  const config = { phases: [{ id: 'one' }, { id: 'two' }, { id: 'three' }] };
  assert.equal(removeRule(config, 'two'), 'one');
  assert.deepEqual(config.phases.map((item) => item.id), ['one', 'three']);
  assert.equal(nextID({ phases: [{ id: 'rule-1' }, { id: 'rule-3' }] }, 'rule'), 'rule-2');
});

test('location parsing validates bounds and supplies a fallback label', () => {
  assert.deepEqual(parseLocationDraft({ label: '', latitude: '40', longitude: '-74', timezone: 'America/New_York' }), {
    location: { label: 'America/New_York', latitude: 40, longitude: -74, timezone: 'America/New_York', source: 'manual' },
  });
  assert.match(parseLocationDraft({ label: '', latitude: '91', longitude: '0', timezone: 'UTC' }).error, /Latitude/);
});

test('escapeHTML escapes text and attribute metacharacters', () => {
  assert.equal(escapeHTML(`<script a="b">'&`), '&lt;script a=&quot;b&quot;&gt;&#039;&amp;');
});
