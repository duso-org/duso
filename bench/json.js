// JSON round-trip - JSON.parse/stringify (stdlib)
const n = 5000;
const objs = [];
for (let i = 1; i <= n; i++) {
  objs.push({ id: i, name: "user-" + i, active: i % 3 === 0, score: i * 1.5, tags: ["a", "b", "c"] });
}

let start = process.hrtime.bigint();
const text = JSON.stringify(objs);
const encoded = Number(process.hrtime.bigint() - start) / 1e6;

start = process.hrtime.bigint();
const back = JSON.parse(text);
const decoded = Number(process.hrtime.bigint() - start) / 1e6;

console.log(`JSON encode ${n} objects in ${encoded}ms, decode in ${decoded}ms`);
