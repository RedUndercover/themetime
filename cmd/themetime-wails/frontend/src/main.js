import './styles.css';

import { backend, waitForBackend } from './backend.js';
import {
  actionOption,
  addAction,
  clamp,
  createRule,
  escapeAttr,
  escapeHTML,
  locationDraftFrom,
  normaliseConfig,
  parseLocationDraft,
  phaseByID,
  removeAction,
  removeRule,
  setStart,
} from './domain.js';
import { icon } from './icons.js';
import { renderApp, renderLocationPreviewHTML } from './views.js';

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

bindEvents();
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

async function refreshState() {
  if (dirty || openSheetName === 'settings') return;
  const selected = selectedID;
  try {
    appState = await backend().GetState();
    selectedID = phaseByID(appState, selected) ? selected : appState.config.phases?.[0]?.id || '';
    render();
  } catch (error) {
    setStatus(error.message || String(error), 'warn');
  }
}

function viewContext() {
  return { appState, selectedID, openSheetName, settingsTab, locationDraft, locationPreview };
}

function render() {
  root.innerHTML = renderApp(viewContext());
}

function bindEvents() {
  root.addEventListener('click', handleClick);
  root.addEventListener('input', handleInput);
  root.addEventListener('change', handleChange);
  root.addEventListener('keydown', handleRootKeyDown);
}

function handleClick(event) {
  const target = event.target.closest('[data-select], [data-trigger], [data-open-settings], [data-open-sheet], [data-close-sheet], [data-settings-tab], [data-color], [data-preset-kind], [data-offset], [data-offset-step], [data-remove-action], [data-action]');
  if (!target || !root.contains(target)) return;

  if (target.hasAttribute('data-close-sheet')) {
    if (event.target === target) closeSheet();
    return;
  }
  if (target.dataset.select !== undefined) {
    selectedID = target.dataset.select;
    render();
    return;
  }
  if (target.dataset.trigger !== undefined) {
    const phase = selectedPhase();
    if (phase && target.dataset.trigger) {
      setStart(phase, target.dataset.trigger, phase.start.offsetMinutes || 0);
      renderDirty();
    }
    return;
  }
  if (target.dataset.openSettings !== undefined) {
    openSettings(target.dataset.openSettings, target);
    return;
  }
  if (target.dataset.openSheet !== undefined) {
    openSheet(target.dataset.openSheet, target);
    return;
  }
  if (target.dataset.settingsTab !== undefined) {
    settingsTab = target.dataset.settingsTab;
    render();
    return;
  }
  if (target.dataset.color !== undefined) {
    selectedPhase().color = target.dataset.color;
    renderDirty();
    return;
  }
  if (target.dataset.presetKind !== undefined) {
    const phase = selectedPhase();
    setStart(phase, target.dataset.presetKind, Number(target.dataset.presetOffset || 0));
    phase.color = target.dataset.presetColor;
    renderDirty();
    return;
  }
  if (target.dataset.offset !== undefined) {
    setOffset(Number(target.dataset.offset));
    return;
  }
  if (target.dataset.offsetStep !== undefined) {
    setOffset(clamp((selectedPhase().start.offsetMinutes || 0) + Number(target.dataset.offsetStep), -180, 180));
    return;
  }
  if (target.dataset.removeAction !== undefined) {
    removeAction(selectedPhase(), Number(target.dataset.removeAction));
    renderDirty();
    return;
  }

  const actions = {
    'close-sheet': closeSheet,
    'new-rule': newRule,
    'remove-rule': removeSelectedRule,
    'apply-rule': applyRule,
    save: () => saveConfig('Configuration saved.'),
    'add-action': addSelectedAction,
    'save-location': saveLocation,
    'refresh-doctor': refreshDoctor,
    'install-service': installService,
  };
  actions[target.dataset.action]?.();
}

function handleInput(event) {
  const target = event.target;
  if (target.matches('.rule-name')) {
    selectedPhase().name = target.value;
    dirty = true;
    return;
  }
  if (target.matches('.color-input')) {
    selectedPhase().color = target.value;
    dirty = true;
    return;
  }
  if (target.matches('.offset-slider')) {
    selectedPhase().start.offsetMinutes = Number(target.value);
    dirty = true;
    const number = root.querySelector('.offset-number');
    if (number) number.value = target.value;
    return;
  }
  if (target.closest('.location-form') && target.name) {
    updateLocationDraft(target);
  }
}

function handleChange(event) {
  const target = event.target;
  if (target.matches('.trigger-select')) {
    const phase = selectedPhase();
    setStart(phase, target.value, phase.start.offsetMinutes || 0);
    renderDirty();
  } else if (target.matches('.offset-number')) {
    setOffset(clamp(Number(target.value || 0), -180, 180));
  } else if (target.matches('.offset-slider')) {
    renderDirty();
  } else if (target.matches('.clock-hour, .clock-minute')) {
    syncClock();
  } else if (target.matches('.new-action-type')) {
    updateActionBuilder(target.value);
  }
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
  if (!sheet) return;
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

function updateActionBuilder(type) {
  const definition = actionOption(appState, type);
  root.querySelector('.new-action-value').placeholder = definition?.placeholder || 'value';
  root.querySelector('.action-warning').textContent = definition?.warning || '';
  root.querySelector('#action-values').innerHTML = (definition?.choices || []).map((choice) => `<option value="${escapeAttr(choice.id)}">${escapeHTML(choice.name || choice.id)}</option>`).join('');
}

function addSelectedAction() {
  const type = root.querySelector('.new-action-type').value;
  const value = root.querySelector('.new-action-value').value.trim();
  if (!type || !value) {
    setStatus('Choose an action and value first.', 'warn');
    return;
  }
  addAction(selectedPhase(), type, value);
  renderDirty();
}

function newRule() {
  const phase = createRule(appState.config, appState.inventory);
  appState.config.phases.push(phase);
  selectedID = phase.id;
  renderDirty();
}

function removeSelectedRule() {
  selectedID = removeRule(appState.config, selectedID);
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
    selectedID = phaseByID(appState, selected) ? selected : appState.config.phases?.[0]?.id || '';
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
  locationDraft = locationDraftFrom(appState.config.location);
  locationPreview = null;
}

function updateLocationDraft(target) {
  ensureLocationDraft();
  locationDraft[target.name] = target.value;
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
    if (previewNode) previewNode.innerHTML = renderLocationPreviewHTML(viewContext());
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
  return parseLocationDraft(locationDraft);
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
  return phaseByID(appState, selectedID);
}
