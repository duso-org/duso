// Regex - RegExp (stdlib/language builtin), single realistic ops on a real doc
const fs = require('fs');
const doc = fs.readFileSync('../docs/learning-duso.md', 'utf8');
const bytes = Buffer.byteLength(doc, 'utf8');

let start = process.hrtime.bigint();
const hasDatastore = /datastore\(/.test(doc);
const containsMs = Number(process.hrtime.bigint() - start) / 1e6;

start = process.hrtime.bigint();
const words = doc.match(/\w+/g) || [];
const findMs = Number(process.hrtime.bigint() - start) / 1e6;

console.log(`contains() on ${bytes}-byte doc in ${containsMs}ms (found=${hasDatastore})`);
console.log(`find() on ${bytes}-byte doc in ${findMs}ms (${words.length} matches)`);
