export type NoteKind = 'notice' | 'shopping'

export type Note = {
  id: number
  kind: NoteKind
  body: string
  author: string
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

type APIErrorBody = {
  error?: {
    code?: string
    message?: string
  }
  requestId?: string
  timestamp?: string
}

export class APIError extends Error {
  status: number
  code?: string
  requestId?: string
  timestamp?: string

  constructor(status: number, message: string, code?: string, requestId?: string, timestamp?: string) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
    this.requestId = requestId
    this.timestamp = timestamp
  }
}

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {})
    }
  })

  if (!res.ok) {
    let body: APIErrorBody | undefined
    try {
      body = (await res.json()) as APIErrorBody
    } catch {
      body = undefined
    }
    throw new APIError(
      res.status,
      body?.error?.message ?? `request failed: ${res.status}`,
      body?.error?.code,
      body?.requestId,
      body?.timestamp
    )
  }

  if (res.status === 204) {
    return undefined as T
  }

  return (await res.json()) as T
}

export async function fetchDashboard(signal?: AbortSignal): Promise<DashboardResponse> {
  return requestJSON<DashboardResponse>('/api/v1/dashboard', {
    method: 'GET',
    signal
  })
}

export async function createNote(kind: NoteKind, body: string, author = ''): Promise<Note> {
  return requestJSON<Note>('/api/v1/notes', {
    method: 'POST',
    body: JSON.stringify({ kind, body, author })
  })
}

export async function updateNotePin(id: number, pinned: boolean): Promise<Note> {
  return requestJSON<Note>(`/api/v1/notes/${id}/pin`, {
    method: 'PATCH',
    body: JSON.stringify({ pinned })
  })
}

export async function updateNoteDone(id: number, done: boolean): Promise<Note> {
  return requestJSON<Note>(`/api/v1/notes/${id}/done`, {
    method: 'PATCH',
    body: JSON.stringify({ done })
  })
}

export async function deleteNote(id: number): Promise<{ deleted: boolean }> {
  return requestJSON<{ deleted: boolean }>(`/api/v1/notes/${id}`, {
    method: 'DELETE'
  })
}
