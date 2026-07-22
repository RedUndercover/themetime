import {
  actionIconName,
  actionOption as findActionOption,
  clockOf,
  colorPresets,
  displayTransitionClock,
  escapeAttr,
  escapeHTML,
  humanTrigger,
  locationDraftFrom,
  minuteOfDay,
  offsetClock,
  offsetLabel,
  option,
  percent,
  range,
  shortValue,
  startPresets,
  triggerIcon,
  triggerText,
  two,
} from './domain.js';
import { icon } from './icons.js';

let appState;
let selectedID;
let openSheetName;
let settingsTab;
let locationDraft;
let locationPreview;

function useContext(context) {
  ({ appState, selectedID, openSheetName, settingsTab, locationDraft, locationPreview } = context);
}

export function renderApp(context) {
  useContext(context);
  const active = appState.plan.active;
  const next = appState.plan.next;
  const selected = selectedPhase();
  return `
    <div class="app-shell">
      <header class="app-header">
        <div class="brand">
          <span class="brand-mark" aria-hidden="true"></span>
          <div>
            <div class="brand-title">ThemeTime</div>
            <div class="brand-subtitle">Solar theme scheduler</div>
          </div>
        </div>
        <button class="location-button" data-open-settings="location" title="Edit location">
          ${icon('location')}
          <span><b>${escapeHTML(appState.today)}</b><em>${escapeHTML(locationLabel())}</em></span>
          ${icon('chevron')}
        </button>
        <div class="header-actions">
          <span class="status-pill"><i class="status-dot active-dot"></i><span>Active</span><b>${escapeHTML(active?.phaseName || 'None')}</b></span>
          <span class="status-pill next"><span>Next</span><b>${escapeHTML(next ? `${next.phaseName} ${next.clock}` : 'None')}</b></span>
          <button class="icon-button quiet" data-open-settings="system" aria-label="Open settings" title="Location and system">${icon('settings')}</button>
        </div>
      </header>

      <section class="timeline-card" aria-label="Solar timeline">
        <div class="timeline-heading">
          <div><span class="eyebrow">Today</span><h1>Solar timeline</h1></div>
          <div class="timeline-legend" aria-label="Timeline legend">
            <span><i class="legend-line now"></i>Now</span>
            <span><i class="legend-line selected"></i>Selected rule</span>
          </div>
        </div>
        ${renderSolarRibbon()}
      </section>

      <main class="workspace">
        <aside class="rule-rail panel">
          <div class="rail-head">
            <div><span class="eyebrow">Schedule</span><h2>Rules</h2></div>
            <button class="icon-button" data-action="new-rule" aria-label="New rule" title="New rule">${icon('plus')}</button>
          </div>
          <div class="rule-list">${renderRuleList()}</div>
          <div class="rail-actions">
            <button data-action="remove-rule" class="secondary">${icon('trash')}<span>Remove</span></button>
            <button data-action="apply-rule" class="secondary">${icon('play')}<span>Apply now</span></button>
          </div>
        </aside>

        <section class="editor panel">
          ${selected ? renderEditor(selected) : renderEmptyEditor()}
        </section>

        <aside class="inspector panel" aria-label="Today and system status">
          ${renderInspector()}
        </aside>
      </main>
    </div>
    <button class="compact-inspector-button" data-open-sheet="inspector">${icon('sunrise')}<span>Today</span></button>
    ${openSheetName === 'settings' ? renderSettingsSheet() : ''}
    ${openSheetName === 'inspector' ? renderInspectorSheet() : ''}
  `;
}

