import { create } from 'zustand';
import { persist } from 'zustand/middleware';

const MAX_FAVORITES = 5;

interface SidebarPrefsState {
  favorites: string[];
  collapsedSections: Record<string, boolean>;
  toggleFavorite: (routeId: string) => void;
  hasFavorite: (routeId: string) => boolean;
  toggleSection: (sectionId: string) => void;
  isSectionCollapsed: (sectionId: string) => boolean;
}

export const useSidebarPrefsStore = create<SidebarPrefsState>()(
  persist(
    (set, get) => ({
      favorites: [],
      collapsedSections: {},
      toggleFavorite: (routeId) =>
        set((state) => {
          const exists = state.favorites.includes(routeId);
          if (exists) {
            return { favorites: state.favorites.filter((id) => id !== routeId) };
          }
          const next = [routeId, ...state.favorites.filter((id) => id !== routeId)].slice(0, MAX_FAVORITES);
          return { favorites: next };
        }),
      hasFavorite: (routeId) => get().favorites.includes(routeId),
      toggleSection: (sectionId) =>
        set((state) => ({
          collapsedSections: {
            ...state.collapsedSections,
            [sectionId]: !state.collapsedSections[sectionId],
          },
        })),
      isSectionCollapsed: (sectionId) => Boolean(get().collapsedSections[sectionId]),
    }),
    { name: 'neuralops-sidebar-prefs' },
  ),
);
