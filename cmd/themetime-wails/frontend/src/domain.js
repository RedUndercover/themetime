export const colorPresets = [
  ['Night', '#1b213f'],
  ['Blue', '#356fe5'],
  ['Dawn', '#cf8b62'],
  ['Day', '#58a9d6'],
  ['Gold', '#f2b84b'],
  ['Dusk', '#805a9e'],
  ['Evening', '#324063'],
  ['Neutral', '#4e79a7'],
];

export const startPresets = [
  ['Blue dawn', 'nauticalDawn', 0, '#2b6f9f', 'diamond'],
  ['Golden dawn', 'sunrise', 35, '#f2b84b', 'sparkle'],
  ['Golden dusk', 'sunset', -35, '#f48153', 'sparkle'],
  ['Blue dusk', 'civilDusk', 0, '#805a9e', 'diamond'],
];

export function phaseByID(state, id) {
  return state?.config?.phases?.find((phase) => phase.id === id);
}

export function actionOption(state, type) {
  return state?.actionOptions?.find((item) => item.type === type);
}

export function triggerLabel(state, kind) {
  return state?.triggerOptions?.find((item) => item.kind === kind)?.label || kind;
}

export function solarKinds(state) {
  return (state?.triggerOptions || []).filter((item) => item.kind !== 'clock').map((item) => item.kind);
}

export function nextID(config, prefix) {
  const used = new Set((config.phases || []).map((phase) => phase.id));
  for (let index = 1; ; index += 1) {
    const id = `${prefix}-${index}`;
    if (!used.has(id)) return id;
  }
}

export function createRule(config, inventory) {
  return {
    id: nextID(config, 'rule'),
    name: 'New rule',
    color: '#4e79a7',
    enabled: true,
    start: { kind: 'clock', clock: '12:00' },
    actions: [{ type: 'colorScheme', value: inventory.colorSchemes?.[0]?.id || 'BreezeLight' }],
  };
}

export function removeRule(config, id) {
  const index = config.phases.findIndex((phase) => phase.id === id);
  if (index < 0) return id;
  config.phases.splice(index, 1);
  return config.phases[Math.max(0, index - 1)]?.id || '';
}

export function addAction(phase, type, value) {
  phase.actions = phase.actions || [];
  phase.actions.push({ type, value });
}

export function removeAction(phase, index) {
  phase.actions?.splice(index, 1);
}

export function setStart(phase, kind, offset = 0) {
  phase.start.kind = kind;
  if (kind === 'clock') {
    phase.start.clock = phase.start.clock || '12:00';
    phase.start.offsetMinutes = 0;
  } else {
    phase.start.clock = '';
    phase.start.offsetMinutes = Number(offset || 0);
  }
}

export function normaliseConfig(config) {
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

export function locationDraftFrom(location) {
  return {
    label: location.label || '',
    latitude: String(location.latitude ?? ''),
    longitude: String(location.longitude ?? ''),
    timezone: location.timezone || '',
  };
}

export function parseLocationDraft(draft) {
  const latitude = Number(draft.latitude);
  const longitude = Number(draft.longitude);
  if (!Number.isFinite(latitude) || latitude < -90 || latitude > 90) return { error: 'Latitude must be a number from −90 to 90.' };
  if (!Number.isFinite(longitude) || longitude < -180 || longitude > 180) return { error: 'Longitude must be a number from −180 to 180.' };
  if (!draft.timezone.trim()) return { error: 'Timezone is required.' };
  return {
    location: {
      label: draft.label.trim() || draft.timezone.trim(),
      latitude,
      longitude,
      timezone: draft.timezone.trim(),
      source: 'manual',
    },
  };
}

export function triggerText(state, start) {
  if (start.kind === 'clock') return start.clock || '12:00';
  const offset = start.offsetMinutes || 0;
  return `${triggerLabel(state, start.kind)}${offset ? ` ${offsetLabel(offset)}` : ''}`;
}

export function humanTrigger(state, start) {
  if (start.kind === 'clock') return `At ${start.clock || '12:00'}`;
  const label = triggerLabel(state, start.kind);
  const offset = Number(start.offsetMinutes || 0);
  if (!offset) return `At ${label.toLowerCase()}`;
  return `${Math.abs(offset)} minutes ${offset < 0 ? 'before' : 'after'} ${label.toLowerCase()}`;
}

export function triggerIcon(kind) {
  if (kind === 'clock') return 'clock';
  if (kind === 'sunrise' || kind.endsWith('Dawn')) return 'sunrise';
  if (kind === 'sunset' || kind.endsWith('Dusk')) return 'sunset';
  if (kind === 'solarNoon') return 'sun';
  return 'moon';
}

export function actionIconName(type) {
  if (type === 'staticWallpaper') return 'image';
  if (type === 'videoWallpaper') return 'play';
  if (type === 'accentColor' || type === 'colorScheme') return 'palette';
  if (type === 'customCommand') return 'terminal';
  return 'diamond';
}

export function minuteOfDay(value) {
  const date = new Date(value);
  return date.getHours() * 60 + date.getMinutes() + date.getSeconds() / 60;
}

export function percent(minute) {
  return Math.max(0, Math.min(100, (minute / 1440) * 100));
}

export function offsetClock(minute) {
  if (minute >= 1440) return '24:00';
  return `${two(Math.floor(minute / 60))}:${two(Math.round(minute % 60))}`;
}

export function clockOf(value) {
  const date = new Date(value);
  return `${two(date.getHours())}:${two(date.getMinutes())}`;
}

export function displayTransitionClock(transition) {
  if (!transition?.at) return transition?.clock || '--:--';
  return offsetClock(minuteOfDay(transition.at));
}

export function offsetLabel(value) {
  if (!value) return '0m';
  return `${value > 0 ? '+' : ''}${value}m`;
}

export function shortValue(value) {
  if (!value) return '';
  const parts = value.split(/[\\/]/);
  return parts[parts.length - 1] || value;
}

export function range(size) {
  return Array.from({ length: size }, (_, index) => index);
}

export function two(value) {
  return String(value).padStart(2, '0');
}

export function option(value, selected, label = value) {
  return `<option value="${escapeAttr(value)}" ${value === selected ? 'selected' : ''}>${escapeHTML(label)}</option>`;
}

export function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, value));
}

export function escapeHTML(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

export function escapeAttr(value) {
  return escapeHTML(value);
}