function renderSolarRibbon() {
  const events = appState.solarEvents.filter((event) => event.at && !event.error);
  const eventMap = Object.fromEntries(events.map((event) => [event.kind, event]));
  const sunrise = eventMap.sunrise ? minuteOfDay(eventMap.sunrise.at) : 360;
  const noon = eventMap.solarNoon ? minuteOfDay(eventMap.solarNoon.at) : 720;
  const sunset = eventMap.sunset ? minuteOfDay(eventMap.sunset.at) : 1080;
  const sunriseX = percent(sunrise) * 10;
  const noonX = percent(noon) * 10;
  const sunsetX = percent(sunset) * 10;
  const sunArc = solarArcPath(sunriseX, noonX, sunsetX);
  const currentMinute = minuteOfDay(appState.now || new Date());
  const selectedTransition = appState.transitions.find((item) => item.phaseId === selectedID);
  const segments = buildPhaseSegments();

  return `
    <div class="solar-ribbon" style="--sunrise:${percent(sunrise)}%;--noon:${percent(noon)}%;--sunset:${percent(sunset)}%">
      <div class="sky-scene">
        <div class="stars" aria-hidden="true"></div>
        <svg class="landscape" viewBox="0 0 1000 150" preserveAspectRatio="none" aria-hidden="true">
          <path class="ridge ridge-back" d="M0 118 L70 96 L145 111 L220 83 L300 109 L385 88 L470 111 L555 89 L640 107 L730 78 L820 108 L915 90 L1000 115 L1000 150 L0 150 Z" />
          <path class="ridge ridge-front" d="M0 128 L95 106 L185 130 L275 104 L370 128 L465 112 L560 132 L660 101 L760 126 L855 105 L1000 128 L1000 150 L0 150 Z" />
          <path class="sun-arc" d="${sunArc}" vector-effect="non-scaling-stroke" />
        </svg>
        ${events.map(renderSkyEvent).join('')}
        <div class="now-marker" style="left:${percent(currentMinute)}%"><span>${clockOf(appState.now)}</span></div>
      </div>
      <div class="hour-scale" aria-hidden="true">
        ${range(13).map((item) => `<span style="left:${percent(item * 120)}%">${two(item * 2)}:00</span>`).join('')}
      </div>
      <div class="phase-track">
        ${segments.map((segment) => `
          <button class="phase-segment ${segment.phaseId === selectedID ? 'selected' : ''}"
            style="left:${percent(segment.start)}%;width:${Math.max(0.35, percent(segment.end - segment.start))}%;--phase-color:${escapeAttr(segment.color || '#4e79a7')}"
            data-select="${escapeAttr(segment.phaseId)}" title="${escapeAttr(`${segment.phaseName} · ${offsetClock(segment.start)}–${offsetClock(segment.end)}`)}">
            <span>${escapeHTML(segment.phaseName)}</span>
          </button>
        `).join('')}
        ${selectedTransition ? `<div class="selected-rule-marker" style="left:${percent(minuteOfDay(selectedTransition.at))}%" aria-hidden="true"></div>` : ''}
      </div>
    </div>
  `;
}

function solarArcPath(sunriseX, noonX, sunsetX) {
  const horizonY = 116;
  const apexY = 18;
  const morningSpan = Math.max(1, noonX - sunriseX);
  const eveningSpan = Math.max(1, sunsetX - noonX);
  const morningHandle = morningSpan * 0.42;
  const eveningHandle = eveningSpan * 0.42;
  return [
    `M ${sunriseX} ${horizonY}`,
    `C ${sunriseX + morningHandle} ${horizonY}`,
    `${noonX - morningHandle} ${apexY}`,
    `${noonX} ${apexY}`,
    `C ${noonX + eveningHandle} ${apexY}`,
    `${sunsetX - eveningHandle} ${horizonY}`,
    `${sunsetX} ${horizonY}`,
  ].join(' ');
}

function renderSkyEvent(event) {
  const minute = minuteOfDay(event.at);
  const major = ['sunrise', 'solarNoon', 'sunset'].includes(event.kind);
  const label = event.kind === 'solarNoon' ? 'Noon' : event.shortLabel;
  return `
    <button class="sky-event ${major ? 'major' : 'minor'} ${escapeAttr(event.kind)}" style="left:${percent(minute)}%"
      data-trigger="${escapeAttr(event.kind)}" title="Set selected rule to ${escapeAttr(event.label)} at ${escapeAttr(event.clock)}">
      <span class="sky-event-icon">${icon(triggerIcon(event.kind))}</span>
      <b>${escapeHTML(event.clock)}</b>
      <em>${escapeHTML(label)}</em>
    </button>
  `;
}

