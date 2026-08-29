import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import ConfirmDialog from './ConfirmDialog.vue'

function mountDialog(props: Record<string, unknown> = {}) {
  return mount(ConfirmDialog, {
    attachTo: document.body,
    props: { open: true, title: 'Revoke this token?', ...props },
    slots: { default: 'This cannot be undone.' },
  })
}

describe('ConfirmDialog', () => {
  it('renders nothing while closed', () => {
    const wrapper = mountDialog({ open: false })
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('renders the title, body and labels', () => {
    const wrapper = mountDialog({ confirmLabel: 'Revoke' })
    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.text()).toContain('Revoke this token?')
    expect(dialog.text()).toContain('This cannot be undone.')
    expect(dialog.text()).toContain('Revoke')
    expect(dialog.text()).toContain('Cancel')
  })

  // A dialog whose title is not wired to aria-labelledby is announced as an
  // unnamed group by screen readers.
  it('names itself from its own title element', () => {
    const wrapper = mountDialog()
    const labelledBy = wrapper.get('[role="dialog"]').attributes('aria-labelledby')
    expect(labelledBy).toBeTruthy()
    expect(wrapper.get(`#${CSS.escape(labelledBy!)}`).text()).toBe('Revoke this token?')
  })

  it('emits confirm and cancel from the two buttons', async () => {
    const wrapper = mountDialog({ confirmLabel: 'Revoke' })
    const buttons = wrapper.findAll('[role="dialog"] button')

    await buttons.find((b) => b.text() === 'Cancel')!.trigger('click')
    await buttons.find((b) => b.text() === 'Revoke')!.trigger('click')

    expect(wrapper.emitted('cancel')).toHaveLength(1)
    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })

  it('cancels on Escape and on a click outside the panel', async () => {
    const wrapper = mountDialog()

    await wrapper.get('[role="dialog"]').trigger('keydown', { key: 'Escape' })
    expect(wrapper.emitted('cancel')).toHaveLength(1)

    // The overlay is the dialog's parent; only a click on the overlay itself
    // counts, never one that bubbled up from inside the panel.
    await wrapper.get('div.fixed').trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(2)

    await wrapper.get('[role="dialog"]').trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(2)
  })

  // Opening a modal without moving focus into it leaves keyboard users stranded
  // behind the backdrop.
  it('moves focus to the cancel button when it opens', async () => {
    const wrapper = mountDialog({ open: false })
    await wrapper.setProps({ open: true })
    await nextTick()

    expect((document.activeElement as HTMLElement | null)?.textContent?.trim()).toBe('Cancel')
    wrapper.unmount()
  })

  it('moves focus in even when it is mounted already open', async () => {
    const wrapper = mountDialog()
    await nextTick()

    expect((document.activeElement as HTMLElement | null)?.textContent?.trim()).toBe('Cancel')
    wrapper.unmount()
  })

  it('disables the confirm button while busy', () => {
    const wrapper = mountDialog({ confirmLabel: 'Revoke', busy: true })
    const confirm = wrapper.findAll('[role="dialog"] button').find((b) => b.text() === 'Revoke')
    expect(confirm!.attributes('disabled')).toBeDefined()
  })
})
