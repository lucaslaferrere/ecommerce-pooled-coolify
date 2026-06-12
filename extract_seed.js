const fs = require('fs');
const raw = fs.readFileSync('C:/Users/lafer/.claude/projects/c--Users-lafer-commerce-repo/227fe0e4-9370-4484-828d-d190e57611e2.jsonl', 'utf8');
const marker = 'Cada producto ahora tiene benefits';
const pos = raw.lastIndexOf(marker);
if (pos < 0) { console.log('marker not found'); process.exit(1); }

// Find start of JSON object (the { before "role")
const rolePos = raw.lastIndexOf('"role":"user"', pos);
const objStart = raw.lastIndexOf('\n', rolePos) + 1;

// Find the seed content directly from the raw string
const seedStart = raw.indexOf('// SEED SCRIPT v2', pos);
if (seedStart < 0) { console.log('seed start not found'); process.exit(1); }

// Find end - look for Que opinas?" and then "}}
const qPos = raw.indexOf('Que opinas?', seedStart);
const endPos = qPos > 0 ? qPos : seedStart + 50000;

// Extract raw JSON string content
const rawContent = raw.substring(seedStart, endPos).trim();

// Unescape JSON string escapes (the content is inside a JSON string)
const unescaped = rawContent
  .replace(/\\n/g, '\n')
  .replace(/\\t/g, '\t')
  .replace(/\\"/g, '"')
  .replace(/\\\\/g, '\\');

fs.writeFileSync('seed_pooled_v2_raw.js', unescaped, 'utf8');
console.log('Written, lines:', unescaped.split('\n').length, 'chars:', unescaped.length);