function buildPhaseSegments() {
  const transitions = (appState.transitions || [])
    .filter((item) => item.at)
    .map((item) => ({ ...item, minute: minuteOfDay(item.at) }))
    .sort((a, b) => a.minute - b.minute);
  if (!transitions.length) return [];
  const segments = [];
  if (transitions[0].minute > 0) {
    const previous = transitions[transitions.length - 1];
    segments.push({ ...previous, start: 0, end: transitions[0].minute });
  }
  transitions.forEach((item, index) => {
    const end = transitions[index + 1]?.minute ?? 1440;
    if (end > item.minute) segments.push({ ...item, start: item.minute, end });
  });
  return segments;
}

function renderRuleList() {
  const phases = appState.config.phases || [];
  if (!phases.length) return '<div class="empty-list">No rules yet</div>';
  return phases.map((phase) => {
    const selected = phase.id === selectedID;
    const active = phase.id === appState.plan.active?.phaseId;
    const transition = appState.transitions.find((item) => item.phaseId === phase.id);
    const actionCount = phase.actions?.length || 0;
    return `
      <button class="rule-row ${selected ? 'selected' : ''}" data-select="${escapeAttr(phase.id)}" aria-current="${selected ? 'true' : 'false'}">
        <span class="rule-icon" style="--phase-color:${escapeAttr(phase.color || '#4e79a7')}">${icon(triggerIcon(phase.start.kind))}</span>
        <span class="rule-main">
          <b>${escapeHTML(phase.name || 'Untitled rule')}${active ? '<i>Active</i>' : ''}</b>
          <em>${escapeHTML(transition ? displayTransitionClock(transition) : triggerText(appState, phase.start))} · ${actionCount} ${actionCount === 1 ? 'action' : 'actions'}</em>
        </span>
        ${icon('chevron')}
      </button>
    `;
  }).join('');
}

function renderEditor(phase) {
  return `
    <div class="editor-head">
      <span class="editor-symbol" style="--phase-color:${escapeAttr(phase.color || '#4e79a7')}">${icon(triggerIcon(phase.start.kind))}</span>
      <div class="editor-title">
        <label for="rule-name">Rule name</label>
        <input id="rule-name" class="rule-name" value="${escapeAttr(phase.name || '')}" />
      </div>
      <button class="primary" data-action="save">${icon('save')}<span>Save changes</span></button>
    </div>

    <section class="edit-section timing-section">
      <div class="section-heading"><div><span class="eyebrow">When</span><h3>Start time</h3></div><p>${escapeHTML(humanTrigger(appState, phase.start))}</p></div>
      <div class="timing-controls">
        <label class="field"><span>Trigger</span><select class="trigger-select">
          ${appState.triggerOptions.map((item) => option(item.kind, phase.start.kind, item.label)).join('')}
        </select></label>
        ${phase.start.kind === 'clock' ? renderClockControls(phase) : renderOffsetStepper(phase)}
      </div>
      <details class="advanced timing-advanced">
        <summary>${icon('tune')}<span>Presets and precise timing</span>${icon('chevron')}</summary>
        <div class="advanced-content">
          <div class="preset-grid">
            ${startPresets.map(([label, kind, offset, color, symbol]) => `
              <button class="preset ${phase.start.kind === kind && (phase.start.offsetMinutes || 0) === offset ? 'selected' : ''}"
                data-preset-kind="${kind}" data-preset-offset="${offset}" data-preset-color="${color}">
                ${icon(symbol)}<span>${label}</span>
              </button>
            `).join('')}
          </div>
          ${phase.start.kind !== 'clock' ? renderOffsetControls(phase) : ''}
        </div>
      </details>
    </section>

    <section class="edit-section">
      <div class="section-heading"><div><span class="eyebrow">Then</span><h3>Actions</h3></div><p>${phase.actions?.length || 0} configured</p></div>
      <p class="action-layer-note">Actions layer onto earlier rules. Settings omitted here keep their latest value; a new wallpaper replaces only the matching wallpaper target.</p>
      <div class="action-list">${renderActionList(phase)}</div>
      <details class="advanced add-action">
        <summary>${icon('plus')}<span>Add an action</span>${icon('chevron')}</summary>
        <div class="advanced-content">${renderActionBuilder()}</div>
      </details>
    </section>

    <details class="advanced appearance-section">
      <summary>${icon('palette')}<span>Rule appearance</span>${icon('chevron')}</summary>
      <div class="advanced-content">
        <div class="swatches">
          ${colorPresets.map(([label, value]) => `<button class="swatch ${phase.color?.toLowerCase() === value.toLowerCase() ? 'selected' : ''}" data-color="${value}" style="--swatch:${value}"><span></span>${label}</button>`).join('')}
        </div>
        <label class="field color-field"><span>Custom hex color</span><input class="color-input" value="${escapeAttr(phase.color || '#4e79a7')}" /></label>
      </div>
    </details>
  `;
}

