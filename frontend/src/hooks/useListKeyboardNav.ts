import { useCallback, useEffect, useState } from 'react';

/** j/k list navigation with Enter to select and / to focus search. */
export function useListKeyboardNav<T>(
  items: T[],
  options: {
    onSelect: (item: T, index: number) => void;
    searchInputRef?: React.RefObject<HTMLInputElement | null>;
    enabled?: boolean;
  },
) {
  const { onSelect, searchInputRef, enabled = true } = options;
  const [activeIndex, setActiveIndex] = useState(0);

  useEffect(() => {
    setActiveIndex(0);
  }, [items.length]);

  const onKeyDown = useCallback(
    (event: KeyboardEvent) => {
      if (!enabled || items.length === 0) return;
      if (event.key === '/' && searchInputRef?.current) {
        event.preventDefault();
        searchInputRef.current.focus();
        return;
      }
      if (event.key === 'j') {
        event.preventDefault();
        setActiveIndex((i) => Math.min(items.length - 1, i + 1));
      }
      if (event.key === 'k') {
        event.preventDefault();
        setActiveIndex((i) => Math.max(0, i - 1));
      }
      if (event.key === 'Enter') {
        const item = items[activeIndex];
        if (item) onSelect(item, activeIndex);
      }
    },
    [activeIndex, enabled, items, onSelect, searchInputRef],
  );

  useEffect(() => {
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [onKeyDown]);

  return { activeIndex, setActiveIndex };
}
