import './styles.css';

const triggerLabels = {
  clock: 'Clock',
  astronomicalDawn: 'Astronomical dawn',
  nauticalDawn: 'Nautical dawn',
  civilDawn: 'Civil dawn',
  sunrise: 'Sunrise',
  solarNoon: 'Solar noon',
  sunset: 'Sunset',
  civilDusk: 'Civil dusk',
  nauticalDusk: 'Nautical dusk',
  astronomicalDusk: 'Astronomical dusk',
};

const solarKinds = [
  'astronomicalDawn',
  'nauticalDawn',
  'civilDawn',
  'sunrise',
  'solarNoon',
  'sunset',
  'civilDusk',
  'nauticalDusk',
  'astronomicalDusk',
];

const colorPresets = [
  ['Night', '#1b213f'],
  ['Blue', '#356fe5'],
  ['Dawn', '#cf8b62'],
  ['Day', '#58a9d6'],
  ['Gold', '#f2b84b'],
  ['Dusk', '#805a9e'],
  ['Evening', '#324063'],
  ['Neutral', '#4e79a7'],
];

const startPresets = [
  ['Blue dawn', 'nauticalDawn', 0, '#2b6f9f', 'diamond'],
  ['Golden dawn', 'sunrise', 35, '#f2b84b', 'sparkle'],
  ['Golden dusk', 'sunset', -35, '#f48153', 'sparkle'],
  ['Blue dusk', 'civilDusk', 0, '#805a9e', 'diamond'],
];

let appState = null;
let selectedID = '';
let dirty = false;
let openSheetName = '';
let settingsTab = 'location';
let locationDraft = null;
let locationPreview = null;
let locationPreviewTimer = 0;
let locationPreviewSequence = 0;
let restoreFocus = null;

const root = document.getElementById('app');

init();

async function init() {
  root.innerHTML = `<div class="boot"><span class="brand-mark large"></span><b>ThemeTime</b></div>`;
  try {
    await waitForBackend();
    appState = await backend().GetState();
    selectedID = appState.config.phases?.[0]?.id || '';
    render();
    window.setInterval(refreshState, 60_000);
  } catch (error) {
    root.innerHTML = `<div class="fatal"><b>ThemeTime could not start.</b><span>${escapeHTML(error.message || String(error))}</span></div>`;
  }
}

function backend() {
  return window.go?.main?.App;
}

async function waitForBackend() {
  for (let i = 0; i < 80; i += 1) {
    if (backend()) return;
    await sleep(50);
  }
  throw new Error('Wails bindings are not available.');
}

function sleep(ms) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

async function refreshState() {
  if (dirty || openSheetName === 'settings') return;
  const selected = selectedID;
  try {
    appState = await backend().GetState();
    selectedID = phaseByID(selected) ? selected : appState.config.phases?.[0]?.id || '';
    render();
  } catch (error) {
    setStatus(error.message || String(error), 'warn');
  }
}

