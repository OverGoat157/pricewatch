const rubFmt = new Intl.NumberFormat('ru-RU', {
  style: 'currency',
  currency: 'RUB',
  maximumFractionDigits: 0,
})

export function formatRub(value) {
  if (value === null || value === undefined) return '—'
  return rubFmt.format(value)
}

const dateTimeFmt = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

export function formatDateTime(value) {
  if (!value) return '—'
  return dateTimeFmt.format(new Date(value))
}

const shortDateFmt = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
})

export function formatShortDate(value) {
  if (!value) return ''
  return shortDateFmt.format(new Date(value))
}
