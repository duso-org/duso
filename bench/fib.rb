def fib(n)
  return n if n <= 1
  fib(n - 1) + fib(n - 2)
end

start = (Time.now.to_f * 1000).to_i
result = fib(30)
elapsed = (Time.now.to_f * 1000).to_i - start

puts "fib(30) = #{result} in #{elapsed}ms"
