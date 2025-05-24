// 'use client';
//
// import { useEffect } from 'react'
// import { subscribe, subscribeDefault } from "@/lib/sse"
//
// export function useSSE(
//   eventName: string,
//   handler: (data: any) => void
// ) {
//   useEffect(() => {
//     const unsubscribe = subscribe(eventName, handler)
//     return () => {
//       unsubscribe()
//     }
//   }, [eventName, handler])
// }
//
// export function useSSEDefault(handler: (data: any) => void) {
//   useEffect(() => {
//     const unsubscribe = subscribeDefault(handler)
//     return () => {
//       unsubscribe()
//     }
//   }, [handler])
// }
//
