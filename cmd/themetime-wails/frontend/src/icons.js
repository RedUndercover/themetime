import { escapeAttr } from './domain.js';

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

export function icon(name) {
  return `<svg class="icon icon-${escapeAttr(name)}" viewBox="0 0 24 24" aria-hidden="true" focusable="false">${paths[name] || paths.diamond}</svg>`;
}
