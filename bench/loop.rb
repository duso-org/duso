start = (Time.now.to_f * 1000).to_i

sum = 0
1.upto(1000) do |i|
  1.upto(1000) do |j|
    sum += i * j
  end
end

elapsed = (Time.now.to_f * 1000).to_i - start

puts "Loop sum (1000x1000) = #{sum} in #{elapsed}ms"
