const arr = Array.from({length: 10000}, () => Math.random() * 10000);

const start = Date.now();
const sorted = [...arr].sort((a, b) => a - b);
const elapsed = Date.now() - start;

console.log(`Sort 10k random numbers in ${elapsed}ms`);
console.log(`First 5: ${JSON.stringify(sorted.slice(0, 5))}`);
