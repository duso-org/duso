#!/usr/bin/env ruby
# Ruby benchmark server: GET /delay responds after 1 second,
# GET /ping responds immediately. Stdlib only, thread per connection
# with keep-alive (webrick is no longer stdlib in ruby 3.x).
require 'socket'

server = TCPServer.new('127.0.0.1', 8399)
server.listen(1024)

loop do
  Thread.new(server.accept) do |c|
    begin
      loop do
        line = c.gets
        break if line.nil?
        path = line.split[1]
        loop do
          h = c.gets
          break if h.nil? || h.strip.empty?
        end
        sleep 1 if path&.start_with?('/delay')
        c.write "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 2\r\n\r\nok"
      end
    rescue
    ensure
      c.close
    end
  end
end
