// Message fetching with pagination support
// Discord returns messages newest-first; we reverse to chronological order

async function fetchAllMessages(ctx, channelId, maxMessages, beforeId) {
  const allMessages = []
  let lastMessageId = null
  const pageSize = 100 // Discord max per request

  while (true) {
    const options = { limit: pageSize }
    if (lastMessageId) {
      options.before = lastMessageId
    }

    const batch = await ctx.discord.messages.list(channelId, options)
    if (!batch || batch.length === 0) {
      break
    }

    for (const msg of batch) {
      // If we hit the beforeId, stop collecting (exclusive)
      if (beforeId && msg.id === beforeId) {
        break
      }
      allMessages.push(msg)
    }

    // Check if the batch contained beforeId (we need to stop)
    if (beforeId && batch.some(m => m.id === beforeId)) {
      break
    }

    lastMessageId = batch[batch.length - 1].id

    if (maxMessages && allMessages.length >= maxMessages) {
      allMessages.splice(maxMessages)
      break
    }
  }

  // Discord returns newest first; reverse to chronological order
  return allMessages.reverse()
}

module.exports = { fetchAllMessages }
