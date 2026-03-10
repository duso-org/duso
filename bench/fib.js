// Fibonacci iterative - 10000 iterations
function fib(n) {
  if (n <= 1) return n;
  let a = 0, b = 1;
  for (let i = 2; i <= n; i++) {
    const temp = a + b;
    a = b;
    b = temp;
  }
  return b;
}

const start = Date.now();
let result;
for (let i = 0; i < 10000; i++) {
  result = fib(30);
}
const elapsed = Date.now() - start;

console.log(`fib(30) iterative x10000 = ${result} in ${elapsed}ms`);