function renderEmptyEditor() {
  return `<div class="empty-editor"><span>${icon('sunrise')}</span><h2>Create your first rule</h2><p>Choose a solar event or clock time, then add the KDE settings ThemeTime should apply.</p><button class="primary" data-action="new-rule">${icon('plus')}<span>New rule</span></button></div>`;
}

function renderClockControls(phase) {
  const [hour, minute] = (phase.start.clock || '12:00').split(':');
  return `
    <div class="clock-controls">
      <label class="field"><span>Hour</span><select class="clock-hour">${range(24).map((item) => option(two(item), hour)).join('')}</select></label>
      <label class="field"><span>Minute</span><select class="clock-minute">${range(60).map((item) => option(two(item), minute)).join('')}</select></label>
    </div>
  `;
}

function renderOffsetStepper(phase) {
  const offset = phase.start.offsetMinutes || 0;
  return `
    <label class="field offset-field"><span>Offset</span><span class="stepper">
      <button data-offset-step="-5" aria-label="Decrease offset by five minutes">${icon('minus')}</button>
      <input class="offset-number" type="number" min="-180" max="180" step="5" value="${offset}" aria-label="Offset in minutes" />
      <b>min</b>
      <button data-offset-step="5" aria-label="Increase offset by five minutes">${icon('plus')}</button>
    </span></label>
  `;
}

function renderOffsetControls(phase) {
  const offset = phase.start.offsetMinutes || 0;
  return `
    <div class="offset-panel">
      <input class="offset-slider" type="range" min="-180" max="180" step="5" value="${offset}" aria-label="Trigger offset" />
      <div class="offset-row">
        ${[-120, -60, -30, -15, 0, 15, 30, 60, 120].map((value) => `<button data-offset="${value}" class="${offset === value ? 'selected' : ''}">${offsetLabel(value)}</button>`).join('')}
      </div>
    </div>
  `;
}

function renderActionList(phase) {
  const actions = phase.actions || [];
  if (!actions.length) return '<div class="empty-action">No actions yet. Add one below.</div>';
  return actions.map((action, index) => {
    const definition = actionOption(action.type);
    const warning = action.type === 'videoWallpaper' && !appState.inventory.smartVideoPlugin;
    return `
      <div class="action-row ${warning ? 'warning' : ''}">
        <span class="action-icon">${icon(actionIconName(action.type))}</span>
        <span class="action-text">
          <b>${escapeHTML(definition?.label || action.type)}</b>
          <em>${escapeHTML(shortValue(action.value || ''))}${warning ? ' · plugin missing' : ''}</em>
        </span>
        <button class="ghost" data-remove-action="${index}" aria-label="Remove ${escapeAttr(definition?.label || action.type)}">${icon('trash')}</button>
      </div>
    `;
  }).join('');
}

