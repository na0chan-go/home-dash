export type Note = {
  id: number
  kind: 'notice' | 'shopping'
  body: string
  pinned: boolean
  done: boolean
  created_at: string
  updated_at: string
}

export type DailyGarbage = {
  date: string
  weekday: string
  items: string[]
  label: string
}

export type DashboardResponse = {
  generatedAt: string
  notes: {
    notice: Note[]
    shopping: Note[]
  }
  garbage: {
    today: DailyGarbage
    tomorrow: DailyGarbage
  }
}

export async function fetchDashboard(signal?: AbortSignal): Promise<DashboardResponse> {
  const res = await fetch('/api/v1/dashboard', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
    signal
  })

  if (!res.ok) {
    throw new Error(`failed to fetch dashboard: ${res.status}`)
  }

  return (await res.json()) as DashboardResponse
}
