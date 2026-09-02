import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import {
  FIELD_NAME,
  isProtectedForm,
  handleSubmit,
  install,
  warnIfNoProtectedForm,
} from './his-embed'

function makeForm(inner: string): HTMLFormElement {
  document.body.innerHTML = `<form>${inner}</form>`
  return document.querySelector('form') as HTMLFormElement
}

function field(form: HTMLFormElement): HTMLInputElement | null {
  return form.querySelector(`input[name="${FIELD_NAME}"]`)
}

describe('his-embed', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('targets forms carrying an ALTCHA widget', () => {
    expect(isProtectedForm(makeForm('<altcha-widget></altcha-widget>'))).toBe(true)
  })

  it('targets forms carrying an ALTCHA payload field', () => {
    expect(isProtectedForm(makeForm('<input name="altcha" />'))).toBe(true)
  })

  it('leaves unrelated forms alone', () => {
    // A search box should not silently gain a hidden field or send signals.
    expect(isProtectedForm(makeForm('<input name="q" />'))).toBe(false)
  })

  it('writes the aggregates into a hidden field on submit', () => {
    const form = makeForm('<altcha-widget></altcha-widget>')
    handleSubmit({ target: form } as unknown as Event)

    const written = field(form)
    expect(written).not.toBeNull()
    expect(written!.type).toBe('hidden')

    const payload = JSON.parse(written!.value)
    // The keys are the contract with the Go scorer (his.Signals json tags).
    expect(Object.keys(payload).sort()).toEqual([
      'duration_ms',
      'key_interval_stdev_ms',
      'keydowns',
      'pointer_distance',
      'pointer_events',
      'scrolls',
      'time_to_first_ms',
      'touches',
    ])
  })

  it('reuses the same field across repeated submits', () => {
    const form = makeForm('<altcha-widget></altcha-widget>')
    handleSubmit({ target: form } as unknown as Event)
    handleSubmit({ target: form } as unknown as Event)

    expect(form.querySelectorAll(`input[name="${FIELD_NAME}"]`)).toHaveLength(1)
  })

  it('adds no field to a form without ALTCHA', () => {
    const form = makeForm('<input name="q" />')
    handleSubmit({ target: form } as unknown as Event)

    expect(field(form)).toBeNull()
  })

  it('ignores submit events from non-form targets', () => {
    expect(() => handleSubmit({ target: document.body } as unknown as Event)).not.toThrow()
  })

  describe('warnIfNoProtectedForm', () => {
    let warn: ReturnType<typeof vi.spyOn>

    beforeEach(() => {
      warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
    })

    afterEach(() => {
      warn.mockRestore()
    })

    it('stays quiet when a protected form is present', () => {
      makeForm('<altcha-widget></altcha-widget>')
      warnIfNoProtectedForm()

      expect(warn).not.toHaveBeenCalled()
    })

    it('names the widget-outside-the-form case, the likeliest mistake', () => {
      document.body.innerHTML = '<altcha-widget></altcha-widget><form><input name="q" /></form>'
      warnIfNoProtectedForm()

      expect(warn).toHaveBeenCalledTimes(1)
      const message = String(warn.mock.calls[0][0])
      expect(message).toContain('sits outside every <form>')
      expect(message).toContain('Move the widget inside the form')
    })

    it('reports a page with no widget at all, and points at the JS path', () => {
      document.body.innerHTML = '<form><input name="q" /></form>'
      warnIfNoProtectedForm()

      const message = String(warn.mock.calls[0][0])
      expect(message).toContain('no ALTCHA widget was found')
      expect(message).toContain('globalThis.gatechaHIS.signals()')
      // Telling someone to move a widget they do not have would send them
      // hunting for it.
      expect(message).not.toContain('Move the widget')
    })
  })

  it('exposes the collector on window for JS-driven integrations', () => {
    install()

    expect(typeof globalThis.gatechaHIS?.signals).toBe('function')
    expect(globalThis.gatechaHIS!.signals().duration_ms).toBeGreaterThanOrEqual(0)
    expect(() => globalThis.gatechaHIS!.stop()).not.toThrow()
  })
})