function renderActionBuilder() {
  const first = appState.actionOptions[0];
  return `
    <div class="action-builder">
      <label class="field"><span>Action</span><select class="new-action-type">
        ${appState.actionOptions.map((item) => `<option value="${escapeAttr(item.type)}">${escapeHTML(item.label)}</option>`).join('')}
      </select></label>
      <label class="field grow"><span>Value</span><input class="new-action-value" list="action-values" placeholder="${escapeAttr(first?.placeholder || 'value')}" /></label>
      <datalist id="action-values">${(first?.choices || []).map((choice) => `<option value="${escapeAttr(choice.id)}">${escapeHTML(choice.name || choice.id)}</option>`).join('')}</datalist>
      <button class="primary add-action-button" data-action="add-action">${icon('plus')}<span>Add</span></button>
      <div class="action-warning">${escapeHTML(first?.warning || '')}</div>
    </div>
  `;
}

function renderInspector() {
  const sunrise = solarEvent('sunrise');
  const sunset = solarEvent('sunset');
  const problems = appState.checks.filter((check) => !check.ok);
  return `
    <section class="side-section sun-summary">
      <div class="side-heading"><span class="eyebrow">Astronomy</span><h2>Today</h2></div>
      <button class="sun-card sunrise" data-trigger="sunrise">
        <span>${icon('sunrise')}</span><span><em>Sunrise</em><b>${escapeHTML(sunrise?.clock || '--:--')}</b></span>
      </button>
      <button class="sun-card sunset" data-trigger="sunset">
        <span>${icon('sunset')}</span><span><em>Sunset</em><b>${escapeHTML(sunset?.clock || '--:--')}</b></span>
      </button>
    </section>
    <section class="side-section transitions-section">
      <div class="side-heading inline"><h3>Schedule today</h3><span>${appState.transitions.length}</span></div>
      <div class="transition-stack">
        ${appState.transitions.map((item) => `
          <button class="transition-item ${item.phaseId === selectedID ? 'active' : ''}" data-select="${escapeAttr(item.phaseId)}">
            <span style="background:${escapeAttr(item.color || '#4e79a7')}"></span>
            <b>${escapeHTML(displayTransitionClock(item))}</b>
            <em>${escapeHTML(item.phaseName)}</em>
          </button>
        `).join('') || '<div class="empty-list compact">No enabled rules</div>'}
      </div>
    </section>
    <button class="health-summary ${problems.length ? 'attention' : 'healthy'}" data-open-settings="system">
      <span>${icon(problems.length ? 'warning' : 'check')}</span>
      <span><b>${problems.length ? `${problems.length} system ${problems.length === 1 ? 'notice' : 'notices'}` : 'System ready'}</b><em>View diagnostics</em></span>
      ${icon('chevron')}
    </button>
  `;
}

function renderInspectorSheet() {
  return `
    <div class="sheet-backdrop" data-close-sheet>
      <section class="side-sheet compact-sheet" role="dialog" aria-modal="true" aria-labelledby="today-sheet-title" tabindex="-1">
        <div class="sheet-head"><div><span class="eyebrow">At a glance</span><h2 id="today-sheet-title">Today</h2></div><button class="icon-button quiet" data-action="close-sheet" aria-label="Close">${icon('close')}</button></div>
        <div class="sheet-scroll inspector-sheet-content">${renderInspector()}</div>
      </section>
    </div>
  `;
}

function renderSettingsSheet() {
  return `
    <div class="sheet-backdrop" data-close-sheet>
      <section class="side-sheet settings-sheet" role="dialog" aria-modal="true" aria-labelledby="settings-title" tabindex="-1">
        <div class="sheet-head">
          <div><span class="eyebrow">ThemeTime</span><h2 id="settings-title">Settings</h2></div>
          <button class="icon-button quiet" data-action="close-sheet" aria-label="Close settings">${icon('close')}</button>
        </div>
        <div class="sheet-tabs" role="tablist">
          <button class="${settingsTab === 'location' ? 'active' : ''}" data-settings-tab="location" role="tab" aria-selected="${settingsTab === 'location'}">${icon('location')}<span>Location</span></button>
          <button class="${settingsTab === 'system' ? 'active' : ''}" data-settings-tab="system" role="tab" aria-selected="${settingsTab === 'system'}">${icon('health')}<span>System</span></button>
        </div>
        <div class="sheet-scroll">
          ${settingsTab === 'location' ? renderLocationSettings() : renderSystemSettings()}
        </div>
      </section>
    </div>
  `;
}

