import { createContext, useContext } from 'react';

export interface LayoutChromeContextValue {
  fullscreen: boolean;
  setFullscreen: (value: boolean) => void;
  onOpenNav: () => void;
}

// Default (no-op setter) so any component can call useLayoutChrome() safely
// even if it somehow renders outside MainLayout's provider — a page opting
// into fullscreen there would just silently do nothing instead of crashing.
export const LayoutChromeContext = createContext<LayoutChromeContextValue>({
    fullscreen: false,
    setFullscreen: () => {},
    onOpenNav: () => {},
});

export function useLayoutChrome() {
  return useContext(LayoutChromeContext);
}
