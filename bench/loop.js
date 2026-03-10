const start = Date.now();

let sum = 0;
for (let i = 1; i <= 1000; i++) {
  for (let j = 1; j <= 1000; j++) {
    sum += i * j;
  }
}

const elapsed = Date.now() - start;

console.log(`Loop sum (1000x1000) = ${sum} in ${elapsed}ms`);