function renderLocationSettings() {
  const draft = locationDraft || locationDraftFrom(appState.config.location);
  return `
    <form class="settings-form location-form">
      <div class="settings-intro"><span class="settings-icon">${icon('location')}</span><div><h3>Solar location</h3><p>ThemeTime uses these coordinates and timezone to calculate every twilight event.</p></div></div>
      <label class="field"><span>Location name</span><input name="label" value="${escapeAttr(draft.label)}" autocomplete="off" /></label>
      <div class="coordinate-grid">
        <label class="field"><span>Latitude</span><input name="latitude" inputmode="decimal" value="${escapeAttr(draft.latitude)}" /></label>
        <label class="field"><span>Longitude</span><input name="longitude" inputmode="decimal" value="${escapeAttr(draft.longitude)}" /></label>
      </div>
      <label class="field"><span>IANA timezone</span><input name="timezone" value="${escapeAttr(draft.timezone)}" spellcheck="false" placeholder="America/New_York" /></label>
      <div class="location-error" role="alert"></div>
      <div class="location-preview">${renderLocationPreview()}</div>
      <div class="sheet-actions"><button type="button" class="primary" data-action="save-location">${icon('save')}<span>Save configuration</span></button></div>
    </form>
  `;
}

function renderLocationPreview() {
  if (!locationPreview) return `<div class="preview-placeholder"><span>${icon('sunrise')}</span><p>Edit the fields to preview today’s solar schedule.</p></div>`;
  const sunrise = locationPreview.solarEvents?.find((event) => event.kind === 'sunrise');
  const sunset = locationPreview.solarEvents?.find((event) => event.kind === 'sunset');
  return `
    <div class="preview-head"><span>Preview</span><b>${escapeHTML(locationPreview.today)}</b></div>
    <div class="preview-times">
      <span>${icon('sunrise')}<em>Sunrise</em><b>${escapeHTML(sunrise?.clock || '--:--')}</b></span>
      <span>${icon('sunset')}<em>Sunset</em><b>${escapeHTML(sunset?.clock || '--:--')}</b></span>
    </div>
    <div class="preview-transitions">${(locationPreview.transitions || []).map((item) => `<span><i style="background:${escapeAttr(item.color || '#4e79a7')}"></i><em>${escapeHTML(item.phaseName)}</em><b>${escapeHTML(item.clock)}</b></span>`).join('')}</div>
  `;
}

export function renderLocationPreviewHTML(context) {
  useContext(context);
  return renderLocationPreview();
}

function renderSystemSettings() {
  return `
    <div class="system-settings">
      <div class="settings-intro"><span class="settings-icon">${icon('health')}</span><div><h3>System health</h3><p>ThemeTime checks the KDE tools and services used by your scheduled actions.</p></div></div>
      <div class="health-list">
        ${appState.checks.map((check) => `
          <div class="health-row ${check.ok ? 'ok' : check.warning ? 'warn' : 'bad'}">
            <span class="health-dot">${icon(check.ok ? 'check' : check.warning ? 'warning' : 'close')}</span>
            <span><b>${escapeHTML(check.name)}</b><em>${escapeHTML(check.detail || '')}</em></span>
          </div>
        `).join('')}
      </div>
      <div class="sheet-actions split"><button class="secondary" data-action="refresh-doctor">${icon('refresh')}<span>Refresh checks</span></button><button class="primary" data-action="install-service">${icon('download')}<span>Install user service</span></button></div>
    </div>
  `;
}

function selectedPhase() {
  return appState?.config?.phases?.find((phase) => phase.id === selectedID);
}

function solarEvent(kind) {
  return appState.solarEvents.find((event) => event.kind === kind);
}

function actionOption(type) {
  return findActionOption(appState, type);
}

function locationLabel() {
  return appState.config.location.label || appState.config.location.timezone;
}
