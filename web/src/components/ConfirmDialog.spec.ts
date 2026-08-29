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

function dialogEl(wrapper: ReturnType<typeof mountDialog>) {
  return wrapper.get('dialog').element as HTMLDialogElement
}

describe('ConfirmDialog', () => {
  it('stays closed and renders no panel until asked to open', () => {
    const wrapper = mountDialog({ open: false })
    expect(dialogEl(wrapper).open).toBe(false)
    expect(wrapper.find('dialog button').exists()).toBe(false)
  })

  it('renders the title, body and labels', async () => {
    const wrapper = mountDialog({ confirmLabel: 'Revoke' })
    await nextTick()

    const dialog = wrapper.get('dialog')
    expect(dialog.text()).toContain('Revoke this token?')
    expect(dialog.text()).toContain('This cannot be undone.')
    expect(dialog.text()).toContain('Revoke')
    expect(dialog.text()).toContain('Cancel')
  })

  // showModal is what makes the rest of the page inert; a plain open attribute
  // would render the panel without any of that.
  it('opens and closes as a modal', async () => {
    const wrapper = mountDialog({ open: false })

    await wrapper.setProps({ open: true })
    await nextTick()
    expect(dialogEl(wrapper).open).toBe(true)

    await wrapper.setProps({ open: false })
    await nextTick()
    expect(dialogEl(wrapper).open).toBe(false)
  })

  // A dialog whose title is not wired to aria-labelledby is announced unnamed.
  it('names itself from its own title element', async () => {
    const wrapper = mountDialog()
    await nextTick()

    const labelledBy = wrapper.get('dialog').attributes('aria-labelledby')
    expect(labelledBy).toBeTruthy()
    expect(wrapper.get(`#${CSS.escape(labelledBy!)}`).text()).toBe('Revoke this token?')
  })

  it('emits confirm and cancel from the two buttons', async () => {
    const wrapper = mountDialog({ confirmLabel: 'Revoke' })
    await nextTick()
    const buttons = wrapper.findAll('dialog button')

    await buttons.find((b) => b.text() === 'Cancel')!.trigger('click')
    await buttons.find((b) => b.text() === 'Revoke')!.trigger('click')

    expect(wrapper.emitted('cancel')).toHaveLength(1)
    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })

  // Escape reaches us as the dialog's own cancel event. The default close is
  // suppressed so the caller's prop stays the one thing that opens and closes it.
  it('cancels on Escape without closing itself behind the caller', async () => {
    const wrapper = mountDialog()
    await nextTick()

    await wrapper.get('dialog').trigger('cancel')

    expect(wrapper.emitted('cancel')).toHaveLength(1)
    expect(dialogEl(wrapper).open).toBe(true)
  })

  it('cancels on a click outside the panel but not inside it', async () => {
    const wrapper = mountDialog()
    await nextTick()

    // Clicks on the backdrop are dispatched to the dialog element itself.
    await wrapper.get('dialog').trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(1)

    await wrapper.get('dialog > div').trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(1)
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

  it('disables the confirm button while busy', async () => {
    const wrapper = mountDialog({ confirmLabel: 'Revoke', busy: true })
    await nextTick()

    const confirm = wrapper.findAll('dialog button').find((b) => b.text() === 'Revoke')
    expect(confirm!.attributes('disabled')).toBeDefined()
  })
})
