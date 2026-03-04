import { describe, expect, it } from 'vitest'
import { createAsyncQueue } from '../../src/utils/asyncQueue'

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms)
  })
}

describe('createAsyncQueue', () => {
  it('タスクを順番に実行する', async () => {
    const enqueue = createAsyncQueue()
    const timeline: string[] = []

    const first = enqueue(async () => {
      timeline.push('first:start')
      await sleep(30)
      timeline.push('first:end')
      return 'first'
    })

    const second = enqueue(async () => {
      timeline.push('second:start')
      timeline.push('second:end')
      return 'second'
    })

    await expect(first).resolves.toBe('first')
    await expect(second).resolves.toBe('second')
    expect(timeline).toEqual(['first:start', 'first:end', 'second:start', 'second:end'])
  })

  it('途中で失敗しても後続タスクを実行する', async () => {
    const enqueue = createAsyncQueue()
    const timeline: string[] = []

    const failed = enqueue(async () => {
      timeline.push('failed:start')
      throw new Error('boom')
    })

    const after = enqueue(async () => {
      timeline.push('after:start')
      timeline.push('after:end')
      return 1
    })

    await expect(failed).rejects.toThrow('boom')
    await expect(after).resolves.toBe(1)
    expect(timeline).toEqual(['failed:start', 'after:start', 'after:end'])
  })
})
