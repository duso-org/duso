# Discord Module

Discord API integration for duso. Provides webhooks and Gateway client for real-time event handling.

## Usage

```duso
discord = require("discord")
```

## Functions

### post_webhook(url, payload)

Post a message to a Discord webhook.

**Parameters:**
- `url` (string) - Webhook URL from Discord
- `payload` (object or string) - Message payload (content, embeds, etc.) or plain text string

**Returns:** Boolean (true if successful)

**Example:**
```duso
discord.post_webhook("https://discordapp.com/api/webhooks/...", {
  content = "Hello from Duso!"
})
```

### session(config)

Create a Gateway client connection for real-time events.

**Parameters:**
- `config` (object) - Configuration:
  - `token` (string, required) - Discord bot token
  - `intents` (number, optional) - Gateway intents (default: guilds + guild_messages + message_content)

**Returns:** Session object

**Example:**
```duso
bot = discord.session({
  token = env("DISCORD_TOKEN")
})
```

## Gateway Intents

Intents control what events your bot receives. Combine intent constants with `+`:

```duso
discord = require("discord")

// Use individual intents
intents = discord.intents.guilds + discord.intents.guild_messages + discord.intents.message_content

bot = discord.session({
  token = env("DISCORD_TOKEN"),
  intents = intents
})
```

Available intents:
- `guilds` (1)
- `guild_members` (2)
- `guild_bans` (4)
- `guild_emojis_and_stickers` (8)
- `guild_integrations` (16)
- `guild_webhooks` (32)
- `guild_invites` (64)
- `guild_voice_states` (128)
- `guild_presences` (256)
- `guild_messages` (512)
- `guild_message_reactions` (1024)
- `guild_message_typing` (2048)
- `direct_messages` (4096)
- `direct_message_reactions` (8192)
- `direct_message_typing` (16384)
- `message_content` (32768)
- `guild_scheduled_events` (65536)
- `auto_moderation_rules` (1048576)
- `auto_moderation_execution` (2097152)

## Session Object

Returned by `session()`. Handles Gateway connection and event reading.

### Methods

- `read([timeout])` - Block until an event arrives. Returns event object with `type`, `data`, and `seq` fields, or nil on timeout.
- `heartbeat()` - Manually send a heartbeat (usually automatic).
- `send_message(channel_id, payload)` - Send a message to a channel via the REST API.
- `is_connected()` - Check if connection is still open.
- `close()` - Close the connection.

## Event Structure

Events returned by `read()` have this structure:

```duso
// Example event structure
event = {
  type = "MESSAGE_CREATE",
  data = {content = "hello", author = {id = 123}},
  seq = 1234
}
```

Common event types:
- `READY` - Bot has authenticated and is ready
- `GUILD_CREATE` - Guild information received
- `MESSAGE_CREATE` - New message in a channel
- `MESSAGE_UPDATE` - Message was edited
- `MESSAGE_DELETE` - Message was deleted
- `INTERACTION_CREATE` - Slash command or button interaction

## Example: Echo Bot

```duso
discord = require("discord")

bot = discord.session({
  token = env("DISCORD_TOKEN"),
  intents = discord.intents.guilds + discord.intents.guild_messages + discord.intents.message_content
})

print("Bot connected")

while bot.is_connected() do
  event = bot.read(timeout=30)
  
  if event == nil then
    continue
  end

  if event.type == "MESSAGE_CREATE" then
    msg = event.data
    
    // Don't respond to bot messages
    if msg.author.bot then
      continue
    end

    // Don't respond to messages in DMs
    if msg.guild_id == nil then
      continue
    end

    // Send reply
    bot.send_message(msg.channel_id, {
      content = "You said: " + msg.content
    })
  end
end

bot.close()
```

## Example: Webhook Alert

```duso
discord = require("discord")

webhook_url = "https://discordapp.com/api/webhooks/..."

discord.post_webhook(webhook_url, {
  content = "Alert: Server CPU at 90%",
  embeds = [
    {
      title = "System Alert",
      color = 16711680,  // Red
      fields = [
        {
          name = "Status",
          value = "Critical",
          inline = true
        },
        {
          name = "Timestamp",
          value = format_time(now()),
          inline = true
        }
      ]
    }
  ]
})
```

## Notes

- Bot token should be stored in environment variables (e.g., `env("DISCORD_TOKEN")`)
- Webhook URLs can be posted to without authentication
- Gateway requires a valid bot token and appropriate intents
- Heartbeat is handled automatically by the client
- For best performance, use the Gateway for real-time interaction and webhooks for simple notifications
- Discord API enforces rate limits; consider adding delays between rapid requests

## See Also

- [websocket() - WebSocket client](/docs/reference/websocket.md)
- [fetch() - HTTP requests](/docs/reference/fetch.md)
- [Discord Developer Portal](https://discord.com/developers/applications)
