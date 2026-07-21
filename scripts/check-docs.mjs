import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const markdownFiles = [
  'README.md',
  ...markdownFilesIn('docs'),
  ...markdownFilesIn('wiki'),
].sort();

const failures = [];
let localLinkCount = 0;
let jsonBlockCount = 0;

for (const relativeFile of markdownFiles) {
  const absoluteFile = path.join(root, relativeFile);
  const markdown = fs.readFileSync(absoluteFile, 'utf8');

  checkFences(relativeFile, markdown);
  checkLinks(relativeFile, markdown);
  checkJSON(relativeFile, markdown);
}

if (failures.length > 0) {
  for (const failure of failures) {
    console.error(failure);
  }
  process.exit(1);
}

console.log(
  `Documentation OK: ${markdownFiles.length} Markdown files, ` +
    `${localLinkCount} local links, ${jsonBlockCount} JSON examples.`,
);

function markdownFilesIn(directory) {
  return fs
    .readdirSync(path.join(root, directory), { withFileTypes: true })
    .filter((entry) => entry.isFile() && entry.name.endsWith('.md'))
    .map((entry) => path.join(directory, entry.name));
}

function checkFences(relativeFile, markdown) {
  const fences = markdown
    .split('\n')
    .filter((line) => line.trimStart().startsWith('```')).length;
  if (fences % 2 !== 0) {
    failures.push(`${relativeFile}: unbalanced fenced code blocks`);
  }
}

function checkLinks(relativeFile, markdown) {
  for (const match of markdown.matchAll(/\[[^\]]*\]\(([^)]+)\)/g)) {
    const destination = match[1].trim().replace(/^</, '').replace(/>$/, '');
    if (!destination || /^(https?:|mailto:)/.test(destination)) {
      continue;
    }

    localLinkCount += 1;
    const hashIndex = destination.indexOf('#');
    const targetPart = hashIndex >= 0 ? destination.slice(0, hashIndex) : destination;
    const anchor = hashIndex >= 0 ? destination.slice(hashIndex + 1) : '';
    let resolved = targetPart
      ? path.normalize(path.join(path.dirname(relativeFile), targetPart))
      : relativeFile;

    if (!path.extname(resolved) && path.dirname(relativeFile) === 'wiki') {
      resolved += '.md';
    }

    const absoluteTarget = path.join(root, resolved);
    if (!fs.existsSync(absoluteTarget)) {
      failures.push(
        `${relativeFile}: missing local link ${destination} (resolved as ${resolved})`,
      );
      continue;
    }

    if (anchor && resolved.endsWith('.md')) {
      const targetMarkdown = fs.readFileSync(absoluteTarget, 'utf8');
      if (!headingAnchors(targetMarkdown).has(anchor.toLowerCase())) {
        failures.push(`${relativeFile}: missing #${anchor} in ${resolved}`);
      }
    }
  }
}

function headingAnchors(markdown) {
  const anchors = new Set();
  const duplicates = new Map();

  for (const line of markdown.split('\n')) {
    if (!/^#+\s+/.test(line)) {
      continue;
    }

    const base = line
      .replace(/^#+\s+/, '')
      .trim()
      .toLowerCase()
      .replace(/<[^>]+>/g, '')
      .replace(/[^\p{L}\p{N}\s_-]/gu, '')
      .replace(/\s+/g, '-')
      .replace(/-+/g, '-');
    const duplicate = duplicates.get(base) ?? 0;
    const anchor = duplicate === 0 ? base : `${base}-${duplicate}`;
    duplicates.set(base, duplicate + 1);
    anchors.add(anchor);
  }

  return anchors;
}

function checkJSON(relativeFile, markdown) {
  const fence = '`'.repeat(3);
  const pattern = new RegExp(`${fence}json\\s*\\n([\\s\\S]*?)\\n${fence}`, 'g');

  for (const match of markdown.matchAll(pattern)) {
    jsonBlockCount += 1;
    try {
      JSON.parse(match[1]);
    } catch (error) {
      failures.push(`${relativeFile}: invalid fenced JSON: ${error.message}`);
    }
  }
}
