#!/usr/bin/env ruby
require 'net/http'
require 'uri'
require 'thread'

def fetch_url(url, timeout = 15)
  uri = URI(url)
  Net::HTTP.start(uri.host, uri.port, use_ssl: uri.scheme == 'https', open_timeout: timeout, read_timeout: timeout) do |http|
    request = Net::HTTP::Get.new(uri.path)
    response = http.request(request)
    { ok: response.code.to_i == 200, status: response.code.to_i }
  end
rescue => e
  { ok: false, error: e.message }
end

def worker(worker_id)
  results = []
  5.times do |i|
    begin
      response = fetch_url('https://httpbin.org/delay/1', 15)
      results << { attempt: i + 1, status: response[:status] }
    rescue => e
      results << { attempt: i + 1, error: e.message }
    end
  end
  { worker_id: worker_id, requests: results.length }
end

def main
  puts "=== Ruby I/O-Bound Benchmark ==="
  puts "Each worker makes 5 HTTP requests to httpbin.org/delay/1"
  puts ""

  start = (Time.now.to_f * 1000).to_i

  # Spawn 500 workers using threads
  spawn_start = (Time.now.to_f * 1000).to_i
  threads = (1..500).map { |i| Thread.new { worker(i) } }
  spawn_time = (Time.now.to_f * 1000).to_i - spawn_start

  # Wait for all to complete
  wait_start = (Time.now.to_f * 1000).to_i
  results = threads.map(&:value)
  wait_time = (Time.now.to_f * 1000).to_i - wait_start

  total_time = (Time.now.to_f * 1000).to_i - start

  puts "Spawn 500 workers: #{spawn_time}ms"
  puts "Wait for completion: #{wait_time}ms"
  puts "Total time: #{total_time}ms"
  puts "Average per worker: #{(wait_time.to_f/500).round(3)}ms"
end

main
