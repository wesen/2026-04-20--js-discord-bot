function trimText(value) {
  return String(value || "").trim()
}

function showLabel(show) {
  if (!show) {
    return "(unknown show)"
  }
  return `${trimText(show.artist) || "Unknown artist"} — ${trimText(show.displayDate) || trimText(show.dateISO) || "TBD"}`
}

function showAnnouncementPayload(show) {
  const title = `🎵 ${showLabel(show)}`
  const fields = [
    { name: "Artist", value: trimText(show.artist) || "TBD", inline: true },
    { name: "Date", value: trimText(show.displayDate) || trimText(show.dateISO) || "TBD", inline: true },
    { name: "Doors", value: trimText(show.doorsTime) || "TBD", inline: true },
    { name: "Age", value: trimText(show.ageRestriction) || "TBD", inline: true },
    { name: "Price", value: trimText(show.price) || "TBD", inline: true },
  ]
  if (trimText(show.notes)) {
    fields.push({ name: "Notes", value: trimText(show.notes), inline: false })
  }
  return {
    content: "",
    embeds: [
      {
        title,
        description: `Doors: ${trimText(show.doorsTime) || "TBD"}`,
        color: 0x5865f2,
        fields,
      },
    ],
  }
}

function upcomingShowsText(shows) {
  const list = Array.isArray(shows) ? shows : []
  if (list.length === 0) {
    return "📅 Upcoming Shows\n\nNo upcoming shows right now."
  }

  const lines = ["📅 Upcoming Shows", ""]
  list.forEach((show) => {
    lines.push(`🎵 ${trimText(show.artist) || "Unknown artist"} — ${trimText(show.displayDate) || trimText(show.dateISO) || "TBD"}`)
    lines.push(`${trimText(show.doorsTime) || "TBD"} | ${trimText(show.ageRestriction) || "TBD"} | ${trimText(show.price) || "TBD"}`)
    if (trimText(show.notes)) {
      lines.push(trimText(show.notes))
    }
    lines.push("")
  })
  return lines.join("\n").trim()
}

function cancellationNotice(show) {
  return {
    content: `⚠️ ${trimText(show.artist) || "A show"} on ${trimText(show.displayDate) || trimText(show.dateISO) || "an upcoming date"} has been cancelled.`,
  }
}

function showDetailPayload(show) {
  const fields = [
    { name: "ID", value: `#${trimText(show.id) || "(unknown)"}`, inline: true },
    { name: "Artist", value: trimText(show.artist) || "TBD", inline: true },
    { name: "Date", value: trimText(show.displayDate) || trimText(show.dateISO) || "TBD", inline: true },
    { name: "Doors", value: trimText(show.doorsTime) || "TBD", inline: true },
    { name: "Age", value: trimText(show.ageRestriction) || "TBD", inline: true },
    { name: "Price", value: trimText(show.price) || "TBD", inline: true },
    { name: "Status", value: trimText(show.status) || "confirmed", inline: true },
  ]
  if (trimText(show.discordMessageId)) {
    fields.push({ name: "Discord Message", value: `#${trimText(show.discordMessageId)}`, inline: false })
  }
  if (trimText(show.notes)) {
    fields.push({ name: "Notes", value: trimText(show.notes), inline: false })
  }
  return {
    content: "",
    embeds: [{
      title: `🎵 ${showLabel(show)}`,
      description: "Show record details",
      color: 0x57f287,
      fields,
    }],
  }
}

function pastShowsText(shows) {
  const list = Array.isArray(shows) ? shows : []
  if (list.length === 0) {
    return "🗂 Past Shows\n\nNo archived shows found."
  }
  const lines = ["🗂 Past Shows", ""]
  list.forEach((show) => {
    lines.push(`#${trimText(show.id) || "?"} — ${trimText(show.artist) || "Unknown artist"} — ${trimText(show.displayDate) || trimText(show.dateISO) || "TBD"}`)
    lines.push(`${trimText(show.status) || "archived"} | ${trimText(show.doorsTime) || "TBD"} | ${trimText(show.price) || "TBD"}`)
    if (trimText(show.notes)) {
      lines.push(trimText(show.notes))
    }
    lines.push("")
  })
  return lines.join("\n").trim()
}

function parseShowTitle(title) {
  const text = trimText(title)
  const match = text.match(/^🎵\s*(.*?)\s*[—-]\s*(.+)$/)
  if (!match) {
    return { artist: "", dateText: "" }
  }
  return { artist: trimText(match[1]), dateText: trimText(match[2]) }
}

module.exports = {
  showLabel,
  showAnnouncementPayload,
  upcomingShowsText,
  cancellationNotice,
  showDetailPayload,
  pastShowsText,
  parseShowTitle,
}
