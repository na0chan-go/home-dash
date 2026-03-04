import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import NoticeBoard from '../../src/components/NoticeBoard.vue'
import type { Note } from '../../src/api/client'

function flushPromises(): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, 0)
  })
}

function noticeNote(id: number, body: string): Note {
  const now = '2026-03-04T00:00:00Z'
  return {
    id,
    kind: 'notice',
    body,
    pinned: false,
    done: false,
    created_at: now,
    updated_at: now
  }
}

describe('NoticeBoard', () => {
  it('追加時に本文をtrimしてonAddを呼び、入力欄をクリアする', async () => {
    const onAdd = vi.fn().mockResolvedValue(undefined)
    const wrapper = mount(NoticeBoard, {
      props: {
        items: [],
        pendingIds: [],
        onAdd,
        onTogglePin: vi.fn().mockResolvedValue(undefined),
        onDeleteNote: vi.fn().mockResolvedValue(undefined)
      }
    })

    const input = wrapper.get('input[placeholder="連絡を入力"]')
    await input.setValue('  牛乳を買う  ')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(onAdd).toHaveBeenCalledTimes(1)
    expect(onAdd).toHaveBeenCalledWith('牛乳を買う')
    expect((input.element as HTMLInputElement).value).toBe('')
  })

  it('空本文ではバリデーションエラーを表示してonAddを呼ばない', async () => {
    const onAdd = vi.fn().mockResolvedValue(undefined)
    const wrapper = mount(NoticeBoard, {
      props: {
        items: [],
        pendingIds: [],
        onAdd,
        onTogglePin: vi.fn().mockResolvedValue(undefined),
        onDeleteNote: vi.fn().mockResolvedValue(undefined)
      }
    })

    await wrapper.get('input[placeholder="連絡を入力"]').setValue('   ')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(onAdd).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('本文を入力してください')
  })

  it('削除確認ダイアログでキャンセル/削除が正しく動く', async () => {
    const note = noticeNote(1, '連絡テスト')
    const onDeleteNote = vi.fn().mockResolvedValue(undefined)
    const wrapper = mount(NoticeBoard, {
      props: {
        items: [note],
        pendingIds: [],
        onAdd: vi.fn().mockResolvedValue(undefined),
        onTogglePin: vi.fn().mockResolvedValue(undefined),
        onDeleteNote
      }
    })

    const findListDeleteButton = () =>
      wrapper
        .findAll('button')
        .find((button) => button.text() === '削除' && button.attributes('type') === 'button')

    await findListDeleteButton()!.trigger('click')
    expect(wrapper.find('[aria-label="削除確認"]').exists()).toBe(true)

    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'キャンセル')!
      .trigger('click')
    expect(wrapper.find('[aria-label="削除確認"]').exists()).toBe(false)
    expect(onDeleteNote).not.toHaveBeenCalled()

    await findListDeleteButton()!.trigger('click')
    const deleteButtons = wrapper.findAll('button').filter((button) => button.text() === '削除')
    await deleteButtons[deleteButtons.length - 1]!.trigger('click')
    await flushPromises()

    expect(onDeleteNote).toHaveBeenCalledTimes(1)
    expect(onDeleteNote).toHaveBeenCalledWith(note)
  })
})
