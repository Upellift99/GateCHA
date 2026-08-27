// Entry point for the standalone HIS collector script served at
// /api/public/his.js, for sites that integrate GateCHA without using this
// dashboard's build.
//
// It wraps the exact collector the admin login uses (./his), so there is one
// implementation of the signal maths rather than a copy that can drift from
// what the server scores.
//
// Behaviour once loaded:
//   - starts collecting interaction aggregates for the page
//   - on submit of a form that carries an ALTCHA widget or payload, writes
//     those aggregates as JSON into a hidden `gatecha_his_signals` field, for
//     the server to forward to POST /api/v1/verify
//   - exposes `globalThis.gatechaHIS` for integrations that call /verify from JS
//
// It never reads field values, pointer coordinates or key contents; see ./his.

import { startHISCollector, type HISCollector, type HISSignals } from './his'

/** Hidden field the script fills in, and the JSON key /verify expects. */
export const FIELD_NAME = 'gatecha_his_signals'

let collector: HISCollector | null = null

function ensureCollector(): HISCollector {
  collector ??= startHISCollector(document)
  return collector
}

/**
 * Only forms that actually reach /verify get a field. Adding a hidden input to
 * every form on the page (search boxes, newsletter widgets) would be surprising
 * and would send signals nobody asked for.
 */
export function isProtectedForm(form: HTMLFormElement): boolean {
  return form.querySelector('altcha-widget, [name="altcha"]') !== null
}

export function writeSignals(form: HTMLFormElement): void {
  let field = form.querySelector<HTMLInputElement>(`input[name="${FIELD_NAME}"]`)
  if (!field) {
    field = document.createElement('input')
    field.type = 'hidden'
    field.name = FIELD_NAME
    form.appendChild(field)
  }
  field.value = JSON.stringify(ensureCollector().signals())
}

export function handleSubmit(event: Event): void {
  const form = event.target
  if (form instanceof HTMLFormElement && isProtectedForm(form)) {
    writeSignals(form)
  }
}

export function install(): void {
  // Capture phase, so the field is present before a framework's own submit
  // handler serialises the form.
  document.addEventListener('submit', handleSubmit, true)
  ensureCollector()

  globalThis.gatechaHIS = {
    signals: () => ensureCollector().signals(),
    stop: () => collector?.stop(),
  }
}

declare global {
  // Declared on the global scope rather than on Window so `globalThis` sees it.
  // eslint-disable-next-line no-var
  var gatechaHIS:
    | {
        /** Aggregates collected so far, ready to send as `his_signals`. */
        signals(): HISSignals
        /** Detach listeners. */
        stop(): void
      }
    | undefined
}

install()
