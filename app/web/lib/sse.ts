// 'use client';
//
// type Handler = (data: any) => void
//
// const eventSource = new EventSource('http://localhost:8801/sse?nt=tylyu')
//
// export function subscribe(eventName: string, handler: Handler) {
//   const listener = (e: MessageEvent) => {
//     try {
//       const payload = JSON.parse(e.data)
//       handler(payload)
//     } catch {
//       console.warn('cannot parse SSE payload:', e.data)
//     }
//   }
//   eventSource.addEventListener(eventName, listener)
//   return () => {
//     eventSource.removeEventListener(eventName, listener)
//   }
// }
//
// export function subscribeMessage(handler: Handler) {
//   const listener = (e: MessageEvent) => {
//     try {
//       const payload = JSON.parse(e.data)
//       handler(payload)
//     } catch {
//       console.warn('cannot parse SSE payload:', e.data)
//     }
//   }
//   eventSource.addEventListener('message', listener)
//   return () => {
//     eventSource.removeEventListener('message', listener)
//   }
// }
//
// let defaultListener: ((e: MessageEvent) => void) | null = null
//
// export function subscribeDefault(handler: Handler) {
//   defaultListener = (e: MessageEvent) => {
//     try {
//       const payload = JSON.parse(e.data)
//       handler(payload)
//     } catch {
//       console.warn('cannot parse SSE payload:', e.data)
//     }
//   }
//
//   eventSource.onmessage = defaultListener
//
//   return () => {
//     if (defaultListener === eventSource.onmessage) {
//       eventSource.onmessage = null
//     }
//     defaultListener = null
//   }
// }
