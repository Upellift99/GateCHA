import { describe, it, expect, beforeEach } from 'vitest'
import { FIELD_NAME, isProtectedForm, handleSubmit, install } from './his-embed'

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

  it('exposes the collector on window for JS-driven integrations', () => {
    install()

    expect(typeof globalThis.gatechaHIS?.signals).toBe('function')
    expect(globalThis.gatechaHIS!.signals().duration_ms).toBeGreaterThanOrEqual(0)
    expect(() => globalThis.gatechaHIS!.stop()).not.toThrow()
  })
})
