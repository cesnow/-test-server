"use client";

import React, { useEffect } from "react";
import {createInstance, useMatomo} from "@cesnow/matomo-next";

export function AppProvider({ children }: { children: React.ReactNode }) {
  const { isInitialized, setInstance, trackEvent, pushTrackerUrl, addTracker } = useMatomo();

  useEffect(() => {
    if (isInitialized) {
      console.log("Matomo is ready:", window.Matomo);
      console.log(window.Matomo.initialized);
      const tracker = window.Matomo.getAsyncTracker();
      tracker.setTrackerUrl(
        "https://5fa9-59-115-174-54.ngrok-free.app/matomo.php2"
      );
      tracker.setSiteId(333);
      const newTracker = window.Matomo.getTracker(
        "https://5fa9-59-115-174-54.ngrok-free.app/matomo.php3", 1
      );
      addTracker("http://", 300);
      trackEvent({ category: "test", action: "test" });
    }
  }, [isInitialized]);

  useEffect(() => {
    const instance = createInstance({
      urlBase: "https://5fa9-59-115-174-54.ngrok-free.app",
      siteId: 3,
      userId: "UID76903202", // optional, default value: `undefined`.
      permanentTitle: "My Awesome App", // optional, always use this title for tracking, ignores document.title. Useful for SPAs.
      permanentHref: "/", // optional, always use this href for tracking, ignores window.location.href. Useful for SPAs.
      disabled: false, // optional, false by default. Makes all tracking calls no-ops if set to true.
      heartBeat: {
        // optional, enabled by default
        active: true, // optional, default value: true
        seconds: 10, // optional, default value: 15
      },
      linkTracking: false, // optional, default value: true
      configurations: {
        // optional, default value: {}
        // any valid matomo configuration, all below are optional
        disableCookies: true,
        setSecureCookie: true,
        setRequestMethod: "POST",
      },
    });
    setInstance(instance);
  }, []);

  return <>{children}</>;
}
