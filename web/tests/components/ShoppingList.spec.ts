import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ShoppingList from '../../src/components/ShoppingList.vue'
import type { Note } from '../../src/api/client'

function flushPromises(): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, 0)
  })
}

function shoppingNote(id: number, body: string, done: boolean): Note {
  const now = '2026-03-04T00:00:00Z'
  return {
    id,
    kind: 'shopping',
    body,
    author: '',
    pinned: false,
    acknowledged: false,
    done,
    created_at: now,
    updated_at: now
  }
}

describe('ShoppingList', () => {
  it('追加時に本文をtrimしてonAddを呼び、入力欄をクリアする', async () => {
    const onAdd = vi.fn().mockResolvedValue(undefined)
    const wrapper = mount(ShoppingList, {
      props: {
        items: [],
        pendingIds: [],
        onAdd,
        onToggleDone: vi.fn().mockResolvedValue(undefined),
        onDeleteNote: vi.fn().mockResolvedValue(undefined)
      }
    })

    const input = wrapper.get('input[placeholder="買い物メモを入力"]')
    await input.setValue('  卵  ')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(onAdd).toHaveBeenCalledTimes(1)
    expect(onAdd).toHaveBeenCalledWith('卵')
    expect((input.element as HTMLInputElement).value).toBe('')
  })

  it('固定候補を押すとそのまま追加できる', async () => {
    const onAdd = vi.fn().mockResolvedValue(undefined)
    const wrapper = mount(ShoppingList, {
      props: {
        items: [],
        pendingIds: [],
        onAdd,
        onToggleDone: vi.fn().mockResolvedValue(undefined),
        onDeleteNote: vi.fn().mockResolvedValue(undefined)
      }
    })

    await wrapper
      .findAll('button')
      .find((button) => button.text() === '牛乳')!
      .trigger('click')
    await flushPromises()

    expect(onAdd).toHaveBeenCalledWith('牛乳')
  })

  it('完了切替ボタンでonToggleDoneを呼ぶ', async () => {
    const note = shoppingNote(1, '牛乳', false)
    const onToggleDone = vi.fn().mockResolvedValue(undefined)
    const wrapper = mount(ShoppingList, {
      props: {
        items: [note],
        pendingIds: [],
        onAdd: vi.fn().mockResolvedValue(undefined),
        onToggleDone,
        onDeleteNote: vi.fn().mockResolvedValue(undefined)
      }
    })

    await wrapper
      .findAll('button')
      .find((button) => button.text() === '完了')!
      .trigger('click')
    await flushPromises()

    expect(onToggleDone).toHaveBeenCalledTimes(1)
    expect(onToggleDone).toHaveBeenCalledWith(note)
  })

  it('デフォルトは未完了のみ表示し、チェックで完了も表示する', async () => {
    const wrapper = mount(ShoppingList, {
      props: {
        items: [shoppingNote(1, '牛乳', false), shoppingNote(2, '完了メモ', true)],
        pendingIds: [],
        onAdd: vi.fn().mockResolvedValue(undefined),
        onToggleDone: vi.fn().mockResolvedValue(undefined),
        onDeleteNote: vi.fn().mockResolvedValue(undefined)
      }
    })

    const listText = wrapper.find('ul.hd-list').text()
    expect(listText).toContain('牛乳')
    expect(listText).not.toContain('完了メモ')

    await wrapper.get('input[type="checkbox"]').setValue(true)

    expect(wrapper.find('ul.hd-list').text()).toContain('完了メモ')
  })

  it('最近追加した項目は重複なく表示する', () => {
    const wrapper = mount(ShoppingList, {
      props: {
        items: [
          shoppingNote(1, '牛乳', false),
          shoppingNote(2, '卵', false),
          shoppingNote(3, '牛乳', true)
        ],
        pendingIds: [],
        onAdd: vi.fn().mockResolvedValue(undefined),
        onToggleDone: vi.fn().mockResolvedValue(undefined),
        onDeleteNote: vi.fn().mockResolvedValue(undefined)
      }
    })

    const recentTitle = wrapper.text()
    expect(recentTitle).toContain('最近追加した項目')
    expect(wrapper.findAll('button').filter((button) => button.text() === '牛乳')).toHaveLength(2)
  })

  it('削除確認ダイアログでキャンセル/削除が正しく動く', async () => {
    const note = shoppingNote(1, '削除テスト', false)
    const onDeleteNote = vi.fn().mockResolvedValue(undefined)
    const wrapper = mount(ShoppingList, {
      props: {
        items: [note],
        pendingIds: [],
        onAdd: vi.fn().mockResolvedValue(undefined),
        onToggleDone: vi.fn().mockResolvedValue(undefined),
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
