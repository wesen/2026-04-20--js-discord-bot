const MONTHS = {
  jan: 1,
  january: 1,
  feb: 2,
  february: 2,
  mar: 3,
  march: 3,
  apr: 4,
  april: 4,
  may: 5,
  jun: 6,
  june: 6,
  jul: 7,
  july: 7,
  aug: 8,
  august: 8,
  sep: 9,
  sept: 9,
  september: 9,
  oct: 10,
  october: 10,
  nov: 11,
  november: 11,
  dec: 12,
  december: 12,
}

const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]
const MONTH_NAMES = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]

function trimText(value) {
  return String(value || "").trim()
}

function localDateISO(date) {
  if (!(date instanceof Date) || Number.isNaN(date.getTime())) {
    return ""
  }
  const year = String(date.getFullYear())
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  return `${year}-${month}-${day}`
}

function todayISO(referenceDate) {
  return localDateISO(referenceDate instanceof Date ? referenceDate : new Date())
}

function formatDisplayDate(date, timeZone) {
  void timeZone
  if (!(date instanceof Date) || Number.isNaN(date.getTime())) {
    return ""
  }
  const weekday = WEEKDAYS[date.getDay()] || ""
  const month = MONTH_NAMES[date.getMonth()] || ""
  const day = String(date.getDate())
  const year = String(date.getFullYear())
  if (!weekday || !month || !day || !year) {
    return ""
  }
  return `${weekday} ${month} ${day}, ${year}`
}

function parseShowDate(input, options) {
  const text = trimText(input)
  if (!text) {
    return { ok: false, error: "Date is required." }
  }

  const now = options && options.referenceDate instanceof Date ? options.referenceDate : new Date()
  const timeZone = options && options.timeZone
  let year
  let month
  let day
  let explicitYear = false

  const isoMatch = text.match(/^(\d{4})-(\d{2})-(\d{2})$/)
  if (isoMatch) {
    year = Number(isoMatch[1])
    month = Number(isoMatch[2])
    day = Number(isoMatch[3])
    explicitYear = true
  } else {
    const namedMatch = text.match(/^([A-Za-z]+)\s+(\d{1,2})(?:,\s*(\d{4}))?$/)
    if (!namedMatch) {
      return { ok: false, error: `Could not parse date ${JSON.stringify(text)}.` }
    }
    const monthName = namedMatch[1].toLowerCase()
    month = MONTHS[monthName]
    day = Number(namedMatch[2])
    if (!month) {
      return { ok: false, error: `Unknown month ${JSON.stringify(namedMatch[1])}.` }
    }
    if (namedMatch[3]) {
      year = Number(namedMatch[3])
      explicitYear = true
    } else {
      year = now.getFullYear()
    }
  }

  if (!Number.isInteger(year) || !Number.isInteger(month) || !Number.isInteger(day)) {
    return { ok: false, error: `Could not parse date ${JSON.stringify(text)}.` }
  }

  let date = new Date(year, month - 1, day)
  if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) {
    return { ok: false, error: `Invalid calendar date ${JSON.stringify(text)}.` }
  }

  if (!explicitYear) {
    const candidateISO = localDateISO(date)
    const nowISO = localDateISO(now)
    if (candidateISO && nowISO && candidateISO < nowISO) {
      date = new Date(year + 1, month - 1, day)
      if (date.getFullYear() !== year + 1 || date.getMonth() !== month - 1 || date.getDate() !== day) {
        return { ok: false, error: `Invalid calendar date ${JSON.stringify(text)}.` }
      }
    }
  }

  return {
    ok: true,
    dateISO: localDateISO(date),
    displayDate: formatDisplayDate(date, timeZone),
    year: date.getFullYear(),
    month: date.getMonth() + 1,
    day: date.getDate(),
  }
}

function compareDateISO(a, b) {
  return String(a || "").localeCompare(String(b || ""))
}

function isPastDate(dateISO, referenceDate) {
  const today = todayISO(referenceDate)
  return String(dateISO || "") < today
}

module.exports = {
  parseShowDate,
  formatDisplayDate,
  localDateISO,
  todayISO,
  compareDateISO,
  isPastDate,
}
