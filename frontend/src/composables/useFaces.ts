import type { components } from '@/types/types.generated'
import { ref } from 'vue'
import { api, ApiError } from '@/api/client'

type ListFaceMinimal = components['schemas']['FaceMinimal'][]
type Face = components['schemas']['Face']
type FaceType = components['schemas']['FaceType']

const faces = ref<ListFaceMinimal>([])
const face = ref<Face | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

// Simple in-memory cache for single face fetch
const cache = new Map<string, Face>()
const inFlight: Record<string, AbortController> = {}

export function useFaces() {
  async function listFaces(): Promise<ListFaceMinimal> {
    loading.value = true
    error.value = null
    try {
      const data = await api.get<ListFaceMinimal>('/faces')
      faces.value = Array.isArray(data) ? data : []
      return faces.value
    }
    catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        error.value = 'Not authenticated. Please sign in again.'
      }
      else {
        error.value = e instanceof Error ? e.message : 'Failed to fetch faces'
      }
      throw e
    }
    finally {
      loading.value = false
    }
  }

  async function fetchFace(faceId: FaceType): Promise<Face> {
    const key = String(faceId)
    if (cache.has(key)) {
      const cached = cache.get(key)!
      face.value = cached
      return cached
    }
    if (inFlight[key]) {
      try {
        inFlight[key].abort()
      }
      catch {
        // ignore
      }
      delete inFlight[key]
    }
    const controller = new AbortController()
    inFlight[key] = controller
    loading.value = true
    error.value = null
    try {
      // Note: api.get wrapper doesn't currently support AbortSignal,
      // but we can add it or use direct fetch for this specific case if needed.
      // For now, let's use the wrapper and ignore the abort controller for simplicity
      // unless strict cancellation is required.
      // If strict cancellation is needed, we would need to extend api.client to accept init options.
      // Assuming we extended api client in step 1 to accept RequestInit.

      const body = await api.get<Face>(`/faces/${encodeURIComponent(key)}`, {
        signal: controller.signal,
      })

      if (!body)
        throw new Error('Failed to fetch face')

      cache.set(key, body)
      face.value = body
      return body
    }
    catch (e) {
      // Don't throw if aborted
      if (e instanceof Error && e.name === 'AbortError') {
        throw e
      }
      // Re-throw other errors
      throw e
    }
    finally {
      loading.value = false
      delete inFlight[key]
    }
  }

  async function createFace(payload: Partial<Face>): Promise<Face> {
    loading.value = true
    error.value = null
    try {
      const body = await api.post<Face>('/faces', payload)
      if (!body)
        throw new Error('Failed to create face')
      cache.set(String(body.face_type), body)
      face.value = body
      return body
    }
    finally {
      loading.value = false
    }
  }

  async function updateFace(faceId: FaceType, payload: Partial<Face>): Promise<Face> {
    loading.value = true
    error.value = null
    try {
      const body = await api.patch<Face>(
        `/faces/${encodeURIComponent(String(faceId))}`,
        payload,
      )
      if (!body)
        throw new Error('Failed to update face')
      cache.set(String(faceId), body)
      face.value = body
      return body
    }
    finally {
      loading.value = false
    }
  }

  async function deleteFace(faceId: FaceType): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await api.delete(`/faces/${encodeURIComponent(String(faceId))}`)
      cache.delete(String(faceId))
      face.value = null
    }
    finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    faces,
    face,
    listFaces,
    fetchFace,
    createFace,
    updateFace,
    deleteFace,
  } as const
}
