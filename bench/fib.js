function fib(n) {
  if (n <= 1) return n;
  return fib(n - 1) + fib(n - 2);
}

const start = Date.now();
const result = fib(30);
const elapsed = Date.now() - start;

console.log(`fib(30) = ${result} in ${elapsed}ms`);
