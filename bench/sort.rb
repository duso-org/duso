arr = Array.new(10000) { rand * 10000 }

start = (Time.now.to_f * 1000).to_i
sorted = arr.sort
elapsed = (Time.now.to_f * 1000).to_i - start

puts "Sort 10k random numbers in #{elapsed}ms"
puts "First 5: #{sorted[0..4].inspect}"
