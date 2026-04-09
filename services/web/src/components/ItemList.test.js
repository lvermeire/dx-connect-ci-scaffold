import { mount, flushPromises } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ItemList from './ItemList.vue'

describe('ItemList', () => {
  beforeEach(() => {
    global.fetch = vi.fn()
  })

  it('renders items fetched on mount', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([{ id: '1', name: 'widget' }]),
    })

    const wrapper = mount(ItemList)
    await flushPromises()

    const items = wrapper.findAll('[data-testid="item"]')
    expect(items).toHaveLength(1)
    expect(items[0].text()).toBe('widget')
  })

  it('renders empty list when no items exist', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([]),
    })

    const wrapper = mount(ItemList)
    await flushPromises()

    expect(wrapper.findAll('[data-testid="item"]')).toHaveLength(0)
  })

  it('adds an item on form submit and appends to list', async () => {
    global.fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: '1', name: 'gadget' }),
      })

    const wrapper = mount(ItemList)
    await flushPromises()

    await wrapper.find('[data-testid="item-input"]').setValue('gadget')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    const items = wrapper.findAll('[data-testid="item"]')
    expect(items).toHaveLength(1)
    expect(items[0].text()).toBe('gadget')
    expect(wrapper.find('[data-testid="item-input"]').element.value).toBe('')
  })

  it('shows an error when fetch fails', async () => {
    global.fetch.mockRejectedValueOnce(new Error('network error'))

    const wrapper = mount(ItemList)
    await flushPromises()

    expect(wrapper.find('[data-testid="error"]').exists()).toBe(true)
  })
})
