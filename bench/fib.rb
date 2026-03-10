def fib(n)
  return n if n <= 1
  a, b = 0, 1
  (2..n).each do |i|
    a, b = b, a + b
  end
  b
end

start = (Time.now.to_f * 1000).to_i
result = nil
10000.times do
  result = fib(30)
end
elapsed = (Time.now.to_f * 1000).to_i - start

puts "fib(30) iterative x10000 = #{result} in #{elapsed}ms"
