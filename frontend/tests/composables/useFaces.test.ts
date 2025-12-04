import type { components } from '@/types/types.generated'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../../src/api/client'
import { useFaces } from '../../src/composables/useFaces'

vi.mock('../../src/api/client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
  ApiError: class extends Error {},
}))

describe('useFaces', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('listFaces fetches faces', async () => {
    const { listFaces, faces } = useFaces()

    type FaceMinimal = components['schemas']['FaceMinimal']
    const mockFaces: FaceMinimal[] = [{ face_type: 'wa_40cm_full', face_name: 'WA 40cm Full' }]

    vi.mocked(api.get).mockResolvedValue(mockFaces)

    const result = await listFaces()

    expect(vi.mocked(api.get)).toHaveBeenCalledWith('/faces')
    expect(result).toEqual(mockFaces)
    expect(faces.value).toEqual(mockFaces)
  })

  it('fetchFace fetches single face', async () => {
    const { fetchFace, face } = useFaces()

    type Face = components['schemas']['Face']
    const mockFace: Face = {
      face_type: 'wa_40cm_full',
      rings: [],
      face_name: 'WA 40cm Full',
      render_cross: true,
      spots: [],
      viewBox: 100,
    }

    vi.mocked(api.get).mockResolvedValue(mockFace)

    const result = await fetchFace('wa_40cm_full')

    expect(vi.mocked(api.get)).toHaveBeenCalledWith('/faces/wa_40cm_full', expect.anything())
    expect(result).toEqual(mockFace)
    expect(face.value).toEqual(mockFace)
  })
})