function render() {
  const active = appState.plan.active;
  const next = appState.plan.next;
  const selected = selectedPhase();
  root.innerHTML = `
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
  bindEvents();
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
          <em>${escapeHTML(transition ? displayTransitionClock(transition) : triggerText(phase.start))} · ${actionCount} ${actionCount === 1 ? 'action' : 'actions'}</em>
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
      <div class="section-heading"><div><span class="eyebrow">When</span><h3>Start time</h3></div><p>${escapeHTML(humanTrigger(phase.start))}</p></div>
      <div class="timing-controls">
        <label class="field"><span>Trigger</span><select class="trigger-select">
          ${['clock', ...solarKinds].map((kind) => option(kind, phase.start.kind, triggerLabels[kind])).join('')}
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
    const actionDefinition = actionOption(action.type);
    const warning = action.type === 'videoWallpaper' && !appState.inventory.smartVideoPlugin;
    return `
      <div class="action-row ${warning ? 'warning' : ''}">
        <span class="action-icon">${icon(actionIconName(action.type))}</span>
        <span class="action-text">
          <b>${escapeHTML(actionDefinition?.label || action.type)}</b>
          <em>${escapeHTML(shortValue(action.value || ''))}${warning ? ' · plugin missing' : ''}</em>
        </span>
        <button class="ghost" data-remove-action="${index}" aria-label="Remove ${escapeAttr(actionDefinition?.label || action.type)}">${icon('trash')}</button>
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
  ensureLocationDraft();
  return `
    <form class="settings-form location-form">
      <div class="settings-intro"><span class="settings-icon">${icon('location')}</span><div><h3>Solar location</h3><p>ThemeTime uses these coordinates and timezone to calculate every twilight event.</p></div></div>
      <label class="field"><span>Location name</span><input name="label" value="${escapeAttr(locationDraft.label)}" autocomplete="off" /></label>
      <div class="coordinate-grid">
        <label class="field"><span>Latitude</span><input name="latitude" inputmode="decimal" value="${escapeAttr(locationDraft.latitude)}" /></label>
        <label class="field"><span>Longitude</span><input name="longitude" inputmode="decimal" value="${escapeAttr(locationDraft.longitude)}" /></label>
      </div>
      <label class="field"><span>IANA timezone</span><input name="timezone" value="${escapeAttr(locationDraft.timezone)}" spellcheck="false" placeholder="America/New_York" /></label>
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

function bindEvents() {
  root.querySelectorAll('[data-select]').forEach((button) => {
    button.addEventListener('click', () => {
      selectedID = button.dataset.select;
      render();
    });
  });
  root.querySelectorAll('[data-trigger]').forEach((button) => {
    button.addEventListener('click', () => {
      const phase = selectedPhase();
      if (!phase || !button.dataset.trigger) return;
      setStart(phase, button.dataset.trigger, phase.start.offsetMinutes || 0);
      renderDirty();
    });
  });
  root.querySelectorAll('[data-open-settings]').forEach((button) => button.addEventListener('click', () => openSettings(button.dataset.openSettings, button)));
  root.querySelectorAll('[data-open-sheet="inspector"]').forEach((button) => button.addEventListener('click', () => openSheet('inspector', button)));
  root.querySelectorAll('[data-action="close-sheet"]').forEach((button) => button.addEventListener('click', closeSheet));
  root.querySelectorAll('[data-close-sheet]').forEach((backdrop) => backdrop.addEventListener('click', (event) => {
    if (event.target === backdrop) closeSheet();
  }));
  root.querySelectorAll('[data-settings-tab]').forEach((button) => button.addEventListener('click', () => {
    settingsTab = button.dataset.settingsTab;
    render();
  }));
  root.querySelectorAll('[data-action="new-rule"]').forEach((button) => button.addEventListener('click', newRule));
  root.querySelector('[data-action="remove-rule"]')?.addEventListener('click', removeRule);
  root.querySelector('[data-action="apply-rule"]')?.addEventListener('click', applyRule);
  root.querySelector('[data-action="save"]')?.addEventListener('click', () => saveConfig('Configuration saved.'));
  root.querySelector('.rule-name')?.addEventListener('input', (event) => {
    selectedPhase().name = event.target.value;
    dirty = true;
  });
  root.querySelector('.color-input')?.addEventListener('input', (event) => {
    selectedPhase().color = event.target.value;
    dirty = true;
  });
  root.querySelector('.trigger-select')?.addEventListener('change', (event) => {
    setStart(selectedPhase(), event.target.value, selectedPhase().start.offsetMinutes || 0);
    renderDirty();
  });
  root.querySelectorAll('[data-color]').forEach((button) => button.addEventListener('click', () => {
    selectedPhase().color = button.dataset.color;
    renderDirty();
  }));
  root.querySelectorAll('[data-preset-kind]').forEach((button) => button.addEventListener('click', () => {
    const phase = selectedPhase();
    setStart(phase, button.dataset.presetKind, Number(button.dataset.presetOffset || 0));
    phase.color = button.dataset.presetColor;
    renderDirty();
  }));
  root.querySelectorAll('[data-offset]').forEach((button) => button.addEventListener('click', () => setOffset(Number(button.dataset.offset))));
  root.querySelectorAll('[data-offset-step]').forEach((button) => button.addEventListener('click', () => setOffset(clamp((selectedPhase().start.offsetMinutes || 0) + Number(button.dataset.offsetStep), -180, 180))));
  root.querySelector('.offset-number')?.addEventListener('change', (event) => setOffset(clamp(Number(event.target.value || 0), -180, 180)));
  root.querySelector('.offset-slider')?.addEventListener('input', (event) => {
    selectedPhase().start.offsetMinutes = Number(event.target.value);
    dirty = true;
    root.querySelector('.offset-number')?.setAttribute('value', event.target.value);
  });
  root.querySelector('.offset-slider')?.addEventListener('change', renderDirty);
  root.querySelector('.clock-hour')?.addEventListener('change', syncClock);
  root.querySelector('.clock-minute')?.addEventListener('change', syncClock);
  root.querySelectorAll('[data-remove-action]').forEach((button) => button.addEventListener('click', () => {
    selectedPhase().actions.splice(Number(button.dataset.removeAction), 1);
    renderDirty();
  }));
  root.querySelector('.new-action-type')?.addEventListener('change', updateActionBuilder);
  root.querySelector('[data-action="add-action"]')?.addEventListener('click', addAction);
  root.querySelector('.location-form')?.addEventListener('input', updateLocationDraft);
  root.querySelector('[data-action="save-location"]')?.addEventListener('click', saveLocation);
  root.querySelector('[data-action="refresh-doctor"]')?.addEventListener('click', refreshDoctor);
  root.querySelector('[data-action="install-service"]')?.addEventListener('click', installService);
  root.onkeydown = handleRootKeyDown;
}

function openSettings(tab, opener) {
  settingsTab = tab || 'location';
  ensureLocationDraft(true);
  openSheet('settings', opener);
  if (settingsTab === 'location') scheduleLocationPreview();
}

function openSheet(name, opener) {
  openSheetName = name;
  restoreFocus = opener || document.activeElement;
  render();
  window.requestAnimationFrame(() => root.querySelector('.side-sheet')?.focus());
}

function closeSheet() {
  openSheetName = '';
  window.clearTimeout(locationPreviewTimer);
  render();
  window.requestAnimationFrame(() => restoreFocus?.focus?.());
}

function handleRootKeyDown(event) {
  if (!openSheetName) return;
  if (event.key === 'Escape') {
    event.preventDefault();
    closeSheet();
    return;
  }
  if (event.key !== 'Tab') return;
  const sheet = root.querySelector('.side-sheet');
  const focusable = [...sheet.querySelectorAll('button:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex="0"]')];
  if (!focusable.length) return;
  const first = focusable[0];
  const last = focusable[focusable.length - 1];
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault();
    first.focus();
  }
}

function renderDirty() {
  dirty = true;
  render();
}

function setStart(phase, kind, offset) {
  phase.start.kind = kind;
  if (kind === 'clock') {
    phase.start.clock = phase.start.clock || '12:00';
    phase.start.offsetMinutes = 0;
  } else {
    phase.start.clock = '';
    phase.start.offsetMinutes = Number(offset || 0);
  }
}

function setOffset(value) {
  selectedPhase().start.offsetMinutes = Math.round(Number(value || 0) / 5) * 5;
  renderDirty();
}

function syncClock() {
  const hour = root.querySelector('.clock-hour')?.value || '12';
  const minute = root.querySelector('.clock-minute')?.value || '00';
  const phase = selectedPhase();
  phase.start.kind = 'clock';
  phase.start.clock = `${hour}:${minute}`;
  phase.start.offsetMinutes = 0;
  renderDirty();
}

function updateActionBuilder() {
  const type = root.querySelector('.new-action-type').value;
  const actionDefinition = actionOption(type);
  root.querySelector('.new-action-value').placeholder = actionDefinition?.placeholder || 'value';
  root.querySelector('.action-warning').textContent = actionDefinition?.warning || '';
  root.querySelector('#action-values').innerHTML = (actionDefinition?.choices || []).map((choice) => `<option value="${escapeAttr(choice.id)}">${escapeHTML(choice.name || choice.id)}</option>`).join('');
}

function addAction() {
  const type = root.querySelector('.new-action-type').value;
  const value = root.querySelector('.new-action-value').value.trim();
  if (!type || !value) {
    setStatus('Choose an action and value first.', 'warn');
    return;
  }
  selectedPhase().actions = selectedPhase().actions || [];
  selectedPhase().actions.push({ type, value });
  renderDirty();
}

function newRule() {
  const id = nextID('rule');
  const phase = {
    id,
    name: 'New rule',
    color: '#4e79a7',
    enabled: true,
    start: { kind: 'clock', clock: '12:00' },
    actions: [{ type: 'colorScheme', value: appState.inventory.colorSchemes?.[0]?.id || 'BreezeLight' }],
  };
  appState.config.phases.push(phase);
  selectedID = id;
  renderDirty();
}

function removeRule() {
  const index = appState.config.phases.findIndex((phase) => phase.id === selectedID);
  if (index < 0) return;
  appState.config.phases.splice(index, 1);
  selectedID = appState.config.phases[Math.max(0, index - 1)]?.id || '';
  renderDirty();
}

async function applyRule() {
  if (!selectedID) return;
  try {
    const result = await backend().ApplyPhase(selectedID);
    const failed = result.filter((item) => item.error);
    setStatus(failed.length ? failed[0].error : 'Rule applied.', failed.length ? 'warn' : 'ok');
  } catch (error) {
    setStatus(error.message || String(error), 'warn');
  }
}

async function saveConfig(message = 'Saved.') {
  try {
    normaliseConfig(appState.config);
    const selected = selectedID;
    appState = await backend().SaveConfig(appState.config);
    selectedID = phaseByID(selected) ? selected : appState.config.phases?.[0]?.id || '';
    dirty = false;
    locationDraft = null;
    setStatus(message, 'ok');
    render();
  } catch (error) {
    setStatus(error.message || String(error), 'warn');
  }
}

function ensureLocationDraft(reset = false) {
  if (locationDraft && !reset) return;
  const location = appState.config.location;
  locationDraft = {
    label: location.label || '',
    latitude: String(location.latitude ?? ''),
    longitude: String(location.longitude ?? ''),
    timezone: location.timezone || '',
  };
  locationPreview = null;
}

function updateLocationDraft(event) {
  if (!event.target.name) return;
  locationDraft[event.target.name] = event.target.value;
  scheduleLocationPreview();
}

function scheduleLocationPreview() {
  window.clearTimeout(locationPreviewTimer);
  locationPreviewTimer = window.setTimeout(previewLocation, 250);
}

async function previewLocation() {
  const parsed = parsedLocation();
  const errorNode = root.querySelector('.location-error');
  if (parsed.error) {
    if (errorNode) errorNode.textContent = parsed.error;
    return;
  }
  if (errorNode) errorNode.textContent = '';
  const sequence = ++locationPreviewSequence;
  try {
    const preview = await backend().PreviewLocation(parsed.location);
    if (sequence !== locationPreviewSequence || openSheetName !== 'settings') return;
    locationPreview = preview;
    const previewNode = root.querySelector('.location-preview');
    if (previewNode) previewNode.innerHTML = renderLocationPreview();
  } catch (error) {
    if (sequence !== locationPreviewSequence) return;
    if (errorNode) errorNode.textContent = error.message || String(error);
  }
}

async function saveLocation() {
  const parsed = parsedLocation();
  const errorNode = root.querySelector('.location-error');
  if (parsed.error) {
    if (errorNode) errorNode.textContent = parsed.error;
    return;
  }
  appState.config.location = parsed.location;
  dirty = true;
  openSheetName = '';
  await saveConfig('Location and schedule saved.');
}

function parsedLocation() {
  ensureLocationDraft();
  const latitude = Number(locationDraft.latitude);
  const longitude = Number(locationDraft.longitude);
  if (!Number.isFinite(latitude) || latitude < -90 || latitude > 90) return { error: 'Latitude must be a number from −90 to 90.' };
  if (!Number.isFinite(longitude) || longitude < -180 || longitude > 180) return { error: 'Longitude must be a number from −180 to 180.' };
  if (!locationDraft.timezone.trim()) return { error: 'Timezone is required.' };
  return {
    location: {
      label: locationDraft.label.trim() || locationDraft.timezone.trim(),
      latitude,
      longitude,
      timezone: locationDraft.timezone.trim(),
      source: 'manual',
    },
  };
}

async function refreshDoctor() {
  try {
    appState.checks = await backend().RunDoctor();
    render();
    setStatus('System checks refreshed.', 'ok');
  } catch (error) {
    setStatus(error.message || String(error), 'warn');
  }
}

async function installService() {
  try {
    const path = await backend().InstallUserService();
    setStatus(`User service installed: ${path}`, 'ok');
    appState.checks = await backend().RunDoctor();
    render();
  } catch (error) {
    setStatus(error.message || String(error), 'warn');
  }
}

function normaliseConfig(config) {
  config.phases.forEach((phase) => {
    phase.name = phase.name?.trim() || 'Untitled rule';
    phase.color = phase.color?.trim() || '#4e79a7';
    if (phase.start.kind === 'clock') {
      phase.start.clock = phase.start.clock || '12:00';
      delete phase.start.offsetMinutes;
    } else {
      phase.start.clock = '';
      phase.start.offsetMinutes = Number(phase.start.offsetMinutes || 0);
    }
  });
}

function setStatus(message, kind) {
  document.querySelector('.toast')?.remove();
  const toast = document.createElement('div');
  toast.className = `toast ${kind || ''}`;
  toast.setAttribute('role', 'status');
  toast.setAttribute('aria-live', 'polite');
  toast.innerHTML = `${icon(kind === 'ok' ? 'check' : 'warning')}<span>${escapeHTML(message)}</span>`;
  document.body.appendChild(toast);
  window.setTimeout(() => toast.remove(), 4400);
}

function selectedPhase() {
  return phaseByID(selectedID);
}

function phaseByID(id) {
  return appState?.config?.phases?.find((phase) => phase.id === id);
}

function solarEvent(kind) {
  return appState.solarEvents.find((event) => event.kind === kind);
}

function actionOption(type) {
  return appState.actionOptions.find((item) => item.type === type);
}

function nextID(prefix) {
  const used = new Set(appState.config.phases.map((phase) => phase.id));
  for (let i = 1; ; i += 1) {
    const id = `${prefix}-${i}`;
    if (!used.has(id)) return id;
  }
}

function triggerText(start) {
  if (start.kind === 'clock') return start.clock || '12:00';
  const offset = start.offsetMinutes || 0;
  return `${triggerLabels[start.kind] || start.kind}${offset ? ` ${offsetLabel(offset)}` : ''}`;
}

function humanTrigger(start) {
  if (start.kind === 'clock') return `At ${start.clock || '12:00'}`;
  const label = triggerLabels[start.kind] || start.kind;
  const offset = Number(start.offsetMinutes || 0);
  if (!offset) return `At ${label.toLowerCase()}`;
  return `${Math.abs(offset)} minutes ${offset < 0 ? 'before' : 'after'} ${label.toLowerCase()}`;
}

function triggerIcon(kind) {
  if (kind === 'clock') return 'clock';
  if (kind === 'sunrise' || kind.endsWith('Dawn')) return 'sunrise';
  if (kind === 'sunset' || kind.endsWith('Dusk')) return 'sunset';
  if (kind === 'solarNoon') return 'sun';
  return 'moon';
}

function actionIconName(type) {
  if (type === 'staticWallpaper') return 'image';
  if (type === 'videoWallpaper') return 'play';
  if (type === 'accentColor' || type === 'colorScheme') return 'palette';
  if (type === 'customCommand') return 'terminal';
  return 'diamond';
}

function minuteOfDay(value) {
  const date = new Date(value);
  return date.getHours() * 60 + date.getMinutes() + date.getSeconds() / 60;
}

function percent(minute) {
  return Math.max(0, Math.min(100, (minute / 1440) * 100));
}

function offsetClock(minute) {
  if (minute >= 1440) return '24:00';
  return `${two(Math.floor(minute / 60))}:${two(Math.round(minute % 60))}`;
}

function clockOf(value) {
  const date = new Date(value);
  return `${two(date.getHours())}:${two(date.getMinutes())}`;
}

function displayTransitionClock(transition) {
  if (!transition?.at) return transition?.clock || '--:--';
  return offsetClock(minuteOfDay(transition.at));
}

function offsetLabel(value) {
  if (!value) return '0m';
  return `${value > 0 ? '+' : ''}${value}m`;
}

function shortValue(value) {
  if (!value) return '';
  const parts = value.split(/[\\/]/);
  return parts[parts.length - 1] || value;
}

function locationLabel() {
  return appState.config.location.label || appState.config.location.timezone;
}

function range(size) {
  return Array.from({ length: size }, (_, index) => index);
}

function two(value) {
  return String(value).padStart(2, '0');
}

function option(value, selected, label = value) {
  return `<option value="${escapeAttr(value)}" ${value === selected ? 'selected' : ''}>${escapeHTML(label)}</option>`;
}

function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, value));
}

function icon(name) {
  const paths = {
    plus: '<path d="M12 5v14M5 12h14"/>',
    minus: '<path d="M5 12h14"/>',
    close: '<path d="m6 6 12 12M18 6 6 18"/>',
    chevron: '<path d="m9 7 5 5-5 5"/>',
    settings: '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06-2.83 2.83-.06-.06a1.7 1.7 0 0 0-1.88-.34 1.7 1.7 0 0 0-1.03 1.55V21h-4v-.08A1.7 1.7 0 0 0 9 19.37a1.7 1.7 0 0 0-1.88.34l-.06.06-2.83-2.83.06-.06A1.7 1.7 0 0 0 4.63 15 1.7 1.7 0 0 0 3.08 14H3v-4h.08A1.7 1.7 0 0 0 4.63 9a1.7 1.7 0 0 0-.34-1.88l-.06-.06 2.83-2.83.06.06A1.7 1.7 0 0 0 9 4.63h.01A1.7 1.7 0 0 0 10 3.08V3h4v.08a1.7 1.7 0 0 0 1.03 1.55 1.7 1.7 0 0 0 1.88-.34l.06-.06 2.83 2.83-.06.06A1.7 1.7 0 0 0 19.37 9v.01A1.7 1.7 0 0 0 20.92 10H21v4h-.08A1.7 1.7 0 0 0 19.4 15Z"/>',
    location: '<path d="M20 10c0 5-8 11-8 11S4 15 4 10a8 8 0 1 1 16 0Z"/><circle cx="12" cy="10" r="2.5"/>',
    clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
    play: '<path d="m8 5 11 7-11 7Z"/>',
    trash: '<path d="M4 7h16M9 7V4h6v3M7 7l1 14h8l1-14M10 11v6M14 11v6"/>',
    save: '<path d="M5 3h12l2 2v16H5Z"/><path d="M8 3v6h8V3M8 21v-7h8v7"/>',
    tune: '<path d="M4 7h10M18 7h2M4 17h2M10 17h10"/><circle cx="16" cy="7" r="2"/><circle cx="8" cy="17" r="2"/>',
    palette: '<path d="M12 3a9 9 0 0 0 0 18h1.5a1.5 1.5 0 0 0 0-3H12a2 2 0 0 1 0-4h3a6 6 0 0 0 6-6c0-3-4-5-9-5Z"/><circle cx="7.5" cy="10" r=".8" fill="currentColor"/><circle cx="10" cy="6.5" r=".8" fill="currentColor"/><circle cx="15" cy="7" r=".8" fill="currentColor"/>',
    sunrise: '<path d="M4 18h16M6 14a6 6 0 0 1 12 0M12 3v3M4.9 6.9 7 9M19.1 6.9 17 9"/>',
    sunset: '<path d="M4 18h16M6 14a6 6 0 0 1 12 0M12 3v3M4.9 6.9 7 9M19.1 6.9 17 9M8 21h8"/>',
    sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2M5 5l1.5 1.5M17.5 17.5 19 19M19 5l-1.5 1.5M6.5 17.5 5 19"/>',
    moon: '<path d="M20 15.5A8.5 8.5 0 0 1 8.5 4 8.5 8.5 0 1 0 20 15.5Z"/>',
    diamond: '<path d="m12 3 8 9-8 9-8-9Z"/>',
    sparkle: '<path d="m12 2 1.8 6.2L20 10l-6.2 1.8L12 18l-1.8-6.2L4 10l6.2-1.8Z"/>',
    check: '<path d="m5 12 4 4L19 6"/>',
    warning: '<path d="M12 3 2.5 20h19Z"/><path d="M12 9v4M12 17h.01"/>',
    health: '<path d="M3 12h4l2-5 4 10 2-5h6"/><path d="M20 5a5 5 0 0 0-8-1 5 5 0 0 0-8 6c1 5 8 10 8 10s7-5 8-10a5 5 0 0 0 0-5Z"/>',
    refresh: '<path d="M20 7v5h-5M4 17v-5h5"/><path d="M6.1 8A7 7 0 0 1 18 7l2 5M17.9 16A7 7 0 0 1 6 17l-2-5"/>',
    download: '<path d="M12 3v12M7 10l5 5 5-5M5 21h14"/>',
    image: '<rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="8" cy="9" r="1.5"/><path d="m4 17 5-5 4 4 2-2 5 4"/>',
    terminal: '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="m7 9 3 3-3 3M13 15h4"/>',
  };
  return `<svg class="icon icon-${escapeAttr(name)}" viewBox="0 0 24 24" aria-hidden="true" focusable="false">${paths[name] || paths.diamond}</svg>`;
}

function escapeHTML(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

function escapeAttr(value) {
  return escapeHTML(value);
}
